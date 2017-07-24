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

	"math/rand"

	"github.com/gorilla/websocket"
	"github.com/nhooyr/log"
)

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
	Client   *Client
	EventMux eventMux

	internalEventMux eventMux
	gatewayURL       string

	sessionID string

	closeChan chan struct{}
	wg        sync.WaitGroup

	lastIdentify time.Time
	ready    bool
	resuming bool

	reconnectChan chan struct{}

	wsConn    *websocket.Conn
	writeChan chan *sentPayload

	heartbeatMu           sync.Mutex
	heartbeatAcknowledged bool
	sequenceNumber        int
}

func NewConn() *Conn {
	internalEventMux := newEventMux()
	internalEventMux.Register(func(ctx context.Context, conn *Conn, e *eventReady) {
		conn.sessionID = e.SessionID
	})
	return &Conn{
		internalEventMux: internalEventMux,
		EventMux:         newEventMux(),
		closeChan:        make(chan struct{}),
		reconnectChan:    make(chan struct{}),
		ready:            true,
	}
}

// TODO maybe take context.Context? though I doubt it's necessary
// If it errors without receiving a operationDispatch payload, then
// this method will always return an error.
func (c *Conn) Dial() (err error) {
	if !c.ready {
		return errors.New("already tried to connect and failed")
	}

	c.ready = false
	c.heartbeatAcknowledged = true
	c.lastIdentify = time.Time{}

	if c.gatewayURL == "" {
		c.gatewayURL, err = c.Client.gateway().get()
		if err != nil {
			return err
		}
		c.gatewayURL += "?v=" + apiVersion + "&encoding=json"
	}

	// TODO Need to set read deadline for hello packet and I also need to set write deadlines.
	// TODO also max message
	c.wsConn, _, err = websocket.DefaultDialer.Dial(c.gatewayURL, nil)
	if err != nil {
		return err
	}

	go c.manager()

	return nil
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

type sentPayload struct {
	Operation int         `json:"op"`
	Data      interface{} `json:"d,omitempty"`
	Sequence  int         `json:"s,omitempty"`
}

// TODO What to do with error writes? Especially with Session Invalid Op?
// TODO should I use some sort of queue to guarantee the event gets to discord in order? Not sure.
func (c *Conn) writeLoop(ctx context.Context) {
	t := time.NewTicker(time.Minute)
	defer t.Stop()

	writesLeft := 120
	var err error

writeLoop:
	for {
		select {
		case v := <-c.writeChan:
			if writesLeft == 0 {
				select {
				case <-t.C:
					writesLeft = 120
				case <-ctx.Done():
					break writeLoop
				}
			}

			writesLeft--

			_, ok := v.Data.(*dataOpIdentify)
			if ok {
				now := time.Now()
				resetTime := c.lastIdentify.Add(5 * time.Second)
				time.Sleep(resetTime.Sub(now))
			}

			err := c.wsConn.WriteJSON(v)
			if err != nil {
				log.Print(err)
				select {
				case c.reconnectChan <- struct{}{}:
				case <-ctx.Done():
				}
				break writeLoop
			}

			if ok {
				c.lastIdentify = time.Now()
			}
		case <-t.C:
			writesLeft = 120
		case <-ctx.Done():
			break writeLoop
		}
	}

	closeMsg := websocket.FormatCloseMessage(websocket.CloseNoStatusReceived, "no heartbeat acknowledgment")
	err = c.wsConn.WriteMessage(websocket.CloseMessage, closeMsg)
	if err != nil {
		log.Print(err)
	}

	err = c.wsConn.Close()
	if err != nil {
		log.Print(err)
	}
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

func (c *Conn) write(ctx context.Context, p *sentPayload) {
	select {
	case c.writeChan <- p:
	case <-ctx.Done():
	}
}

func (c *Conn) identify(ctx context.Context) {
	c.resuming = false

	p := &sentPayload{
		Operation: operationIdentify,
		Data: dataOpIdentify{
			Token: c.Client.Token,
			Properties: identifyProperties{
				OS:      runtime.GOOS,
				Browser: c.Client.UserAgent,
			},
			Compress:       true,
			LargeThreshold: 250,
		},
	}

	c.write(ctx, p)
}

type dataOpResume struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	Seq       int    `json:"seq"`
}

func (c *Conn) resume(ctx context.Context) {
	c.resuming = true

	c.heartbeatMu.Lock()
	p := &sentPayload{
		Operation: operationResume,
		Data: dataOpResume{
			Token:     c.Client.Token,
			SessionID: c.sessionID,
			Seq:       c.sequenceNumber,
		},
	}
	c.heartbeatMu.Unlock()
	c.write(ctx, p)
}

func (c *Conn) manager() {
	ctx, cancelFn := context.WithCancel(context.Background())
	c.runWorker(func() {
		c.writeLoop(ctx)
	})
	c.runWorker(func() {
		c.readLoop(ctx)
	})

	if c.sessionID == "" {
		c.identify(ctx)
	} else {
		c.resume(ctx)
	}

	select {
	case <-c.reconnectChan:
		cancelFn()
		c.wg.Wait()

		err := c.Dial()
		if err != nil {
			log.Print(err)
		}
	case <-c.closeChan:
		cancelFn()
		c.wg.Wait()

		c.closeChan <- struct{}{}
	}
}

func (c *Conn) runWorker(fn func()) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		fn()
	}()
}

