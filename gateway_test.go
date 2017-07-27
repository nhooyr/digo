package discgo

import (
	"testing"
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
	conn := NewConn()
	conn.Client = client
	err := conn.dial()
	if err != nil {
		t.Fatal(err)
	}
	select {}
}
