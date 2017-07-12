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
	return &Client{rl: newRateLimiter(), UserAgent: defaultUserAgent}
}

const endpointAPI = "https://discordapp.com/api/"

func (c *Client) newRequest(method, endpoint string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, endpointAPI+endpoint, body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", c.Token)
	req.Header.Set("User-Agent", c.UserAgent)
	return req
}

func (c *Client) newRequestJSON(method, endpoint string, v interface{}) *http.Request {
	body, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	req := c.newRequest(method, endpoint, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func (c *Client) do(req *http.Request, rateLimitPath string) error {
	_, err := c.do_n(req, rateLimitPath, 0)
	return err
}

func (c *Client) do_unmarshal(req *http.Request, rateLimitPath string, v interface{}) error {
	body, err := c.do_n(req, rateLimitPath, 0)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, &v)
}

func (c *Client) do_n(req *http.Request, rateLimitPath string, n int) ([]byte, error) {
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
		return c.do_n(req, rateLimitPath, n+1)
	case http.StatusTooManyRequests:
		return c.do_n(req, rateLimitPath, n)
	default:
		return nil, errors.Errorf("unexpected status code %v %v", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	return body, nil
}

func safeClose(closeFunc func() error, err *error) {
	cerr := closeFunc()
	if cerr != nil && *err == nil {
		*err = cerr
	}
}
