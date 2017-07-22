package discgo

import (
	"compress/zlib"
	"encoding/json"
	"errors"
	"io"
	"net"
	"runtime"
	"sync"
	"time"

	"context"

	"github.com/gorilla/websocket"
	"github.com/nhooyr/log"
)

type Game struct {
	Name string  `json:"name"`
	Type *int    `json:"type"`
	URL  *string `json:"url"`
}

type endpointGateway struct {
	*endpoint
}

func (c *Client) gateway() endpointGateway {
	e2 := c.e.appendMajor("gateway")
	return endpointGateway{e2}
}

func (g endpointGateway) get() (url string, err error) {
	var urlStruct struct {
		URL string `json:"url"`
	}
	return urlStruct.URL, g.doMethod("GET", nil, &urlStruct)
}

// TODO need separate heartbeat goroutine so that i can always close. e.g. before hello received
type Conn struct {
	// TODO use apiClient because it will need to be in the context for eventHandlers
	token      string
	userAgent  string
	gatewayURL string

	sessionID string

	ctx        context.Context
	cancelFn   func()
	wg         sync.WaitGroup
	waitClosed chan struct{}

	reconnectChan chan struct{}

	wsConn *websocket.Conn

	mu                    sync.Mutex
	heartbeatAcknowledged bool
	sequenceNumber        int
}

func NewConn(apiClient *Client) (*Conn, error) {
	gatewayURL, err := apiClient.gateway().get()
	if err != nil {
		return nil, err
	}
	gatewayURL += "?v=" + apiVersion + "&encoding=json"
	ctx, cancelFn := context.WithCancel(context.Background())
	return &Conn{
		token:         apiClient.Token,
		userAgent:     apiClient.UserAgent,
		gatewayURL:    gatewayURL,
		ctx:           ctx,
		cancelFn:      cancelFn,
		waitClosed:    make(chan struct{}),
		reconnectChan: make(chan struct{}),
	}, nil
}

const (
	operationDispatch            = iota
	operationHeartbeat
	operationIdentify
	operationStatusUpdate
	operationVoiceStateUpdate
	operationVoiceServerPing
	operationResume
	operationReconnect
	operationRequestGuildMembers
	operationInvalidSession
	operationHello
	operationHeartbeatACK
)

func (c *Conn) close() error {
	// TODO I think this should be OK, but I'm not sure. Should there be a writerLoop routine to sync writes?
	closeMsg := websocket.FormatCloseMessage(websocket.CloseNoStatusReceived, "no heartbeat acknowledgment")
	err := c.wsConn.WriteMessage(websocket.CloseMessage, closeMsg)
	err2 := c.wsConn.Close()
	if err != nil {
		return err
	}
	return err2
}

func (c *Conn) Close() error {
	c.cancelFn()
	<-c.waitClosed
	return nil
}

type dataOpHello struct {
	HeartbeatInterval int      `json:"heartbeat_interval"`
	Trace             []string `json:"_trace"`
}

// TODO maybe take context.Context? though I doubt it's necessary
func (c *Conn) Dial() (err error) {
	c.heartbeatAcknowledged = true

	// TODO Need to set read deadline for hello packet and I also need to set write deadlines.
	// TODO also max message
	c.wsConn, _, err = websocket.DefaultDialer.Dial(c.gatewayURL, nil)
	if err != nil {
		return err
	}

	if c.sessionID == "" {
		err = c.identify()
	} else {
		err = c.resume()
	}
	if err != nil {
		return err
	}

	go c.manager()

	return nil
}

type dataOpResume struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	Seq       int    `json:"seq"`
}

func (c *Conn) resume() error {
	c.mu.Lock()
	p := &payloadSend{
		Operation: operationResume,
		Data: dataOpResume{
			Token:     c.token,
			SessionID: c.sessionID,
			Seq:       c.sequenceNumber,
		},
	}
	c.mu.Unlock()
	return c.wsConn.WriteJSON(p)
}

type eventReady struct {
	V               int        `json:"v"`
	User            *User      `json:"user"`
	PrivateChannels []*Channel `json:"private_channels"`
	SessionID       string     `json:"session_id"`
	Trace           []string   `json:"_trace"`
}

type payloadReceive struct {
	Operation      int             `json:"op"`
	Data           json.RawMessage `json:"d"`
	SequenceNumber int             `json:"s"`
	Type           string          `json:"t"`
}

func (c *Conn) readPayload() (*payloadReceive, error) {
	var p payloadReceive
	msgType, r, err := c.wsConn.NextReader()
	if err != nil {
		return nil, err
	}
	switch msgType {
	case websocket.BinaryMessage:
		var z io.ReadCloser
		z, err = zlib.NewReader(r)
		if err != nil {
			return nil, err
		}
		defer z.Close()
		return &p, json.NewDecoder(z).Decode(&p)
	case websocket.TextMessage:
		return &p, json.NewDecoder(r).Decode(&p)
	default:
		return nil, errors.New("unexpected websocket message type")
	}
	return &p, err
}

