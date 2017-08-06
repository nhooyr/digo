package discgo

import (
	"compress/zlib"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/rand"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type EndpointGateway struct {
	*endpoint
}

func (c *Client) Gateway() EndpointGateway {
	e2 := c.e.appendMajor("gateway")
	return EndpointGateway{e2}
}

func (e EndpointGateway) Get() (url string, err error) {
	var urlStruct struct {
		URL string `json:"url"`
	}
	return urlStruct.URL, e.doMethod(context.Background(), "GET", nil, &urlStruct)
}

func (e EndpointGateway) Bot() EndpointGatewayBot {
	e2 := e.appendMajor("bot")
	return EndpointGatewayBot{e2}
}

type EndpointGatewayBot struct {
	*endpoint
}

// TODO kinda awkward
type ModelGatewayBotGetResponse struct {
	URL    string `json:"url"`
	Shards int    `json:"shards"`
}

func (e EndpointGatewayBot) Get(ctx context.Context) (resp *ModelGatewayBotGetResponse, err error) {
	return resp, e.doMethod(ctx, "GET", nil, &resp)
}

type EventHandler interface {
	Handle(ctx context.Context, e interface{}) error
}

type Conn struct {
	token        string
	gatewayURL   string
	eventHandler EventHandler
	errorHandler func(err error)
	logf         func(format string, v ...interface{})

	sessionID string

	shard        *[2]int
	lastIdentify time.Time
	ready        bool
	resuming     bool

	reconnectChan chan struct{}
	closeChan     chan struct{}
	wg            sync.WaitGroup

	// TODO use other websocket package, it's better for my usecase.
	wsConn    *websocket.Conn
	writeChan chan *sentPayload

	heartbeatMu           sync.Mutex
	heartbeatAcknowledged bool
	sequenceNumber        int
}

func NewConn(config *DialConfig) *Conn {
	c := &Conn{
		token:        config.Token,
		gatewayURL:   config.GatewayURL + "?v=" + apiVersion + "&encoding=json",
		eventHandler: config.EventHandler,
		errorHandler: config.ErrorHandler,
		logf:         config.Logf,

		closeChan:     make(chan struct{}),
		reconnectChan: make(chan struct{}),
		writeChan:     make(chan *sentPayload),
		ready:         true,
	}
	if config.ShardsCount > 1 {
		c.shard = &[2]int{config.Shard, config.ShardsCount}
	}
	return c
}

type EventHandlerFunc func(ctx context.Context, e interface{}) error

func (h EventHandlerFunc) Handle(ctx context.Context, e interface{}) error {
	return h(ctx, e)
}

type DialConfig struct {
	GatewayURL   string
	Token        string
	EventHandler EventHandler

	// These may be called concurrently.
	ErrorHandler func(err error)
	Logf         func(format string, v ...interface{})

	Shard       int
	ShardsCount int
}

// NewDialConfig returns a new *DialConfig with sane defaults.
// TODO I don't like this because it's not obvious that GatewayURL and Token are required. Hmm.
func NewDialConfig() *DialConfig {
	return &DialConfig{
		ErrorHandler: func(err error) {
			log.Print(err)
		},
		Logf: func(format string, v ...interface{}) {
			log.Printf(format, v...)
		},
	}
}

func Dial(c *Conn, config *DialConfig) error {

	return c.dial()
}

// If it errors without receiving a operationDispatch payload, then
// this method will always return an error.
func (c *Conn) dial() (err error) {
	if !c.ready {
		return errors.New("already tried to connect and failed")
	}

	c.ready = false
	c.resuming = false
	c.heartbeatAcknowledged = true
	c.lastIdentify = time.Time{}

	// TODO Need to set read deadline for hello packet and I also need to set write deadlines.
	// TODO also max message
	c.wsConn, _, err = websocket.DefaultDialer.Dial(c.gatewayURL, nil)
	if err != nil {
		return err
	}

	go c.manager()

	return nil
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
		c.logf("restarting")

		cancelFn()
		c.wg.Wait()

		err := c.dial()
		if err != nil {
			c.log(err)
		}
	case <-c.closeChan:
		c.logf("exiting")

		cancelFn()
		c.wg.Wait()

		c.closeChan <- struct{}{}
	}
}

