package discgo

import (
	"testing"
	"time"
)

func EndpointTestGateway_Get(t *testing.T) {
	e := client.gateway()
	url, err := e.get()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(url)
}

func TestConn_Connect(t *testing.T) {
	config := NewDialConfig()
	config.Client = client
	c, err := Dial(config)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second*2)
	c.Close()
}
