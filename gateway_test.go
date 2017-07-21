package discgo

import (
	"testing"
)

func EndpointTestGateway_Get(t *testing.T) {
	e := c.gateway()
	url, err := e.get()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(url)
}

func TestConn_Connect(t *testing.T) {
	c, err := NewConn(c)
	if err != nil {
		t.Fatal(err)
	}
	err = c.Dial()
	if err != nil {
		t.Fatal(err)
	}
	select {}
}