const (
	operationDispatch = iota
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

func (c *Conn) writeLoop(ctx context.Context) {
	t := time.NewTicker(time.Minute)
	defer t.Stop()

	writesLeft := 120
	var err error

writeLoop:
	for {
		select {
		case p := <-c.writeChan:
			if writesLeft == 0 {
				select {
				case <-t.C:
					writesLeft = 120
				case <-ctx.Done():
					break writeLoop
				}
			}

			writesLeft--

			_, ok := p.Data.(*dataOpIdentify)
			if ok {
				now := time.Now()
				resetTime := c.lastIdentify.Add(5 * time.Second)
				time.Sleep(resetTime.Sub(now))
			}

			err := c.wsConn.WriteJSON(p)
			if err != nil {
				c.errorHandler(err)
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
		c.errorHandler(err)
	}

	err = c.wsConn.Close()
	if err != nil {
		c.errorHandler(err)
	}
}

type dataOpIdentify struct {
	Token          string             `json:"token"`
	Properties     identifyProperties `json:"properties"`
	Compress       bool               `json:"compress"`
	LargeThreshold int                `json:"large_threshold"`
	Shard          *[2]int            `json:"shard,omitempty"`
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
	p := &sentPayload{
		Operation: operationIdentify,
		Data: dataOpIdentify{
			Token: c.token,
			Properties: identifyProperties{
				OS:      runtime.GOOS,
				Browser: userAgent,
				Device:  userAgent,
			},
			Compress:       true,
			LargeThreshold: 250,
			Shard:          c.shard,
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
			Token:     c.token,
			SessionID: c.sessionID,
			Seq:       c.sequenceNumber,
		},
	}
	c.heartbeatMu.Unlock()
	c.write(ctx, p)
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
	Type           string          `json:"t,omitempty"` // omitempty for logging purposes
}

func (c *Conn) readLoop(ctx context.Context) {
	for {
		p, err := c.readPayload()
		if err != nil {
			if !isUseOfClosedError(err) {
				c.errorHandler(err)
				// It's possible we're being shutdown right now too.
				// Or maybe manager is already trying to reconnect.
				select {
				case c.reconnectChan <- struct{}{}:
				case <-ctx.Done():
				}
			}
			return
		}

		// TODO maybe make readPayload return JSON bytes too?
		b, err := json.MarshalIndent(p, "", "    ")
		if err != nil {
			panic(err)
		}
		c.logf("%s", b)

		err = c.onPayload(ctx, p)
		if err != nil {
			c.errorHandler(err)
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
		// TODO handle close frames!!! print out the error code. Probably with new web socket package.
	default:
		return nil, errors.New("unexpected websocket message type")
	}
	return &p, err
}

// TODO https://github.com/golang/go/issues/4373
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
			// We closed the connection and tried resuming but were too late.
			// If c.ready, the connection was not closed prior to resuming but rather we are
			// responding to our active session becoming expired and it turns out we were too late
			// to resume. In that case, we do not need to wait the random duration to stagger reconnects because
			// it won't help and we're not reconnecting. If we were too late to resume, it's safe to the gateway is
			// not under crazy load.
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
		return c.onDispatch(ctx, p)
	default:
		return errors.New("discord gone crazy; unexpected operation type")
	}
	return nil
}

func readEvent(p *receivedPayload) (interface{}, error) {
	e, err := getEventStruct(p.Type)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(p.Data, &e)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (c *Conn) onDispatch(ctx context.Context, p *receivedPayload) error {
	c.ready = true
	c.resuming = false

	c.heartbeatMu.Lock()
	c.sequenceNumber = p.SequenceNumber
	c.heartbeatMu.Unlock()

	e, err := readEvent(p)
	if err != nil {
		return &EventHandlerError{
			EventName: p.Type,
			Err:       err,
		}
	}

	if e, ok := e.(*EventReady); ok {
		c.sessionID = e.SessionID
	}

	if c.eventHandler == nil {
		return nil
	}

	err = c.eventHandler.Handle(ctx, e)
	if err != nil {
		return &EventHandlerError{
			EventName: p.Type,
			Event:     e,
			Err:       err,
		}
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
				c.log(err)
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

// Close closes the connection. It never returns an error.
// All errors will be handled by the ErrorHandler given in the DialConfig.
func (c *Conn) Close() error {
	c.closeChan <- struct{}{}
	<-c.closeChan
	return nil
}

func (c *Conn) log(v interface{}) {
	c.logf("%v", v)
}
