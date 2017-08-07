package discgo

import (
	"context"
	"time"

	"gopkg.in/guregu/null.v3"
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
type EndpointInvites struct {
	*endpoint
}

func (e EndpointChannel) Invites() EndpointInvites {
	e2 := e.appendMajor("invites")
	return EndpointInvites{e2}
}

func (e EndpointInvites) Get(ctx context.Context) (invites []*ModelInvite, err error) {
	return invites, e.doMethod(ctx, "GET", nil, &invites)
}

type ParamsInviteCreate struct {
	MaxAge    null.Int `json:"max_age"`
	MaxUses   null.Int `json:"max_uses"`
	Temporary bool     `json:"temporary,omitempty"`
	Unique    bool     `json:"unique,omitempty"`
}

func (e EndpointInvites) Create(ctx context.Context, params *ParamsInviteCreate) (invite *ModelInvite, err error) {
	return invite, e.doMethod(ctx, "POST", params, &invite)
}

type EndpointInvite struct {
	*endpoint
}

func (c *RESTClient) Invite(inviteCode string) EndpointInvite {
	e2 := c.rootEndpoint().appendMajor("invites").appendMinor(inviteCode)
	return EndpointInvite{e2}
}

func (e EndpointInvite) Get(ctx context.Context) (inv *ModelInvite, err error) {
	return inv, e.doMethod(ctx, "GET", nil, &inv)
}

func (e EndpointInvite) Delete(ctx context.Context) (inv *ModelInvite, err error) {
	return inv, e.doMethod(ctx, "DELETE", nil, &inv)
}

func (e EndpointInvite) Accept(ctx context.Context) (inv *ModelInvite, err error) {
	return inv, e.doMethod(ctx, "POST", nil, &inv)
}
