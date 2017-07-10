package discgo

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"github.com/nhooyr/log"
)

const (
	baseURL = "https://discordapp.com/api"
	version = "0.1.0"
)

var defaultUserAgent = fmt.Sprintf("DiscordBot (https://github.com/nhooyr/digo, %v)", version)

type Client struct {
	Token      string
	UserAgent  string
	HttpClient *http.Client

	rl *rateLimiter
}

func NewClient() *Client {
	return &Client{rl: newRateLimiter()}
}

func (c *Client) do(req *http.Request, rateLimitPath string) ([]byte, error) {
	req.Header.Set("Authorization", c.Token)
	if c.UserAgent == "" {
		req.Header.Set("User-Agent", defaultUserAgent)
	} else {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	prl := c.rl.getPathRateLimiter(rateLimitPath)
	prl.lock()
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		prl.unlock(nil)
		return nil, err
	}
	defer safeClose(resp.Body.Close, &err)
	err = prl.unlock(resp.Header)
	log.Print(resp.Header)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// TODO maybe somehow close resp.Body here? Cannot close twice though :(
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
	case http.StatusNoContent:
	case http.StatusTooManyRequests:
		// TODO try again soon
	default:
		panic(resp.StatusCode)
	}
	return body, nil
}

func safeClose(closeFunc func() error, err *error) {
	cerr := closeFunc()
	if cerr != nil && *err == nil {
		*err = cerr
	}
}

func newRequest(method, path string, body io.Reader) *http.Request {
	req, err := http.NewRequest("GET", baseURL+path, nil)
	if err != nil {
		panic(err)
	}
	return req
}

// TODO I don't always need the baseURL prefix
// TODO think about a way to access the api endpoints elegantly
func apiPath(elements ...string) (path string) {
	for _, e := range elements {
		path += "/" + e
	}
	return path
}

func (c *Client) GetChannel(id string) (ch *Channel, err error) {
	path := apiPath("channels", id)
	req := newRequest("GET", path, nil)
	body, err := c.do(req, path)
	if err != nil {
		return nil, err
	}
	return ch, json.Unmarshal(body, &ch)
}
