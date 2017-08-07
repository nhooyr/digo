package discgo

import (
	"encoding/json"
	"os"
	"testing"
)

var (
	cID   = "331307058660114433"
	mID   = "342651739381432320"
	emoji = "🍰"
)

func TestClient_GetChannelMessages(t *testing.T) {
	params := &ParamsMessagesGet{
		Limit: 5,
	}
	messages, err := client.Channel(cID).Messages().Get(ctx, params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(messages[0].Content)
}

func TestClient_GetChannelMessage(t *testing.T) {
	m, err := client.Channel(cID).Message(mID).Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.MarshalIndent(m, "", "    ")
	t.Logf("%s", b)
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
	msg, err := client.Channel(cID).Messages().Create(ctx, params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(msg)
}

func TestClient_CreateReaction(t *testing.T) {
	err := client.Channel(cID).Message(mID).Reactions().Create(ctx, emoji)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_DeleteReaction(t *testing.T) {
	err := client.Channel(cID).Message(mID).Reaction(emoji, "@me").Delete(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_GetReactions(t *testing.T) {
	users, err := client.Channel(cID).Message(mID).Reactions().Get(ctx, emoji)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(users[0].Username)
}

func TestClient_DeleteReactions(t *testing.T) {
	err := client.Channel(cID).Message(mID).Reactions().Delete(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_UpdateMessage(t *testing.T) {
	params := &ParamsMessageEdit{
		Content: "updated wow",
	}
	m, err := client.Channel(cID).Message(mID).Edit(ctx, params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(m.Content)
}
