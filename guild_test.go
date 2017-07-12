package discgo_test

import (
	"github.com/nhooyr/discgo"
	"testing"
)

var gID = "331307058660114433"

func TestClient_CreateGuild(t *testing.T) {
	params := discgo.ParamsCreateGuild{
		Name: "REKTERONIED",
	}
	g, err := c.CreateGuild(params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(g.ID)
}

func TestClient_DeleteGuild(t *testing.T) {
	// TODO doesn't return any JSON for some reason?
	g, err := c.DeleteGuild("334474961580195852")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(g.ID)
}

func TestClient_GetChannels(t *testing.T) {
	channels, err := c.GetChannels(gID)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(channels[3].Name)
}

func TestClient_GetGuildMember(t *testing.T) {
	gm, err := c.GetGuildMember(gID, uID)
	if err != nil {
		t.Fatal(err)
	}
	if gm.Nick != nil {
		t.Log(*gm.Nick)
	} else {
		t.Log(gm.Nick)
	}
}

func TestClient_GetGuildMembers(t *testing.T) {
	guildMembers, err := c.GetGuildMembers(gID, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(guildMembers[0].User.Username)
}

func TestClient_ModifyGuildMember(t *testing.T) {
	params := &discgo.ParamsModifyGuildMember{
		Nick: "ggez",
	}
	err := c.ModifyGuildMember(gID, "97137587013050368", params)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_ModifyMyNick(t *testing.T) {
	nick, err := c.ModifyMyNick(gID, "LOL REKT")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(nick)
}
