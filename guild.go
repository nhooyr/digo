package discgo

import (
	"encoding/json"
	"net/url"
	"path"
	"strconv"
	"time"
)

type Guild struct {
	ID                          string
	Name                        string
	Icon                        string
	Splash                      string
	OwnerID                     string
	Region                      string
	AFKChannelID                string
	AFKTimeout                  int
	EmbedEnabled                bool
	EmbedChannelID              string
	VerificationLevel           int
	DefaultMessageNotifications int
	Roles                       []*Role
	Emojis                      []*GuildEmoji
	Features                    []string // not sure if this is right, DiscordGo doesn't have anything
	MFALevel                    int
	JoinedAt                    *time.Time
	Large                       bool
	Unavailable                 bool
	MemberCount                 int
	VoiceStates                 []*VoiceState // without guild_id key
	Members                     []*GuildMember
	Channels                    []*Channel
	Presences                   []*Presence // TODO like presence update event sans a roles or guild_id key
}

type Presence struct {
	User   *User
	Game   *Game
	Status string
}

type UnavailableGuild struct {
	ID          string
	Unavailable bool
}

type GuildEmbed struct {
	Enabled   bool
	ChannelID string
}

type GuildMember struct {
	User     *User
	Nick     *string
	Roles    []string
	JoinedAt time.Time
	Deaf     bool
	Mute     bool
}

type Integration struct {
	ID                string
	Name              string
	Type              string
	Enabled           bool
	Syncing           bool
	RoleID            string
	ExpireBehaviour   int
	ExpireGracePeriod int
	User              *User
	Account           *IntegrationAccount
	SyncedAt          time.Time
}

type IntegrationAccount struct {
	ID   string
	Name string
}

type GuildEmoji struct {
	ID            string
	Name          string
	Roles         []string
	RequireColons bool
	Managed       bool
}

type ParamsCreateGuild struct {
	Name                        string                 `json:"name,omitempty"`
	Region                      string                 `json:"region,omitempty"`
	Icon                        string                 `json:"icon,omitempty"`
	VerificationLevel           int                    `json:"verification_level,omitempty"`
	DefaultMessageNotifications int                    `json:"default_message_notifications,omitempty"`
	Roles                       []*Role                `json:"roles,omitempty"`
	Channels                    []*ParamsCreateChannel `json:"channels,omitempty"`
}

type ParamsCreateChannel struct {
	Name                 string       `json:"name"`
	Type                 string       `json:"type,omitempty"`
	Bitrate              int          `json:"bitrate,omitempty"`
	UserLimit            int          `json:"user_limit,omitempty"`
	PermissionOverwrites []*Overwrite `json:"permission_overwrites,omitempty"`
}

func (c *Client) CreateGuild(params ParamsCreateGuild) (g *Guild, err error) {
	endpoint := "guilds"
	req := c.newRequestJSON("POST", endpoint, params)
	body, err := c.do(req, endpoint, 0)
	if err != nil {
		return nil, err
	}
	return g, json.Unmarshal(body, &g)
}

func (c *Client) GetGuild(gID string) (g *Guild, err error) {
	endpoint := path.Join("guilds", gID)
	req := c.newRequest("GET", endpoint, nil)
	body, err := c.do(req, endpoint, 0)
	if err != nil {
		return nil, err
	}
	return g, json.Unmarshal(body, &g)
}

type ParamsModifyGuild struct {
	Name                        string `json:"name,omitempty"`
	Region                      string `json:"region,omitempty"`
	VerificationLevel           int    `json:"verification_level,omitempty"`
	DefaultMessageNotifications int    `json:"default_message_notifications,omitempty"`
	AFKChannelID                string `json:"afk_channel_id,omitempty"`
	AFKTiemout                  int    `json:"afk_tiemout,omitempty"`
	Icon                        string `json:"icon,omitempty"`
	OwnerID                     string `json:"owner_id,omitempty"`
	Splash                      string `json:"splash,omitempty"`
}

func (c *Client) ModifyGuild(gID string, params *ParamsModifyGuild) (g *Guild, err error) {
	endpoint := path.Join("guilds", gID)
	req := c.newRequestJSON("PATACH", endpoint, params)
	body, err := c.do(req, endpoint, 0)
	if err != nil {
		return nil, err
	}
	return g, json.Unmarshal(body, &g)
}

func (c *Client) DeleteGuild(gID string) (g *Guild, err error) {
	endpoint := path.Join("guilds", gID)
	req := c.newRequest("DELETE", endpoint, nil)
	body, err := c.do(req, endpoint, 0)
	if err != nil {
		return nil, err
	}
	return g, json.Unmarshal(body, &g)
}

