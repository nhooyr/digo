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

	closeOnce           sync.Once
	closeChan           chan struct{}
	closeConfirmChannel chan struct{}

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
		token:      apiClient.Token,
		userAgent:  apiClient.UserAgent,
		gatewayURL: gatewayURL,
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
	close(c.closeChan)
	closeMsg := websocket.FormatCloseMessage(websocket.CloseNoStatusReceived, "no heartbeat acknowledgment")
	err := c.wsConn.WriteMessage(websocket.CloseMessage, closeMsg)
	err2 := c.wsConn.Close()
	if err != nil {
		return err
	}
	return err2
}

func (c *Conn) Close() error {
	err := c.close()
	// Wait for eventloop and heartbeat goroutines.
	<-c.closeConfirmChannel
	<-c.closeConfirmChannel
	return err
}

func (c *Conn) reconnect() {
	c.closeOnce.Do(func() {
		err := c.close()
		if err != nil {
			log.Print(err)
		}
		// Wait for eventloop or heartbeat goroutine.
		<-c.closeConfirmChannel
		err = c.Dial()
		if err != nil {
			log.Print(err)
		}
	})
}

type helloOPData struct {
	HeartbeatInterval int      `json:"heartbeat_interval"`
	Trace             []string `json:"_trace"`
}

func (c *Conn) Dial() (err error) {
	c.heartbeatAcknowledged = true

	c.closeOnce = sync.Once{}
	c.closeChan = make(chan struct{})
	c.closeConfirmChannel = make(chan struct{})

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

func (c *Conn) eventLoop() {
	for {
		p, err := c.nextPayload()
		if err != nil {
			select {
			case <-c.closeChan:
				c.closeConfirmChannel <- struct{}{}
			default:
				// TODO use sync.Once to prevent race condition?
				c.reconnect()
			}
			return
		}

		switch p.Operation {
		case helloOperation:
			var hello helloOPData
			err = json.Unmarshal(p.Data, &hello)
			if err != nil {
				err = c.close()
				if err != nil {
					log.Print(err)
				}
				return
			}
			go c.heartbeat(&hello)
		case heartbeatACKOperation:
			c.mu.Lock()
			c.heartbeatAcknowledged = true
			c.mu.Unlock()
		case invalidSessionOperation:
			// TODO only once do this or sleep too? not sure, confusing docs
			err := c.identify()
			if err != nil {
				log.Print(err)
				continue
			}
		case dispatchOperation:
			c.mu.Lock()
			c.sequenceNumber = p.SequenceNumber
			c.mu.Unlock()

			switch p.Type {
			case "READY":
				var ready readyEvent
				err = json.Unmarshal(p.Data, &ready)
				if err != nil {
					err = c.close()
					if err != nil {
						log.Print(err)
					}
					return
				}
				c.sessionID = ready.SessionID
			}
			// TODO state tracking
		}
		log.Print(p.Operation)
		log.Print(p.Type)
		log.Printf("%s", p.Data)
		log.Print(p.SequenceNumber)
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

func (c *Conn) heartbeat(hello *helloOPData) {
	ticker := time.NewTicker(time.Duration(hello.HeartbeatInterval) * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-c.closeChan:
			c.closeConfirmChannel <- struct{}{}
			return
		}
		c.mu.Lock()
		if !c.heartbeatAcknowledged {
			c.mu.Unlock()
			c.reconnect()
			return
		}
		sequenceNumber := c.sequenceNumber
		c.heartbeatAcknowledged = false
		c.mu.Unlock()

		p := &sendPayload{Operation: heartbeatOperation, Data: sequenceNumber}
		err := c.wsConn.WriteJSON(p)
		if err != nil {
			log.Print(err)
			err = c.close()
			if err != nil {
				log.Print(err)
			}
			return
		}
	}
}

type sendPayload struct {
	Operation int         `json:"op"`
	Data      interface{} `json:"d,omitempty"`
	Sequence  int         `json:"s,omitempty"`
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