func (c *Conn) runWorker(fn func()) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		fn()
	}()
}

func (c *Conn) readLoop(ctx context.Context) {
	for {
		p, err := c.readPayload()
		if err != nil {
			if !isUseOfClosedError(err) {
				log.Print(err)
				// It's possible we're being shutdown right now too.
				// Or maybe manager is already trying to reconnect.
				select {
				case c.reconnectChan <- struct{}{}:
				case <-ctx.Done():
				}
			}
			return
		}

		log.Print(p.Operation)
		if p.Type != "" {
			log.Print(p.Type)
		}
		log.Print(p.SequenceNumber)
		log.Printf("%s", p.Data)
		log.Print()

		err = c.onPayload(ctx, p)
		if err != nil {
			log.Print(err)
			// It's possible we're being shutdown right now.
			// Or maybe manager is already trying to reconnect.
			select {
			case c.reconnectChan <- struct{}{}:
			case <-ctx.Done():
			}
		}
	}
}

// TODO not sure if this is how I should do it
func isUseOfClosedError(err error) bool {
	opErr, ok := err.(*net.OpError)
	if !ok {
		return false
	}
	if opErr.Err.Error() == "use of closed network connection" {
		return true
	}
	return false
}

type payloadSend struct {
	Operation int         `json:"op"`
	Data      interface{} `json:"d,omitempty"`
	Sequence  int         `json:"s,omitempty"`
}

func (c *Conn) onPayload(ctx context.Context, p *payloadReceive) error {
	switch p.Operation {
	case operationHello:
		var hello dataOpHello
		err := json.Unmarshal(p.Data, &hello)
		if err != nil {
			return err
		}
		c.runWorker(func() {
			c.heartbeatLoop(ctx, hello.HeartbeatInterval)
		})
	case operationHeartbeatACK:
		c.mu.Lock()
		c.heartbeatAcknowledged = true
		c.mu.Unlock()
	case operationInvalidSession:
		// Wait out the possible rate limit.
		// TODO Need to have a max limit on this, only one time imo.
		time.Sleep(time.Second * 5)
		err := c.identify()
		if err != nil {
			return err
		}
	case operationReconnect:
		return errors.New("reconnect operation")
	case operationDispatch:
		c.mu.Lock()
		c.sequenceNumber = p.SequenceNumber
		c.mu.Unlock()

		// TODO onEvent stuff
		// TODO close connection if anything goes wrong.
		switch p.Type {
		case "READY":
			var ready eventReady
			err := json.Unmarshal(p.Data, &ready)
			if err != nil {
				return err
			}
			c.sessionID = ready.SessionID
		case "RESUMED":
		default:
			log.Print("unknown dispatch payload type")
		}
	default:
		panic("discord gone crazy; unexpected operation type")
	}
	return nil
}

func (c *Conn) manager() {
	ctx, cancelFn := context.WithCancel(c.ctx)
	c.runWorker(func() {
		c.readLoop(ctx)
	})
	select {
	case <-c.reconnectChan:
		cancelFn()
		err := c.close()
		if err != nil {
			log.Print(err)
		}
		c.wg.Wait()

		err = c.Dial()
		if err != nil {
			log.Print(err)
		}
	case <-ctx.Done():
		err := c.close()
		if err != nil {
			log.Print(err)
		}
		c.wg.Wait()
		c.waitClosed <- struct{}{}
	}
}

func (c *Conn) heartbeatLoop(ctx context.Context, heartbeatInterval int) {
	ticker := time.NewTicker(time.Duration(heartbeatInterval) * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := c.heartbeat()
			if err != nil {
				log.Print(err)
				// Either we signal a reconnect or we have been signalled to close.
				select {
				case c.reconnectChan <- struct{}{}:
				case <-ctx.Done():
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Conn) heartbeat() error {
	c.mu.Lock()
	if !c.heartbeatAcknowledged {
		c.mu.Unlock()
		return errors.New("heartbeat not acknowledged")
	}
	sequenceNumber := c.sequenceNumber
	c.heartbeatAcknowledged = false
	c.mu.Unlock()

	p := &payloadSend{Operation: operationHeartbeat, Data: sequenceNumber}
	return c.wsConn.WriteJSON(p)
}

type dataOpIdentify struct {
	Token          string             `json:"token"`
	Properties     identifyProperties `json:"properties"`
	Compress       bool               `json:"compress"`
	LargeThreshold int                `json:"large_threshold"`
}

type identifyProperties struct {
	OS      string `json:"$os,omitempty"`
	Browser string `json:"$browser,omitempty"`
	Device  string `json:"$device,omitempty"`
}

func (c *Conn) identify() error {
	p := &payloadSend{
		Operation: operationIdentify,
		Data: dataOpIdentify{
			Token: c.token,
			Properties: identifyProperties{
				OS:      runtime.GOOS,
				Browser: c.userAgent,
			},
			Compress:       true,
			LargeThreshold: 250,
		},
	}
	return c.wsConn.WriteJSON(p)
}
