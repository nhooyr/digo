package discgo

import (
	"os"
	"testing"
	"time"
)

func TestGateway_Get(t *testing.T) {
	e := client.Gateway()
	url, err := e.Get()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(url)
}

func TestConn_Connect(t *testing.T) {
	gatewayURL, err := client.Gateway().Get()
	if err != nil {
		t.Fatal(err)
	}
	config := NewDialConfig()
	config.GatewayURL = gatewayURL
	config.Token = os.Getenv("DISCORD_TOKEN")
	c, err := Dial(config)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 6)
	c.reconnectChan <- struct{}{}
	select {}
}
