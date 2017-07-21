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
	c, err := NewConn(c)
	if err != nil {
		t.Fatal(err)
	}
	err = c.Dial()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second*3)
	err = c.Close()
	if err != nil {
		t.Fatal(err)
	}
}
