package discgo

import (
	"net/url"
	"path"
	"strconv"
)

type User struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
	Bot           bool   `json:"bot"`
	MFAEnabled    bool   `json:"mfa_enabled"`
	Verified      bool   `json:"verified"`
	Email         string `json:"email"`
}

type UserGuild struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Owner       bool   `json:"owner"`
	Permissions int    `json:"permissions"`
}

type Connection struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Types        string         `json:"types"`
	Revoked      bool           `json:"revoked"`
	Integrations []*Integration `json:"integrations"`
}

type UserEndpoint struct {
	*endpoint
}

func (c *Client) User(uID string) UserEndpoint {
	e2 := c.e.appendMajor("users").appendMinor(uID)
	return UserEndpoint{e2}
}

func (e UserEndpoint) Get() (u *User, err error) {
	return u, e.doMethod("GET", nil, &u)
}

type MeEndpoint struct {
	*endpoint
}

func (c *Client) Me() MeEndpoint {
	e2 := c.User("@me").endpoint
	return MeEndpoint{e2}
}

func (e MeEndpoint) Get() (u *User, err error) {
	return u, e.doMethod("GET", nil, &u)
}

type MeModifyParams struct {
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

func (e MeEndpoint) Modify(params *MeModifyParams) (u *User, err error) {
	return u, e.doMethod("PATCH", params, &u)
}

type MeGuildsEndpoint struct {
	*endpoint
}

func (e MeEndpoint) Guilds() MeGuildsEndpoint {
	e2 := e.appendMajor("guilds")
	return MeGuildsEndpoint{e2}
}

type MeGuildsGetParams struct {
	BeforeID string
	AfterID  string
	Limit    int
}

func (params *MeGuildsGetParams) rawQuery() string {
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

func (e MeGuildsEndpoint) Get(params *MeGuildsGetParams) (guilds []*UserGuild, err error) {
	req := e.newRequest("GET", nil)
	if params != nil {
		req.URL.RawQuery = params.rawQuery()
	}
	return guilds, e.do(req, &guilds)
}

type MeGuildEndpoint struct {
	*endpoint
}

func (e MeEndpoint) Guild(gID string) MeGuildEndpoint {
	e2 := e.Guilds().appendMajor(gID)
	return MeGuildEndpoint{e2}
}

func (e MeGuildEndpoint) Leave() error {
	return e.doMethod("DELETE", nil, nil)
}

type MeDMChannelsEndpoint struct {
	*endpoint
}

func (e MeEndpoint) DMChannels() MeDMChannelsEndpoint {
	e2 := e.appendMajor("channels")
	return MeDMChannelsEndpoint{e2}
}

func (e MeDMChannelsEndpoint) Get() (dmChannels *[]Channel, err error) {
	return dmChannels, e.doMethod("GET", nil, &dmChannels)
}

type DMChannelsCreateParams struct {
	RecipientID string `json:"recipient_id"`
}

func (e MeDMChannelsEndpoint) Create(params *DMChannelsCreateParams) (ch *Channel, err error) {
	return ch, e.doMethod("POST", params, &ch)
}

type DmChannelsCreateGroupParams struct {
	AccessTokens []string          `json:"access_tokens"`
	Nicks        map[string]string `json:"nicks"`
}

func (e MeDMChannelsEndpoint) CreateGroup(params *DmChannelsCreateGroupParams) (ch *Channel, err error) {
	return ch, e.doMethod("POST", params, &ch)
}

type MeConnectionsEndpoint struct {
	*endpoint
}

func (e MeEndpoint) Connections() MeConnectionsEndpoint {
	e2 := e.appendMajor("connections")
	return MeConnectionsEndpoint{e2}
}

func (e MeConnectionsEndpoint) Get() (connections []*Connection, err error) {
	return connections, e.doMethod("GET", nil, &connections)
}
