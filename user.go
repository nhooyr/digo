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

// uID = "@me" for current user
func (c *Client) GetUser(uID string) (u *User, err error) {
	endpoint := path.Join("users", uID)
	req := c.newRequest("GET", endpoint, nil)
	rateLimitPath := path.Join("users", "*")
	return u, c.doUnmarshal(req, rateLimitPath, &u)
}

type ModifyMeParams struct {
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

func (c *Client) ModifyMe(params *ModifyMeParams) (u *User, err error) {
	endpoint := path.Join("users", "@me")
	req := c.newRequestJSON("POST", endpoint, params)
	rateLimitPath := path.Join("users", "*")
	return u, c.doUnmarshal(req, rateLimitPath, &u)
}

type GetMyGuildsParams struct {
	BeforeID string
	AfterID  string
	Limit    int
}

func (params *GetMyGuildsParams) rawQuery() string {
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

func (c *Client) GetMyGuilds(params *GetMyGuildsParams) (guilds []*UserGuild, err error) {
	endpoint := path.Join("users", "@me", "guilds")
	req := c.newRequest("GET", endpoint, nil)
	if params != nil {
		req.URL.RawQuery = params.rawQuery()
	}
	return guilds, c.doUnmarshal(req, endpoint, &guilds)
}

func (c *Client) LeaveGuild(gID string) error {
	endpoint := path.Join("users", "@me", "guilds", gID)
	req := c.newRequest("DELETE", endpoint, nil)
	return c.do(req, endpoint)
}

func (c *Client) GetMyDMs() (dmChannels *[]Channel, err error) {
	endpoint := path.Join("users", "@me", "channels")
	req := c.newRequest("GET", endpoint, nil)
	return dmChannels, c.doUnmarshal(req, endpoint, &dmChannels)
}

type CreateDMParams struct {
	RecipientID string `json:"recipient_id"`
}

func (c *Client) CreateDM(params *CreateDMParams) (ch *Channel, err error) {
	endpoint := path.Join("users", "@me", "channels")
	req := c.newRequestJSON("POST", endpoint, params)
	return ch, c.doUnmarshal(req, endpoint, &ch)
}

type CreateGroupDMParams struct {
	AccessTokens []string          `json:"access_tokens"`
	Nicks        map[string]string `json:"nicks"`
}

func (c *Client) CreateGroupDM(params *CreateGroupDMParams) (ch *Channel, err error) {
	endpoint := path.Join("users", "@me", "channels")
	req := c.newRequestJSON("POST", endpoint, params)
	return ch, c.doUnmarshal(req, endpoint, &ch)
}

func (c *Client) GetMyConnections() (connections []*Connection, err error) {
	endpoint := path.Join("users", "@me", "connections")
	req := c.newRequest("GET", endpoint, nil)
	return connections, c.doUnmarshal(req, endpoint, &connections)
}
