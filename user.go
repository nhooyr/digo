package discgo

import (
	"context"
	"net/url"
	"strconv"
)

type ModelUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
	Bot           bool   `json:"bot"`
	MFAEnabled    bool   `json:"mfa_enabled"`
	Verified      bool   `json:"verified"`
	Email         string `json:"email"`
}

type ModelUserGuild struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Owner       bool   `json:"owner"`
	Permissions int    `json:"permissions"`
}

type ModelConnection struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Types        string              `json:"types"`
	Revoked      bool                `json:"revoked"`
	Integrations []*ModelIntegration `json:"integrations"`
}

type EndpointUser struct {
	*endpoint
}

func (c *RESTClient) User(uID string) EndpointUser {
	e2 := c.rootEndpoint().appendMajor("users").appendMinor(uID)
	return EndpointUser{e2}
}

func (e EndpointUser) Get(ctx context.Context) (u *ModelUser, err error) {
	return u, e.doMethod(ctx, "GET", nil, &u)
}

type EndpointMe struct {
	*endpoint
}

func (c *RESTClient) Me() EndpointMe {
	e2 := c.User("@me").endpoint
	return EndpointMe{e2}
}

func (e EndpointMe) Get(ctx context.Context) (u *ModelUser, err error) {
	return u, e.doMethod(ctx, "GET", nil, &u)
}

type ParamsMeModify struct {
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

func (e EndpointMe) Modify(ctx context.Context, params *ParamsMeModify) (u *ModelUser, err error) {
	return u, e.doMethod(ctx, "PATCH", params, &u)
}

type EndpointMeGuilds struct {
	*endpoint
}

func (e EndpointMe) Guilds() EndpointMeGuilds {
	e2 := e.appendMajor("guilds")
	return EndpointMeGuilds{e2}
}

type ParamsMeGuildsGet struct {
	BeforeID string
	AfterID  string
	Limit    int
}

func (params *ParamsMeGuildsGet) rawQuery() string {
	v := url.Values{}
	if params.BeforeID != "" {
		v.Set("before", params.BeforeID)
	}
	if params.AfterID != "" {
		v.Set("after", params.AfterID)
	}
	if params.Limit > 0 {
		v.Set("limit", strconv.Itoa(params.Limit))
	}
	return v.Encode()
}

func (e EndpointMeGuilds) Get(ctx context.Context, params *ParamsMeGuildsGet) (guilds []*ModelUserGuild, err error) {
	req := e.newRequest(ctx, "GET", nil)
	if params != nil {
		req.URL.RawQuery = params.rawQuery()
	}
	return guilds, e.do(req, &guilds)
}

type EndpointMeGuild struct {
	*endpoint
}

func (e EndpointMe) Guild(gID string) EndpointMeGuild {
	e2 := e.Guilds().appendMajor(gID)
	return EndpointMeGuild{e2}
}

func (e EndpointMeGuild) Leave(ctx context.Context) error {
	return e.doMethod(ctx, "DELETE", nil, nil)
}

type EndpointMeDMChannels struct {
	*endpoint
}

func (e EndpointMe) DMChannels() EndpointMeDMChannels {
	e2 := e.appendMajor("channels")
	return EndpointMeDMChannels{e2}
}

func (e EndpointMeDMChannels) Get(ctx context.Context) (dmChannels *[]ModelChannel, err error) {
	return dmChannels, e.doMethod(ctx, "GET", nil, &dmChannels)
}

type ParamsDMChannelsCreate struct {
	RecipientID string `json:"recipient_id"`
}

func (e EndpointMeDMChannels) Create(ctx context.Context, params *ParamsDMChannelsCreate) (ch *ModelChannel, err error) {
	return ch, e.doMethod(ctx, "POST", params, &ch)
}

type ParamsDmChannelsCreateGroup struct {
	AccessTokens []string          `json:"access_tokens"`
	Nicks        map[string]string `json:"nicks"`
}

func (e EndpointMeDMChannels) CreateGroup(ctx context.Context, params *ParamsDmChannelsCreateGroup) (ch *ModelChannel, err error) {
	return ch, e.doMethod(ctx, "POST", params, &ch)
}

type EndpointMeConnections struct {
	*endpoint
}

func (e EndpointMe) Connections() EndpointMeConnections {
	e2 := e.appendMajor("connections")
	return EndpointMeConnections{e2}
}

func (e EndpointMeConnections) Get(ctx context.Context) (connections []*ModelConnection, err error) {
	return connections, e.doMethod(ctx, "GET", nil, &connections)
}
