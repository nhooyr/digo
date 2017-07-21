package discgo

import (
	"encoding/json"
	"time"

	"sync"

	"runtime"

	"errors"
	"net"
	"os"

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
	token      string
	userAgent  string
	gatewayURL string

	sessionID string

	closeChan         chan struct{}
	confirmClosedChan chan struct{}
	reconnectChan     chan struct{}

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
		token:             apiClient.Token,
		userAgent:         apiClient.UserAgent,
		gatewayURL:        gatewayURL,
		confirmClosedChan: make(chan struct{}),
		reconnectChan:     make(chan struct{}),
	}, nil
}

const (
	dispatchOperation            = iota
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

func (c *Conn) close() error {
	closeMsg := websocket.FormatCloseMessage(websocket.CloseNoStatusReceived, "no heartbeat acknowledgment")
	err := c.wsConn.WriteMessage(websocket.CloseMessage, closeMsg)
	err2 := c.wsConn.Close()
	if err != nil {
		return err
	}
	return err2
}

func (c *Conn) Close() error {
	close(c.closeChan)
	err := c.close()
	<-c.confirmClosedChan
	<-c.confirmClosedChan
	return err
}

type helloOPData struct {
	HeartbeatInterval int      `json:"heartbeat_interval"`
	Trace             []string `json:"_trace"`
}

func (c *Conn) Dial() (err error) {
	c.heartbeatAcknowledged = true

	c.closeChan = make(chan struct{})

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

	go c.readLoop()

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
			Token:     c.token,
			SessionID: c.sessionID,
			Seq:       c.sequenceNumber,
		},
	}
	c.mu.Unlock()
	return c.wsConn.WriteJSON(p)
}

type readyEvent struct {
	V               int        `json:"v"`
	User            *User      `json:"user"`
	PrivateChannels []*Channel `json:"private_channels"`
	SessionID       string     `json:"session_id"`
	Trace           []string   `json:"_trace"`
}

type receivePayload struct {
	Operation      int             `json:"op"`
	Data           json.RawMessage `json:"d"`
	SequenceNumber int             `json:"s"`
	Type           string          `json:"t"`
}

func (c *Conn) nextPayload() (*receivePayload, error) {
	var p receivePayload
	// TODO compression, see how discordgo does it
	err := c.wsConn.ReadJSON(&p)
	return &p, err
}

func (c *Conn) readLoop() {
	for {
		// TODO somehow reuse payload
		p, err := c.nextPayload()
		if err != nil {
			if err, ok := err.(net.OpError); ok && err.Err == os.ErrClosed {
				c.confirmClosedChan <- struct{}{}
			} else {
				log.Print(err)
				c.reconnectChan <- struct{}{}
			}
			return
		}
		c.onEvent(p)
	}
}

type sendPayload struct {
	Operation int         `json:"op"`
	Data      interface{} `json:"d,omitempty"`
	Sequence  int         `json:"s,omitempty"`
}

func (c *Conn) onEvent(p *receivePayload) error {
	switch p.Operation {
	case helloOperation:
		var hello helloOPData
		err := json.Unmarshal(p.Data, &hello)
		if err != nil {
			return err
		}
		go c.eventLoop(hello.HeartbeatInterval)
	case heartbeatACKOperation:
		c.mu.Lock()
		c.heartbeatAcknowledged = true
		c.mu.Unlock()
	case invalidSessionOperation:
		// TODO only once do this or sleep too? not sure, confusing docs
		err := c.identify()
		if err != nil {
			return err
		}
	case dispatchOperation:
		c.mu.Lock()
		c.sequenceNumber = p.SequenceNumber
		c.mu.Unlock()

		switch p.Type {
		case "READY":
			var ready readyEvent
			err := json.Unmarshal(p.Data, &ready)
			if err != nil {
				return err
			}
			c.sessionID = ready.SessionID
		}
		// TODO state tracking
	default:
		panic("discord gone crazy")
	}
	return nil
}

func (c *Conn) eventLoop(heartbeatInterval int) {
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

				// Wait for readLoop to exit.
				// It's possible that it is trying to reconnect so receive on that channel too.
				select {
				case <-c.confirmClosedChan:
				case <-c.reconnectChan:
				}


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
			c.confirmClosedChan <- struct{}{}
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

	p := &sendPayload{Operation: heartbeatOperation, Data: sequenceNumber}
	return c.wsConn.WriteJSON(p)
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
			Token: c.token,
			Properties: identifyProperties{
				OS:      runtime.GOOS,
				Browser: c.userAgent,
			},
			// TODO COMPRESS!!!
			Compress:       false,
			LargeThreshold: 250,
		},
	}
	return c.wsConn.WriteJSON(p)
}