type receivedPayload struct {
	Operation      int             `json:"op"`
	Data           json.RawMessage `json:"d"`
	SequenceNumber int             `json:"s"`
	Type           string          `json:"t"`
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

func (c *Conn) readPayload() (*receivedPayload, error) {
	var p receivedPayload
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

type dataOpHello struct {
	HeartbeatInterval int      `json:"heartbeat_interval"`
	Trace             []string `json:"_trace"`
}

func (c *Conn) onPayload(ctx context.Context, p *receivedPayload) error {
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
		c.heartbeatMu.Lock()
		c.heartbeatAcknowledged = true
		c.heartbeatMu.Unlock()
	case operationInvalidSession:
		var resumable bool
		err := json.Unmarshal(p.Data, &resumable)
		if err != nil {
			return err
		}

		if resumable {
			c.resume(ctx)
		} else {
			// We reconnected and were trying to resume but we were too late.
			if !c.ready && c.resuming {
				// Sleep for a random amount of time between 1 and 5 seconds.
				randDur := time.Duration(rand.Int63n(4*int64(time.Second))) + 1
				time.Sleep(randDur)
			}
			c.identify(ctx)
		}

	case operationReconnect:
		return errors.New("reconnect operation")
	case operationDispatch:
		c.ready = true
		c.resuming = false

		c.heartbeatMu.Lock()
		c.sequenceNumber = p.SequenceNumber
		c.heartbeatMu.Unlock()

		err := c.internalEventMux.route(ctx, c, p, true)
		if err != nil {
			return err
		}

		err = c.EventMux.route(ctx, c, p, false)
		if err != nil {
			return err
		}
	default:
		panic("discord gone crazy; unexpected operation type")
	}
	return nil
}

func (c *Conn) heartbeatLoop(ctx context.Context, heartbeatInterval int) {
	t := time.NewTicker(time.Duration(heartbeatInterval) * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			err := c.heartbeat()
			if err != nil {
				log.Print(err)
				// Either we signal a reconnect or we have been signaled to close.
				select {
				case c.reconnectChan <- struct{}{}:
				case <-ctx.Done():
				}
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Conn) heartbeat() error {
	c.heartbeatMu.Lock()
	if !c.heartbeatAcknowledged {
		c.heartbeatMu.Unlock()
		return errors.New("heartbeat not acknowledged")
	}
	sequenceNumber := c.sequenceNumber
	c.heartbeatAcknowledged = false
	c.heartbeatMu.Unlock()

	p := &sentPayload{Operation: operationHeartbeat, Data: sequenceNumber}
	return c.wsConn.WriteJSON(p)
}

func (c *Conn) Close() error {
	c.closeChan <- struct{}{}
	<-c.closeChan
	return nil
}
