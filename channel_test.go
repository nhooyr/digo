package discgo

import (
	"context"
	"os"
	"testing"
)

var (
	cID   = "331307058660114433"
	mID   = "334104659767590912"
	emoji = "🍰"
)

func TestClient_GetChannelMessages(t *testing.T) {
	params := &ParamsMessagesGet{
		Limit: 5,
	}
	messages, err := client.Channel(cID).Messages().Get(context.Background(), params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(messages[0].Content)
}

func TestClient_CreateMessage(t *testing.T) {
	f, err := os.Open("/Users/nhooyr/Desktop/screenshot.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	params := &ParamsMessageCreate{
		Content: "boar",
		File: &ParamsFile{
			Name:    "screenshot.png",
			Content: f,
		},
		Embed: &ModelEmbed{
			Description: "heads",
			Image: &ModelEmbedImage{
				URL: "attachment://screenshot.png",
			},
		},
	}
	msg, err := client.Channel(cID).Messages().Create(context.Background(), params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(msg)
}

func TestClient_CreateReaction(t *testing.T) {
	err := client.Channel(cID).Message(mID).Reactions().Create(context.Background(), emoji)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_DeleteReaction(t *testing.T) {
	err := client.Channel(cID).Message(mID).Reaction(emoji, "@me").Delete(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_GetReactions(t *testing.T) {
	users, err := client.Channel(cID).Message(mID).Reactions().Get(context.Background(), emoji)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(users[0].Username)
}

func TestClient_DeleteReactions(t *testing.T) {
	err := client.Channel(cID).Message(mID).Reactions().Delete(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_UpdateMessage(t *testing.T) {
	params := &ParamsMessageEdit{
		Content: "updated wow",
	}
	m, err := client.Channel(cID).Message(mID).Edit(context.Background(), params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(m.Content)
}
