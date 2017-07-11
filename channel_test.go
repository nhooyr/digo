package discgo_test

import (
	"testing"
	"os"
	"github.com/nhooyr/discgo"
)

var (
	cID = "331307058660114433"
	mID = "334104659767590912"
	emoji = "🍰"
)

func TestClient_GetChannelMessages(t *testing.T) {
	pgcms := &discgo.ParamsGetChannelMessages{
		Limit: 5,
	}
	msgs, err := c.GetChannelMessages(cID, pgcms)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(msgs[0].Content)
}

func TestClient_CreateMessage(t *testing.T) {
	f, err := os.Open("/Users/poonam566/Desktop/screenshot.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	cm := &discgo.ParamsCreateMessage{
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
	msg, err := c.CreateMessage(cID, cm)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(msg)
}

func TestClient_CreateReaction(t *testing.T) {
	err := c.CreateReaction(cID, mID, emoji)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_DeleteReaction(t *testing.T) {
	err := c.DeleteReaction(cID, mID, emoji, "@me")
	if err != nil {
		t.Fatal(err)
	}
}
