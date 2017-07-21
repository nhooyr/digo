package discgo

import (
	"encoding/json"
	"time"

	"sync"

	"runtime"

	"compress/zlib"
	"errors"

	"strings"

	"io"

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

type Conn struct {
	// TODO use apiClient because it will need to be in the context for eventHandlers
	token      string
	userAgent  string
	gatewayURL string

	sessionID string

	closeChan chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup

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
	return &Conn{
		token:         apiClient.Token,
		userAgent:     apiClient.UserAgent,
		gatewayURL:    gatewayURL,
		closeChan:     make(chan struct{}),
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
	c.closeOnce.Do(func() {
		close(c.closeChan)
		c.wg.Wait()
	})
	return nil
}

type dataOPHello struct {
	HeartbeatInterval int      `json:"heartbeat_interval"`
	Trace             []string `json:"_trace"`
}

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

	c.runWorker(c.readLoop)

	return nil
}

type dataOPResume struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	Seq       int    `json:"seq"`
}

func (c *Conn) resume() error {
	c.mu.Lock()
	p := &payloadSend{
		Operation: operationResume,
		Data: dataOPResume{
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

func (c *Conn) readLoop() {
	for {
		p, err := c.readPayload()
		if err != nil {
			errStr := err.Error()
			if !strings.Contains(errStr, "use of closed network connection") {
				log.Print(err)
				// It's possible we're being shutdown right now too.
				// Or maybe manager is already trying to reconnect.
				select {
				case c.reconnectChan <- struct{}{}:
				case <-c.closeChan:
				}
			}
			return
		}
		err = c.onPayload(p)
		if err != nil {
			log.Print(err)
			// It's possible we're being shutdown right now.
			// Or maybe manager is already trying to reconnect.
			select {
			case c.reconnectChan <- struct{}{}:
			case <-c.closeChan:
			}
		}
	}
}

type payloadSend struct {
	Operation int         `json:"op"`
	Data      interface{} `json:"d,omitempty"`
	Sequence  int         `json:"s,omitempty"`
}

func (c *Conn) onPayload(p *payloadReceive) error {
	switch p.Operation {
	case operationHello:
		var hello dataOPHello
		err := json.Unmarshal(p.Data, &hello)
		if err != nil {
			return err
		}
		go c.manager(hello.HeartbeatInterval)
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
	log.Print(p.Operation)
	if p.Type != "" {
		log.Print(p.Type)
	}
	log.Print(p.SequenceNumber)
	log.Printf("%s", p.Data)
	log.Print()
	return nil
}

func (c *Conn) manager(heartbeatInterval int) {
	ticker := time.NewTicker(time.Duration(heartbeatInterval) * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := c.heartbeat()
			if err != nil {
				log.Print(err)
				err := c.close()
				if err != nil {
					log.Print(err)
				}
				<-c.reconnectChan

				err = c.Dial()
				if err != nil {
					log.Print(err)
				}
				return
			}
		case <-c.reconnectChan:
			err := c.close()
			if err != nil {
				log.Print(err)
			}

			err = c.Dial()
			if err != nil {
				log.Print(err)
			}
			return
		case <-c.closeChan:
			err := c.close()
			if err != nil {
				log.Print(err)
			}
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

type dataOPIdentify struct {
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
		Data: dataOPIdentify{
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
