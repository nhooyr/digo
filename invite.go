package discgo

import (
	"time"
)

type ModelInvite struct {
	Code    string              `json:"code"`
	Guild   *ModelInviteGuild   `json:"guild"`
	Channel *ModelInviteChannel `json:"channel"`

	// Invite metadata
	Inviter   *ModelUser `json:"inviter"`
	Uses      int        `json:"uses"`
	MaxUses   int        `json:"max_uses"`
	MaxAge    int        `json:"max_age"`
	Temporary bool       `json:"temporary"`
	CreatedAt time.Time  `json:"created_at"`
	Revoked   bool       `json:"revoked"`
}

type ModelInviteGuild struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Splash *string `json:"splash"`
	Icon   *string `json:"icon"`
}

type ModelInviteChannel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type int    `json:"type"`
}

type EndpointInvite struct {
	*endpoint
}

func (c *Client) Invite(inviteCode string) EndpointInvite {
	e2 := c.e.appendMajor("invites").appendMinor(inviteCode)
	return EndpointInvite{e2}
}

func (e EndpointInvite) Get() (inv *ModelInvite, err error) {
	return inv, e.doMethod("GET", nil, &inv)
}

func (e EndpointInvite) Delete() (inv *ModelInvite, err error) {
	return inv, e.doMethod("DELETE", nil, &inv)
}

func (e EndpointInvite) Accept() (inv *ModelInvite, err error) {
	return inv, e.doMethod("POST", nil, &inv)
}
