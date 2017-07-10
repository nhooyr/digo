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

	f, err := os.Open("/Users/nhooyr/Desktop/screenshot.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	cm := &discgo.CreateMessage{
		Content: "boar",
		File: &discgo.File{
			Name:    "screenshot.png",
			Content: f,
		},
		Embed: &discgo.Embed{
			Description: "heads",
			Image: &discgo.EmbedImage{
				URL: "attachment://screenshot.png",
			},
		},
	}
	msg, err := c.CreateMessage("331307058660114433", cm)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(msg)
}
