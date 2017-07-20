package discgo

import (
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
	Type channelType `json:"type"`
}

type InviteEndpoint struct {
	*endpoint
}

func (c *Client) Invite(inviteCode string) InviteEndpoint {
	e2 := c.e.appendMajor("invites").appendMinor(inviteCode)
	return InviteEndpoint{e2}
}

func (e InviteEndpoint) Get() (inv *Invite, err error) {
	return inv, e.doMethod("GET", nil, &inv)
}

func (e InviteEndpoint) Delete() (inv *Invite, err error) {
	return inv, e.doMethod("DELETE", nil, &inv)
}

func (e InviteEndpoint) Accept() (inv *Invite, err error) {
	return inv, e.doMethod("POST", nil, &inv)
}
