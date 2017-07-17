package discgo_test

import (
	"github.com/nhooyr/discgo"
	"testing"
)

var gID = "331307058660114433"

func TestClient_CreateGuild(t *testing.T) {
	params := &discgo.GuildsCreateParams{
		Name: "REKTERONIED",
	}
	g, err := c.Guilds().Create(params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(g.ID)
}

func TestClient_DeleteGuild(t *testing.T) {
	g, err := c.Guild(gID).Delete()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(g.ID)
}

func TestClient_GetChannels(t *testing.T) {
	channels, err := c.Guild(gID).Channels().Get()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(channels[3].Name)
}

func TestClient_GetGuildMember(t *testing.T) {
	gm, err := c.Guild(gID).Member(uID).Get()
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
	guildMembers, err := c.Guild(gID).Members().Get(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(guildMembers[0].User.Username)
}

func TestClient_ModifyGuildMember(t *testing.T) {
	params := &discgo.GuildMemberModifyParams{
		Nick: "fdkg",
	}
	err := c.Guild(gID).Member(uID).Modify(params)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_ModifyMyNick(t *testing.T) {
	nick, err := c.Guild(gID).Me().ModifyNick("xd RssEKT")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(nick)
}
