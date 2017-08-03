package discgo

import (
	"context"
	"testing"
)

var inviteCode = "NP9NQ8v"

func TestClient_GetInvite(t *testing.T) {
	inv, err := client.Invite(inviteCode).Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(inv.Code)
}