func (c *Client) GetChannels(gID string) (channels []*Channel, err error) {
	endpoint := path.Join("guilds", gID, "channels")
	req := c.newRequest("GET", endpoint, nil)
	body, err := c.do(req, endpoint, 0)
	if err != nil {
		return nil, err
	}
	return channels, json.Unmarshal(body, &channels)
}

func (c *Client) CreateChannel(gID string, params *ParamsCreateChannel) (ch *Channel, err error) {
	endpoint := path.Join("guilds", gID, "channels")
	req := c.newRequestJSON("POST", endpoint, params)
	body, err := c.do(req, endpoint, 0)
	if err != nil {
		return nil, err
	}
	return ch, json.Unmarshal(body, &ch)
}

// TODO perhaps just use a channel struct?
// TODO gotta look at makign other names more clear/verbose
type ParamsModifyChannelPositions struct {
	ID       string `json:"id"`
	Position int    `json:"position"`
}

func (c *Client) ModifyChannelPositions(gID string, params *ParamsModifyChannelPositions) (channels *Channel, err error) {
	endpoint := path.Join("guilds", gID, "channels")
	req := c.newRequestJSON("PATCH", endpoint, params)
	body, err := c.do(req, endpoint, 0)
	if err != nil {
		return nil, err
	}
	return channels, json.Unmarshal(body, &channels)
}

func (c *Client) GetGuildMember(gID, uID string) (gm *GuildMember, err error) {
	endpoint := path.Join("guilds", gID, "members", uID)
	req := c.newRequest("GET", endpoint, nil)
	rateLimitPath := path.Join("guilds", gID, "members", "*")
	body, err := c.do(req, rateLimitPath, 0)
	if err != nil {
		return nil, err
	}
	return gm, json.Unmarshal(body, &gm)
}

type ParamsGetGuildMembers struct {
	Limit   int
	AfterID string
}

func (params *ParamsGetGuildMembers) rawQuery() string {
	v := make(url.Values)
	if params.AfterID != "" {
		v.Set("after", params.AfterID)
	}
	if params.Limit > 0 {
		v.Set("limit", strconv.Itoa(params.Limit))
	}
	return v.Encode()
}

func (c *Client) GetGuildMembers(gID string, params *ParamsGetGuildMembers) (guildMembers []*GuildMember, err error) {
	endpoint := path.Join("guilds", gID, "members")
	req := c.newRequest("GET", endpoint, nil)
	if params != nil {
		req.URL.RawQuery = params.rawQuery()
	}
	body, err := c.do(req, endpoint, 0)
	if err != nil {
		return nil, err
	}
	return guildMembers, json.Unmarshal(body, &guildMembers)
}

type ParamsAddGuildMember struct {
	AccessToken string  `json:"access_token"`
	Nick        string  `json:"nick,omitempty"`
	Roles       []*Role `json:"roles,omitempty"`
	Mute        bool    `json:"mute,omitempty"`
	Deaf        bool    `json:"deaf,omitempty"`
}

func (c *Client) AddGuildMember(gID, uID string, params *ParamsAddGuildMember) (gm *GuildMember, err error) {
	endpoint := path.Join("guilds", gID, "members", uID)
	req := c.newRequestJSON("PUT", endpoint, params)
	rateLimitPath := path.Join("guilds", gID, "members", "*")
	body, err := c.do(req, rateLimitPath, 0)
	if err != nil {
		return nil, err
	}
	return gm, json.Unmarshal(body, &gm)
}

// TODO rename this and all other params to postfix params
type ParamsModifyGuildMember struct {
	Nick      string  `json:"nick,omitempty"`
	Roles     []*Role `json:"roles,omitempty"`
	Mute      *bool   `json:"mute,omitempty"` // pointer so that you can set false
	Deaf      *bool   `json:"deaf,omitempty"` // pointer so that you can set false
	ChannelID string  `json:"channel_id,omitempty"`
}

func (c *Client) ModifyGuildMember(gID, uID string, params *ParamsModifyGuildMember) error {
	endpoint := path.Join("guilds", gID, "members", uID)
	req := c.newRequestJSON("PATCH", endpoint, params)
	rateLimitPath := path.Join("guilds", gID, "members", "*")
	_, err := c.do(req, rateLimitPath, 0)
	return err
}

func (c *Client) ModifyMyNick(gID string, nick string) error {
	endpoint := path.Join("guilds", gID, "members", "@me", "nick")
	params := map[string]string{"nick": nick}
	req := c.newRequestJSON("PATCH", endpoint, params)
	// Discord returns the nickname but I have no idea why that would
	// be of any use to anyone.
	// So lets just ignore it.
	_, err := c.do(req, endpoint, 0)
	return err
}
