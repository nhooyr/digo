package discgo

import (
	"path"
	"time"
)

type Invite struct {
	Code    string         `json:"code"`
	Guild   *InviteGuild   `json:"guild"`
	Channel *InviteChannel `json:"channel"`

	// Invite metadata
	Inviter   *User     `json:"inviter"`
	Uses      int       `json:"uses"`
	MaxUses   int       `json:"max_uses"`
	MaxAge    int       `json:"max_age"`
	Temporary bool      `json:"temporary"`
	CreatedAt time.Time `json:"created_at"`
	Revoked   bool      `json:"revoked"`
}

type InviteGuild struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Splash *string `json:"splash"`
	Icon   *string `json:"icon"`
}

type InviteChannel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func (c *Client) GetInvite(inviteCode string) (inv *Invite, err error) {
	endpoint := path.Join("invites", inviteCode)
	req := c.newRequest("GET", endpoint, nil)
	rateLimitPath := path.Join("invites", "*")
	return inv, c.doUnmarshal(req, rateLimitPath, &inv)
}

func (c *Client) DeleteInvite(inviteCode string) (inv *Invite, err error) {
	endpoint := path.Join("invites", inviteCode)
	req := c.newRequest("DELETE", endpoint, nil)
	rateLimitPath := path.Join("invites", "*")
	return inv, c.doUnmarshal(req, rateLimitPath, &inv)
}

func (c *Client) AcceptInvite(inviteCode string) (inv *Invite, err error) {
	endpoint := path.Join("invites", inviteCode)
	req := c.newRequest("POST", endpoint, nil)
	rateLimitPath := path.Join("invites", "*")
	return inv, c.doUnmarshal(req, rateLimitPath, &inv)
}
