package discgo_test

import (
	"github.com/nhooyr/discgo"
	"os"
)

var c *discgo.Client

func init() {
	c = discgo.NewClient()
	// TODO don't like this, feels overly verbose
	// TODO wish I could make it all a simple struct lieke acme's autocert manager but thats so ugly :(
	c.Token = "Bot " + os.Getenv("DISCORD_TOKEN")
}
