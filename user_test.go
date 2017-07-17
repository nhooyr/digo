package discgo_test

import "testing"

var uID = "97133780153683968"

func TestClient_GetUser(t *testing.T) {
	u, err := c.Me().Get()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(u.Username)
}

func TestClient_GetMyGuilds(t *testing.T) {
	guilds, err := c.Me().Guilds().Get(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, g := range guilds {
		t.Log(g.Name)
		t.Log(g.ID)
	}
}
