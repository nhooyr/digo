package discgo

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"bytes"

	"time"
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
	return &Client{
		UserAgent:  defaultUserAgent,
		HttpClient: &http.Client{Timeout: 20 * time.Second},
		rl:         newRateLimiter(),
	}
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
	_, err := c.doN(req, rateLimitPath, 0)
	return err
}

func (c *Client) doUnmarshal(req *http.Request, rateLimitPath string, v interface{}) error {
	body, err := c.doN(req, rateLimitPath, 0)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, &v)
}

func (c *Client) doN(req *http.Request, rateLimitPath string, n int) ([]byte, error) {
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
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
	case http.StatusBadGateway:
		return c.doN(req, rateLimitPath, n+1)
	case http.StatusTooManyRequests:
		// Do not increment n because the next request should always be tried.
		return c.doN(req, rateLimitPath, n)
	default:
		apiErr := &APIError{
			Request:  req,
			Response: resp,
			Body:     body,
		}
		body := []byte{}
		// Ignore error because we may not have a error response at all.
		// And APIError.Error() will print the response body so if there is an
		// error in the JSON, it will be known.
		_ = json.Unmarshal(body, &apiErr.JSON)
		return nil, apiErr
	}
	return body, nil
}

func safeClose(closeFunc func() error, err *error) {
	cerr := closeFunc()
	if cerr != nil && *err == nil {
		*err = cerr
	}
}

type APIErrorJSON struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type APIError struct {
	Request  *http.Request
	Response *http.Response
	Body     []byte

	JSON *APIErrorJSON
}

func (err *APIError) Error() string {
	if err.JSON == nil {
		code := err.Response.StatusCode
		return fmt.Sprintf("Unexpected response %v %v, body: %q", code, http.StatusText(code), err.Body)
	}
	return fmt.Sprintf("Error code: %v, message: %v", err.JSON.Code, err.JSON.Message)
}
