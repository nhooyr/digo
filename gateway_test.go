package discgo

import (
	"testing"
)

func EndpointTestGateway_Get(t *testing.T) {
	e := client.Gateway()
	url, err := e.get()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(url)
}

func TestConn_Connect(t *testing.T) {
	config := NewDialConfig()
	config.Client = client
	_, err := Dial(config)
	if err != nil {
		t.Fatal(err)
	}
	select {}
}
