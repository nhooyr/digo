package discgo

import (
	"os"
	"testing"
	"time"
)

func TestGateway_Get(t *testing.T) {
	e := client.Gateway()
	url, err := e.GetURL()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(url)
}

func TestConn_Connect(t *testing.T) {
	gatewayURL, err := client.Gateway().GetURL()
	if err != nil {
		t.Fatal(err)
	}

	c := &GatewayClient{
		Token:      os.Getenv("DISCORD_TOKEN"),
		GatewayURL: gatewayURL,
	}

	err = c.Connect()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 6)
	c.reconnectChan <- struct{}{}
	select {}
}
