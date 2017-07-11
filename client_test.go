package discgo_test

import (
	"net/http"
	"os"
	"time"

	"github.com/nhooyr/discgo"
)

var c *discgo.Client

func init() {
	c = discgo.NewClient()
	c.Token = "Bot " + os.Getenv("DISCORD_TOKEN")
	c.HttpClient = &http.Client{
		Timeout: time.Second * 15,
	}
}