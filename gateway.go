package discgo

import (
	"encoding/json"
	"time"

	"sync"

	"runtime"

	"github.com/gorilla/websocket"
)

type Game struct {
	Name string  `json:"name"`
	Type *int    `json:"type"`
	URL  *string `json:"url"`
}

type gatewayEndpoint struct {
	*endpoint
}

func (c *Client) gateway() gatewayEndpoint {
	e2 := c.e.appendMajor("gateway")
	return gatewayEndpoint{e2}
}

func (g gatewayEndpoint) get() (url string, err error) {
	var urlStruct struct {
		URL string `json:"url"`
	}
	return urlStruct.URL, g.doMethod("GET", nil, &urlStruct)
}

type Conn struct {
	apiClient  *Client
	gatewayURL string

	closeOnce sync.Once
	closeChan chan struct{}

	wsConn *websocket.Conn

	mu               sync.Mutex
	hearbeatInFlight bool
	sequenceNumber   *int
}

func NewConn(apiClient *Client) *Conn {
	return &Conn{
		apiClient:             apiClient,
		closeChan:             make(chan struct{}),
	}
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

func (c *Conn) Close() {
	c.closeOnce.Do(func() {
		close(c.closeChan)
	})
}

type helloOPData struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
	// There is also the _trace field but what should be done with it?
}

func (c *Conn) Connect() (err error) {
	c.gatewayURL, err = c.apiClient.gateway().get()
	if err != nil {
		return err
	}
	c.gatewayURL += "?v=" + apiVersion + "&encoding=json"
	c.wsConn, _, err = websocket.DefaultDialer.Dial(c.gatewayURL, nil)
	if err != nil {
		return err
	}

	err = c.identify()
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
	go c.heartbeat(hello)
	go c.eventLoop()

	return nil
}

func (c *Conn) eventLoop()  {

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
		if c.hearbeatInFlight {
			// TODO gotta terminate and then reconnect
			panic("TODO")
		}
		// TODO maybe loadint64?
		sequenceNumberCopy := c.sequenceNumber
		c.hearbeatInFlight = true
		c.mu.Unlock()

		p := &sendPayload{Operation: heartbeatOperation, Data: sequenceNumberCopy}
		err := c.wsConn.WriteJSON(p)
		if err != nil {
			c.Close()
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
	Operation int             `json:"op"`
	Data      json.RawMessage `json:"d"`
	Sequence  int             `json:"s"`
	Type      string          `json:"t"`
}

func (c *Conn) nextPayload() (*receivePayload, error) {
	var p receivePayload
	// TODO compression, see how discordgo does it
	err := c.wsConn.ReadJSON(&p)
	return &p, err
}
