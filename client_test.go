package discgo_test

import (
	"github.com/nhooyr/discgo"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	c := discgo.NewClient()
	c.Token = "Bot " + os.Getenv("DISCORD_TOKEN")
	c.HttpClient = &http.Client{
		Timeout: time.Second * 15,
	}
	msg, err := c.GetChannelMessage("331307058660114433", "333262458955366400")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(msg)
}
