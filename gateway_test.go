package discgo

import (
	"testing"
	"time"
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
	c := NewConn(c)
	err := c.Connect()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 41*4)
}
