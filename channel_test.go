package discgo_test

import (
	"github.com/nhooyr/discgo"
	"os"
	"testing"
)

var (
	cID   = "331307058660114433"
	mID   = "334104659767590912"
	emoji = "🍰"
)

func TestClient_GetChannelMessages(t *testing.T) {
	params := &discgo.GetMessagesParams{
		Limit: 5,
	}
	messages, err := c.Channel(cID).Messages().Get(params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(messages[0].Content)
}

func TestClient_CreateMessage(t *testing.T) {
	f, err := os.Open("/Users/poonam566/Desktop/screenshot.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	params := &discgo.CreateMessageParams{
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
	msg, err := c.Channel(cID).Messages().Create(params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(msg)
}

func TestClient_CreateReaction(t *testing.T) {
	err := c.Channel(cID).Message(mID).Reactions().Create(emoji)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_DeleteReaction(t *testing.T) {
	err := c.Channel(cID).Message(mID).Reaction(emoji, "@me").Delete()
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_GetReactions(t *testing.T) {
	users, err := c.GetReactions(cID, mID, emoji)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(users[0].Username)
}

func TestClient_DeleteReactions(t *testing.T) {
	err := c.DeleteReactions(cID, mID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_UpdateMessage(t *testing.T) {
	params := &discgo.MessageEditParams{
		Content: "updated wow",
	}
	m, err := c.EditMessage(cID, mID, params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(m.Content)
}
