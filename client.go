package discgo

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"bytes"

	// TODO gotta use this package everywhere
	"github.com/pkg/errors"
)

const Version = "0.1.0"

var defaultUserAgent = fmt.Sprintf("DiscordBot (https://github.com/nhooyr/discgo, %v)", Version)

type Client struct {
	Token      string
	UserAgent  string
	HttpClient *http.Client

	rl *rateLimiter
}

func NewClient() *Client {
	return &Client{rl: newRateLimiter()}
}

func (c *Client) doSetHeaders(req *http.Request, rateLimitPath string) ([]byte, error) {
	req.Header.Set("Authorization", c.Token)
	if c.UserAgent == "" {
		req.Header.Set("User-Agent", defaultUserAgent)
	} else {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	return c.do(req, rateLimitPath, 0)
}

func (c *Client) do(req *http.Request, rateLimitPath string, n int) ([]byte, error) {
	prl := c.rl.getPathRateLimiter(rateLimitPath)
	prl.lock()
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		prl.unlock(nil)
		return nil, err
	}
	defer safeClose(resp.Body.Close, &err)
	err = prl.unlock(resp.Header)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
	case http.StatusNoContent:
	case http.StatusBadGateway:
		// TODO necessary?
		return c.do(req, rateLimitPath, n+1)
	case http.StatusTooManyRequests:
		return c.do(req, rateLimitPath, n)
	default:
		return nil, errors.Errorf("unexpected status code %q", resp.StatusCode)
	}
	return body, nil
}

func safeClose(closeFunc func() error, err *error) {
	cerr := closeFunc()
	if cerr != nil && *err == nil {
		*err = cerr
	}
}

const baseURL = "https://discordapp.com/api"

func newRequestJSON(method, path string, v interface{}) (*http.Request, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	req, err := newRequest(method, path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func newRequest(method, path string, body io.Reader) (*http.Request, error) {
	return http.NewRequest("GET", baseURL+path, body)
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
	req, err := newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doSetHeaders(req, path)
	if err != nil {
		return nil, err
	}
	return ch, json.Unmarshal(body, &ch)
}
