package discgo

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"bytes"

	"context"
	"time"
)

type RESTClient struct {
	Token      string
	HttpClient *http.Client

	rl *rateLimiter
}

const (
	apiVersion  = "6"
	endpointAPI = "https://discordapp.com/api/v" + apiVersion
	Version     = "0.1.0"
)

func (c *RESTClient) rootEndpoint() *endpoint {
	return &endpoint{c: c, url: endpointAPI}
}

var userAgent = fmt.Sprintf("DiscordBot (https://github.com/nhooyr/discgo, %v)", Version)

func (c *RESTClient) newRequest(ctx context.Context, method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Bot "+c.Token)
	req.Header.Set("User-Agent", userAgent)
	return req.WithContext(ctx)
}

func (c *RESTClient) do(req *http.Request, rateLimitPath string) ([]byte, error) {
	if c.rl == nil {
		c.rl = newRateLimiter()
		if c.HttpClient == nil {
			c.HttpClient = &http.Client{Timeout: 20 * time.Second}
		}
	}
	return c.doN(req, rateLimitPath, 0)
}

// TODO exponential backoff maybe? or too much in this library? not sure.
func (c *RESTClient) doN(req *http.Request, rateLimitPath string, n int) ([]byte, error) {
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

type endpoint struct {
	c             *RESTClient
	url           string
	rateLimitPath string
}

func (e *endpoint) append(urlElement, rateLimitPathElement string) *endpoint {
	e2 := &endpoint{e.c, e.url, e.rateLimitPath}
	e2.url += "/" + urlElement
	e2.rateLimitPath += "/" + rateLimitPathElement
	return e2
}

func (e *endpoint) appendMajor(element string) *endpoint {
	return e.append(element, element)
}

func (e *endpoint) appendMinor(element string) *endpoint {
	return e.append(element, "*")
}

func (e *endpoint) newRequest(ctx context.Context, method string, reqBody io.Reader) *http.Request {
	return e.c.newRequest(ctx, method, e.url, reqBody)
}

// Be careful with this method, it panics if json.Marshal errors.
func (e *endpoint) newRequestJSON(ctx context.Context, method string, v interface{}) *http.Request {
	body, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	req := e.c.newRequest(ctx, method, e.url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func (e *endpoint) do(req *http.Request, v interface{}) error {
	respBody, err := e.c.do(req, e.rateLimitPath)
	if err != nil || v == nil {
		return err
	}
	return json.Unmarshal(respBody, v)
}

func (e *endpoint) doMethod(ctx context.Context, method string, v1 interface{}, v2 interface{}) error {
	var req *http.Request
	if v1 == nil {
		req = e.newRequest(ctx, method, nil)
	} else {
		req = e.newRequestJSON(ctx, method, v1)
	}
	return e.do(req, v2)
}
