package discgo

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestGateway_Get(t *testing.T) {
	e := client.Gateway()
	url, err := e.GetURL(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(url)
}

func TestConn_Connect(t *testing.T) {
	gatewayURL, err := client.Gateway().GetURL(ctx)
	if err != nil {
		t.Fatal(err)
	}

	s := new(State)
	c := &GatewayClient{
		Token:      os.Getenv("DISCORD_TOKEN"),
		GatewayURL: gatewayURL,
		EventHandler: EventHandlerFunc(func(ctx context.Context, e interface{}) error {
			err := s.handle(e)
			if err != nil {
				return err
			}
			switch e := e.(type) {
			case *EventMessageCreate:
				t.Log(e.Content)
			}
			return nil
		}),
	}

	err = c.Connect()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 6)
	c.reconnectChan <- struct{}{}
	time.Sleep(time.Second * 20)
}
