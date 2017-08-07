package discgo

import (
	"testing"
)

var uID = "97133780153683968"

func TestClient_GetUser(t *testing.T) {
	u, err := client.Me().Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(u.Username)
}

func TestClient_GetMyGuilds(t *testing.T) {
	guilds, err := client.Me().Guilds().Get(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, g := range guilds {
		t.Log(g.Name)
		t.Log(g.ID)
	}
}
