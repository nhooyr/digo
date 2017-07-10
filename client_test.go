package discgo_test

import (
	"net/http"
	"os"
	"testing"
	"time"
	"github.com/nhooyr/discgo"
)

func TestClient(t *testing.T) {
	c := discgo.NewClient()
	c.Token = "Bot " + os.Getenv("DISCORD_TOKEN")
	c.HttpClient = &http.Client{
		Timeout: time.Second * 15,
	}
	for i := 0; i < 1000; i++ {
		ch, err := c.GetChannel("331307058660114433")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(ch)
	}
}
