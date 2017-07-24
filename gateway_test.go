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
	conn := NewConn()
	conn.Client = client
	err := conn.Dial()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 2)
	conn.reconnectChan <- struct{}{}
	select {}
}
