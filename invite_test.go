package discgo

import "testing"

var inviteCode = "NP9NQ8v"

func TestClient_GetInvite(t *testing.T) {
	inv, err := c.Invite(inviteCode).Get()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(inv.Code)
}
