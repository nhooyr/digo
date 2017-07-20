package discgo

import (
	"encoding/json"
	"time"

	"sync"

	"runtime"

	"github.com/gorilla/websocket"
	"github.com/nhooyr/log"
)

type Game struct {
	Name string  `json:"name"`
	Type *int    `json:"type"`
	URL  *string `json:"url"`
}

type Endpointgateway struct {
	*endpoint
}

func (c *Client) gateway() Endpointgateway {
	e2 := c.e.appendMajor("gateway")
	return Endpointgateway{e2}
}

func (g Endpointgateway) get() (url string, err error) {
	var urlStruct struct {
		URL string `json:"url"`
	}
	return urlStruct.URL, g.doMethod("GET", nil, &urlStruct)
}

type Conn struct {
	apiClient  *Client
	gatewayURL string
	sessionID  string

	closeOnce         sync.Once
	closeChan         chan struct{}
	confirmClosedChan chan struct{}

	wsConn *websocket.Conn

	mu                    sync.Mutex
	heartbeatAwknowledged bool
	sequenceNumber        *int
}

func NewConn(apiClient *Client) *Conn {
	return &Conn{
		apiClient:         apiClient,
		confirmClosedChan: make(chan struct{}, 2),
	}
}

const (
	dispatchOperation = iota
	heartbeatOperation
	identifyOperation
	statusUpdateOperation
	voiceStateUpdateOperation
	voiceServerPingOperation
	resumeOperation
	reconnectOperation
	requestGuildMembersOperation
	invalidSessionOperation
	helloOperation
	heartbeatACKOperation
)

func (c *Conn) Close() (err error) {
	c.confirmClosedChan <- struct{}{}
	c.closeOnce.Do(func() {
		close(c.closeChan)
		closeMsg := websocket.FormatCloseMessage(1001, "")
		err = c.wsConn.WriteMessage(websocket.CloseMessage, closeMsg)
		err2 := c.wsConn.Close()
		if err == nil {
			err = err2
		}

		// For heartbeat goroutine and eventLoop goroutine
		<-c.confirmClosedChan
		<-c.confirmClosedChan
	})
	return err
}

type helloOPData struct {
	HeartbeatInterval int      `json:"heartbeat_interval"`
	Trace             []string `json:"_trace"`
}

func (c *Conn) Connect() (err error) {
	c.heartbeatAwknowledged = true

	c.closeOnce = sync.Once{}
	c.closeChan = make(chan struct{})

	if c.gatewayURL == "" {
		c.gatewayURL, err = c.apiClient.gateway().get()
		if err != nil {
			return err
		}
		c.gatewayURL += "?v=" + apiVersion + "&encoding=json"
	}

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

	p, err := c.nextPayload()
	if err != nil {
		return err
	}

	var hello *helloOPData
	err = json.Unmarshal(p.Data, &hello)
	if err != nil {
		return err
	}
	// TODO remove
	go c.heartbeat(hello)
	go c.eventLoop()

	return nil
}

type resumeOPData struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	Seq       int    `json:"seq"`
}

func (c *Conn) resume() error {
	c.mu.Lock()
	p := &sendPayload{
		Operation: resumeOperation,
		Data: resumeOPData{
			Token:     c.apiClient.Token,
			SessionID: c.sessionID,
			Seq:       *c.sequenceNumber,
		},
	}
	c.mu.Unlock()
	b, _ := json.MarshalIndent(p, "", "  ")
	log.Printf("%s", b)
	return c.wsConn.WriteJSON(p)
}

type readyEvent struct {
	V               int        `json:"v"`
	User            *User      `json:"user"`
	PrivateChannels []*Channel `json:"private_channels"`
	SessionID       string     `json:"session_id"`
	Trace           []string   `json:"_trace"`
}

func (c *Conn) eventLoop() {
	for {
		p, err := c.nextPayload()
		if err != nil {
			_ = c.Close()
			return
		}

		c.mu.Lock()
		if p.SequenceNumber != nil {
			c.sequenceNumber = p.SequenceNumber
		}
		c.mu.Unlock()

		switch p.Operation {
		case dispatchOperation:
			switch p.Type {
			case "READY":
				var ready readyEvent
				err := json.Unmarshal(p.Data, &ready)
				if err != nil {
					_ = c.Close()
					return
				}
				c.sessionID = ready.SessionID
			}
			// TODO state tracking
		case heartbeatACKOperation:
			c.mu.Lock()
			// TODO change back
			c.heartbeatAwknowledged = false
			c.mu.Unlock()
		case invalidSessionOperation:
			err := c.identify()
			if err != nil {
				_ = c.Close()
				return
			}
		}
		log.Print(p.Operation)
		log.Print(p.Type)
		log.Printf("%s", p.Data)
		log.Print()
	}
}

type identifyOPData struct {
	Token          string             `json:"token"`
	Properties     identifyProperties `json:"properties"`
	Compress       bool               `json:"compress"`
	LargeThreshold int                `json:"large_threshold"`
}

type identifyProperties struct {
	OS              string `json:"$os,omitempty"`
	Browser         string `json:"$browser,omitempty"`
	Device          string `json:"$device,omitempty"`
	Referrer        string `json:"$referrer,omitempty"`
	ReferringDomain string `json:"$referring_domain,omitempty"`
}

func (c *Conn) identify() error {
	p := &sendPayload{
		Operation: identifyOperation,
		Data: identifyOPData{
			Token: c.apiClient.Token,
			Properties: identifyProperties{
				OS:      runtime.GOOS,
				Browser: c.apiClient.UserAgent,
			},
			// TODO COMPRESS!!!
			Compress:       false,
			LargeThreshold: 250,
		},
	}
	return c.wsConn.WriteJSON(p)
}

func (c *Conn) heartbeat(hello *helloOPData) {
	ticker := time.NewTicker(time.Duration(hello.HeartbeatInterval) * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-c.closeChan:
			return
		}
		c.mu.Lock()
		if !c.heartbeatAwknowledged {
			c.mu.Unlock()
			_ = c.Close()
			// TODO log error if unsuccessful connect/close
			log.Print("type something quick")
			time.Sleep(time.Second*5)
			err := c.Connect()
			log.Print(err)
			return
		}
		// TODO maybe loadint64?
		var sequenceNumberCopy *int
		// TODO use > 0 instead of pointer
		if c.sequenceNumber != nil {
			tmpCp := *c.sequenceNumber
			sequenceNumberCopy = &tmpCp
		}
		c.heartbeatAwknowledged = true
		c.mu.Unlock()

		p := &sendPayload{Operation: heartbeatOperation, Data: sequenceNumberCopy}
		err := c.wsConn.WriteJSON(p)
		if err != nil {
			// Log error? from close?
			_ = c.Close()
			return
		}
	}
}

type sendPayload struct {
	Operation int         `json:"op"`
	Data      interface{} `json:"d,omitempty"`
	Sequence  *int        `json:"s,omitempty"`
}

type receivePayload struct {
	Operation      int             `json:"op"`
	Data           json.RawMessage `json:"d"`
	SequenceNumber *int            `json:"s"`
	Type           string          `json:"t"`
}

func (c *Conn) nextPayload() (*receivePayload, error) {
	var p receivePayload
	// TODO compression, see how discordgo does it
	err := c.wsConn.ReadJSON(&p)
	return &p, err
}
