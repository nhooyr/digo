package discgo

import (
	"compress/zlib"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nhooyr/log"
)

type EndpointGateway struct {
	*endpoint
}

func (c *RESTClient) Gateway() EndpointGateway {
	e2 := c.rootEndpoint().appendMajor("gateway")
	return EndpointGateway{e2}
}

// TODO I don't think a ctx is necessary, ever.
func (e EndpointGateway) GetURL(ctx context.Context) (url string, err error) {
	var urlStruct struct {
		URL string `json:"url"`
	}
	return urlStruct.URL, e.doMethod(ctx, "GET", nil, &urlStruct)
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

// TODO better name?
func (e EndpointGatewayBot) Get(ctx context.Context) (resp *ModelGatewayBotGetResponse, err error) {
	return resp, e.doMethod(ctx, "GET", nil, &resp)
}

type EventHandler interface {
	Handle(ctx context.Context, e interface{}) error
}

// Returned by a EventHandler to signal that the event should not be handled further.
// This is used by State to prevent GuildCreate events handlers from running when a guild becomes available again.
var ErrEventDone = errors.New("event is done; no need to handle the event further")

type EventHandlerFunc func(ctx context.Context, e interface{}) error

func (h EventHandlerFunc) Handle(ctx context.Context, e interface{}) error {
	return h(ctx, e)
}

type GatewayClient struct {
	// Configuration. Maybe extract into GatewayClientConfig?
	Token        string
	GatewayURL   string
	EventHandler EventHandler
	ErrorHandler func(err error)
	Logf         func(format string, v ...interface{})
	Debug        bool // enables logging of events.
	Shard        int
	ShardCount   int

	sessionID    string
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

func (c *GatewayClient) log(v interface{}) {
	c.Logf("%v", v)
}

func (c *GatewayClient) Connect() error {
	if c.GatewayURL == "" {
		panic("missing gateway URL")
	}
	if c.Token == "" {
		panic("missing API token")
	}

	if c.Logf == nil {
		c.Logf = func(f string, v ...interface{}) {
			log.Printf(f, v...)
		}
	}

	if c.ErrorHandler == nil {
		c.ErrorHandler = func(err error) {
			c.log(err)
		}
	}

	c.GatewayURL += "?v=" + apiVersion + "&encoding=json"
	c.closeChan = make(chan struct{})
	c.reconnectChan = make(chan struct{})
	c.writeChan = make(chan *sentPayload)
	c.ready = true

	return c.connect()
}

// If it errors without receiving a operationDispatch payload, then
// this method will always return an error.
func (c *GatewayClient) connect() error {
	if !c.ready {
		return errors.New("already tried to connect and failed")
	}

	c.ready = false
	c.resuming = false
	c.heartbeatAcknowledged = true
	c.lastIdentify = time.Time{}

	c.Logf("connecting")
	// TODO Need to set read deadline for hello packet and I also need to set write deadlines.
	// TODO also max message
	var err error
	c.wsConn, _, err = websocket.DefaultDialer.Dial(c.GatewayURL, nil)
	if err != nil {
		return err
	}

	go c.manager()

	return nil
}

func (c *GatewayClient) manager() {
	ctx, cancelFn := context.WithCancel(context.Background())
	c.runWorker(func() {
		c.writeLoop(ctx)
	})
	c.runWorker(func() {
		c.readLoop(ctx)
	})

	if c.sessionID == "" {
		c.Logf("identifying")
		c.identify(ctx)
	} else {
		c.Logf("resuming")
		c.resume(ctx)
	}

	select {
	case <-c.reconnectChan:
		c.Logf("restarting")

		cancelFn()
		c.wg.Wait()

		err := c.connect()
		if err != nil {
			c.log(err)
		}
	case <-c.closeChan:
		c.Logf("exiting")

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

func (c *GatewayClient) writeLoop(ctx context.Context) {
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
				c.ErrorHandler(err)
				select {
				case c.reconnectChan <- struct{}{}:
				case <-ctx.Done():
				}
				break writeLoop
			}

			if c.Debug {
				b, err := json.MarshalIndent(p, "", "    ")
				if err != nil {
					panic(err)
				}
				c.Logf("write: %s", b)
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
		c.ErrorHandler(err)
	}

	err = c.wsConn.Close()
	if err != nil {
		c.ErrorHandler(err)
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

func (c *GatewayClient) write(ctx context.Context, p *sentPayload) {
	select {
	case c.writeChan <- p:
	case <-ctx.Done():
	}
}

func (c *GatewayClient) identify(ctx context.Context) {
	p := &sentPayload{
		Operation: operationIdentify,
		Data: dataOpIdentify{
			Token: c.Token,
			Properties: identifyProperties{
				OS:      runtime.GOOS,
				Browser: userAgent,
				Device:  userAgent,
			},
			Compress:       true,
			LargeThreshold: 250,
		},
	}

	if c.ShardCount > 1 {
		p.Data.(*dataOpIdentify).Shard = &[2]int{c.Shard, c.ShardCount}
	}

	c.write(ctx, p)
}

type dataOpResume struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	Seq       int    `json:"seq"`
}

func (c *GatewayClient) resume(ctx context.Context) {
	c.resuming = true

	c.heartbeatMu.Lock()
	p := &sentPayload{
		Operation: operationResume,
		Data: dataOpResume{
			Token:     c.Token,
			SessionID: c.sessionID,
			Seq:       c.sequenceNumber,
		},
	}
	c.heartbeatMu.Unlock()
	c.write(ctx, p)
}

// TODO maybe export?
func (c *GatewayClient) runWorker(fn func()) {
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

func (c *GatewayClient) readLoop(ctx context.Context) {
	for {
		p, err := c.readPayload()
		if err != nil {
			if !isUseOfClosedError(err) {
				c.ErrorHandler(err)
				// It's possible we're being shutdown right now too.
				// Or maybe manager is already trying to reconnect.
				select {
				case c.reconnectChan <- struct{}{}:
				case <-ctx.Done():
				}
			}
			return
		}

		if c.Debug {
			b, err := json.MarshalIndent(p, "", "    ")
			if err != nil {
				panic(err)
			}
			c.Logf("read: %s", b)
		}

		err = c.onPayload(ctx, p)
		if err != nil {
			c.ErrorHandler(err)
			// It's possible we're being shutdown right now.
			// Or maybe manager is already trying to reconnect.
			select {
			case c.reconnectChan <- struct{}{}:
			case <-ctx.Done():
			}
		}
	}
}

func (c *GatewayClient) readPayload() (*receivedPayload, error) {
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

func (c *GatewayClient) onPayload(ctx context.Context, p *receivedPayload) error {
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

func (c *GatewayClient) onDispatch(ctx context.Context, p *receivedPayload) error {
	c.ready = true
	c.resuming = false

	c.heartbeatMu.Lock()
	c.sequenceNumber = p.SequenceNumber
	c.heartbeatMu.Unlock()

	e, err := readEvent(p)
	if err != nil {
		return &EventHandlerError{
			EventName: p.Type,
			Event:     e,
			Err:       err,
		}
	}

	switch e := e.(type) {
	case *EventReady:
		c.Logf("ready")
		c.sessionID = e.SessionID
	case *eventResumed:
		c.Logf("resumed")
	}

	if c.EventHandler == nil {
		return nil
	}

	err = c.EventHandler.Handle(ctx, e)
	if err != nil {
		// Possible someone forgot to handle ErrEventDone.
		if err == ErrEventDone {
			return nil
		}
		return &EventHandlerError{
			EventName: p.Type,
			Event:     e,
			Err:       err,
		}
	}
	return nil
}

func (c *GatewayClient) heartbeatLoop(ctx context.Context, heartbeatInterval int) {
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

func (c *GatewayClient) heartbeat() error {
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
// All errors will be handled by the ErrorHandler given in the GatewayClientConfig.
func (c *GatewayClient) Close() error {
	c.closeChan <- struct{}{}
	<-c.closeChan
	return nil
}
