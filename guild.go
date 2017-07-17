package discgo

import (
	"net/url"
	"path"
	"strconv"
	"time"
)

type Guild struct {
	ID                          string        `json:"id"`
	Name                        string        `json:"name"`
	Icon                        string        `json:"icon"`
	Splash                      string        `json:"splash"`
	OwnerID                     string        `json:"owner_id"`
	Region                      string        `json:"region"`
	AFKChannelID                string        `json:"afk_channel_id"`
	AFKTimeout                  int           `json:"afk_timeout"`
	EmbedEnabled                bool          `json:"embed_enabled"`
	EmbedChannelID              string        `json:"embed_channel_id"`
	VerificationLevel           int           `json:"verification_level"`
	DefaultMessageNotifications int           `json:"default_message_notifications"`
	Roles                       []*Role       `json:"roles"`
	Emojis                      []*GuildEmoji `json:"emojis"`
	Features                    []string      `json:"features"` // not sure if this is right, DiscordGo doesn't have anything
	MFALevel                    int           `json:"mfa_level"`
	JoinedAt                    time.Time     `json:"joined_at"`

	// These fields are only sent within the GUILD_CREATE event
	Large       *bool           `json:"large"`
	Unavailable *bool           `json:"unavailable"`
	MemberCount *int            `json:"member_count"`
	VoiceStates *[]*VoiceState  `json:"voice_states"` // without guild_id key
	Members     *[]*GuildMember `json:"members"`
	Channels    *[]*Channel     `json:"channels"`
	Presences   *[]*Presence    `json:"presences"` // TODO like presence update event sans a roles or guild_id key
}

// TOOD maybe Guild Presence rename?
type Presence struct {
	User   *User  `json:"user"`
	Game   *Game  `json:"game"`
	Status string `json:"status"`
}

type UnavailableGuild struct {
	ID          string `json:"id"`
	Unavailable bool `json:"unavailable"`
}

type GuildEmbed struct {
	Enabled   bool   `json:"enabled,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`
}

type GuildMember struct {
	User     *User `json:"user"`
	Nick     *string `json:"nick"`
	Roles    []string `json:"roles"`
	JoinedAt time.Time `json:"joined_at"`
	Deaf     bool `json:"deaf"`
	Mute     bool `json:"mute"`
}

type Integration struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Type              string `json:"type"`
	Enabled           bool `json:"enabled"`
	Syncing           bool `json:"syncing"`
	RoleID            string `json:"role_id"`
	ExpireBehaviour   int `json:"expire_behaviour"`
	ExpireGracePeriod int `json:"expire_grace_period"`
	User              *User `json:"user"`
	Account           *IntegrationAccount `json:"account"`
	SyncedAt          time.Time `json:"synced_at"`
}

type IntegrationAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GuildEmoji struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Roles         []string `json:"roles"`
	RequireColons bool `json:"require_colons"`
	Managed       bool `json:"managed"`
}

type GuildsEndpoint struct {
	*endpoint
}

func (c *Client) Guilds() *GuildsEndpoint {
	e2 := c.e.appendMajor("guilds")
	return &GuildsEndpoint{e2}
}

type GuildsCreateParams struct {
	Name                        string                 `json:"name,omitempty"`
	Region                      string                 `json:"region,omitempty"`
	Icon                        string                 `json:"icon,omitempty"`
	VerificationLevel           int                    `json:"verification_level,omitempty"`
	DefaultMessageNotifications int                    `json:"default_message_notifications,omitempty"`
	Roles                       []*Role                `json:"roles,omitempty"`
	Channels                    []*GuildChannelCreateParams `json:"channels,omitempty"`
}

// TODO not sure about the naming on this?
type GuildChannelCreateParams struct {
	Name                 string                 `json:"name"`
	Type                 string                 `json:"type,omitempty"`
	Bitrate              int                    `json:"bitrate,omitempty"`
	UserLimit            int                    `json:"user_limit,omitempty"`
	PermissionOverwrites []*PermissionOverwrite `json:"permission_overwrites,omitempty"`
}

// TODO Docs for this are not clear on what the Channels field should be, and the link for that field is broken.
func (e *GuildsEndpoint) Create(params *GuildsCreateParams) (g *Guild, err error) {
	return g, e.doMethod("POST", params, &g)
}

// TODO all endpoints need to be value receivers and returns!!!
type GuildEndpoint struct {
	*endpoint
}

func (c *Client) Guild(gID string) GuildEndpoint {
	e2 := c.Guilds().appendMajor(gID)
	return GuildEndpoint{e2}
}

func (e GuildEndpoint) Get(gID string) (g *Guild, err error) {
	return g, e.doMethod("GET", nil, &g)
}

type GuildModifyParams struct {
	Name                        string `json:"name,omitempty"`
	Region                      string `json:"region,omitempty"`
	VerificationLevel           int    `json:"verification_level,omitempty"`
	DefaultMessageNotifications int    `json:"default_message_notifications,omitempty"`
	AFKChannelID                string `json:"afk_channel_id,omitempty"`
	AFKTimeout                  int    `json:"afk_tiemout,omitempty"`
	Icon                        string `json:"icon,omitempty"`
	OwnerID                     string `json:"owner_id,omitempty"`
	Splash                      string `json:"splash,omitempty"`
}

func (e GuildEndpoint) Modify(params *GuildModifyParams) (g *Guild, err error) {
	return g, e.doMethod("PATCH", params, &g)
}

func (e GuildEndpoint) Delete() (g *Guild, err error) {
	return g, e.doMethod("DELETE", nil, &g)
}

// TODO not sure if necessary, there is a GuildMemberEndpoint.Modify, not sure if it takes @me. @abalabahaha#9421 on discord said that perhaps its different because modifying my nick and managing others is a different permission but that does not really make sense to me. But whatever...
func (e GuildEndpoint) ModifyMyNick(nick string) (newNick string, err error) {
	e2 := e.Member("@me").appendMajor("nick")
	nickStruct := struct {
		Nick string `json:"nick"`
	}{
		Nick: nick,
	}
	return nickStruct.Nick, e2.doMethod("PATCH", nickStruct, &nickStruct)
}

type GuildChannelsEndpoint struct {
	*endpoint
}

func (e GuildEndpoint) Channels() GuildChannelsEndpoint {
	e2 := e.appendMajor("channels")
	return GuildChannelsEndpoint{e2}
}

func (e *GuildChannelsEndpoint) Get(gID string) (channels []*Channel, err error) {
	return channels, e.doMethod("GET", nil, &channels)
}

func (e *GuildChannelsEndpoint) Create(params *GuildChannelCreateParams) (ch *Channel, err error) {
	return ch, e.doMethod("POST", params, &ch)
}

type GuildChannelsModifyPositionsParams struct {
	ID       string `json:"id"`
	Position int    `json:"position"`
}

func (e *GuildChannelsEndpoint) ModifyPositions(params []*GuildChannelsModifyPositionsParams) (channels *Channel, err error) {
	return channels, e.doMethod("PATCH", params, &channels)
}

type GuildMembersEndpoint struct {
	*endpoint
}

func (g GuildEndpoint) Members() GuildMembersEndpoint {
	e2 := g.appendMajor("members")
	return GuildMembersEndpoint{e2}
}

// TODO necessary???
type GuildMembersGetParams struct {
	Limit   int
	AfterID string
}

func (params *GuildMembersGetParams) rawQuery() string {
	v := make(url.Values)
	if params.AfterID != "" {
		v.Set("after", params.AfterID)
	}
	if params.Limit > 0 {
		v.Set("limit", strconv.Itoa(params.Limit))
	}
	return v.Encode()
}

func (e GuildMembersEndpoint) Get(params *GuildMembersGetParams) (guildMembers []*GuildMember, err error) {
	req := e.newRequest("GET", nil)
	if params != nil {
		req.URL.RawQuery = params.rawQuery()
	}
	return guildMembers, e.do(req, &guildMembers)
}

type GuildMemberEndpoint struct {
	*endpoint
}

func (e GuildEndpoint) Member(uID string) GuildMemberEndpoint {
	e2 := e.Members().appendMinor(uID)
	return GuildMemberEndpoint{e2}
}

type GuildMemberAddParams struct {
	AccessToken string  `json:"access_token"`
	Nick        string  `json:"nick,omitempty"`
	Roles       []*Role `json:"roles,omitempty"`
	Mute        bool    `json:"mute,omitempty"`
	Deaf        bool    `json:"deaf,omitempty"`
}

func (e GuildMemberEndpoint) Add(params *GuildMemberAddParams) (gm *GuildMember, err error) {
	return gm, e.doMethod("PUT", params, &gm)
}

func (e GuildMemberEndpoint) Get() (gm *GuildMember, err error) {
	return gm, e.doMethod("GET", nil, &gm)
}

// TODO rename this and all other params to postfix params
type GuildMemberModifyParams struct {
	Nick      string  `json:"nick,omitempty"`
	Roles     []*Role `json:"roles,omitempty"`
	Mute      *bool   `json:"mute,omitempty"` // pointer so that you can set false
	Deaf      *bool   `json:"deaf,omitempty"` // pointer so that you can set false
	ChannelID string  `json:"channel_id,omitempty"`
}

func (e *GuildMemberEndpoint) Modify(params *GuildMemberModifyParams) error {
	return e.doMethod("PATCH", params, nil)
}

type GuildMemberRoleEndpoint struct {
	*endpoint
}

func (e GuildMemberEndpoint) Role(roleID string) GuildMemberRoleEndpoint {
	e2 := e.appendMajor("roles").appendMinor(roleID)
	return GuildMemberRoleEndpoint{e2}
}

func (e GuildMemberRoleEndpoint) Add() error {
	return e.doMethod("PUT", nil, nil)
}

func (e GuildMemberRoleEndpoint) Remove() error {
	return e.doMethod("DELETE", nil, nil)
}

func (e GuildMemberEndpoint) Remove() error {
	return e.doMethod("DELETE", nil, nil)
}

type GuildBansEndpoint struct {
	*endpoint
}

func (e GuildEndpoint) Bans() GuildBansEndpoint {
	e2 := e.appendMajor("bans")
	return GuildBansEndpoint{e2}
}

func (e GuildBansEndpoint) Get() (users []*User, err error) {
	return users, e.doMethod("GET", nil, &users)
}

type GuildBanEndpoint struct {
	*endpoint
}

func (e GuildEndpoint) Ban(uID string) GuildBanEndpoint {
	e2 := e.Bans().appendMinor(uID)
	return GuildBanEndpoint{e2}
}

type GuildBanCreateParams struct {
	DeleteMessageDays int `json:"delete-message-days"`
}

func (e GuildBanEndpoint) Create(params *GuildBanCreateParams) error {
	return e.doMethod("PUT", params, nil)
}
func (e GuildBanEndpoint) Remove() error {
	return e.doMethod("DELETE", nil, nil)
}

func (c *Client) GetGuildRoles(gID string) (roles []*Role, err error) {
	endpoint := path.Join("guilds", gID, "roles")
	req := c.newRequest("GET", endpoint, nil)
	return roles, c.doUnmarshal(req, endpoint, &roles)
}

type CreateGuildRoleParams struct {
	Name        string `json:"name"`
	Permissions int    `json:"permissions"`
	Color       int    `json:"color"`
	Hoist       bool   `json:"hoist"`
	Mentionable bool   `json:"mentionable"`
}

func (c *Client) CreateGuildRole(gID string, params *CreateGuildRoleParams) (r *Role, err error) {
	endpoint := path.Join("guilds", gID, "roles")
	req := c.newRequestJSON("POST", endpoint, params)
	return r, c.doUnmarshal(req, endpoint, &r)
}

type ModifyGuildRolePositionsParams struct {
	ID       string `json:"id"`
	Position int    `json:"position"`
}

func (c *Client) ModifyGuildRolePositions(gID string, params *ModifyGuildRolePositionsParams) (roles []*Role, err error) {
	endpoint := path.Join("guilds", gID, "roles")
	req := c.newRequestJSON("PATCH", endpoint, params)
	return roles, c.doUnmarshal(req, endpoint, &roles)
}

type ModifyGuildRoleParams struct {
	Name        string `json:"name"`
	Permissions int    `json:"permissions"`
	Color       int    `json:"color"`
	Hoist       bool   `json:"hoist"`
	Mentionable bool   `json:"mentionable"`
}

func (c *Client) ModifyGuildRole(gID, roleID string, params *ModifyGuildRoleParams) (r *Role, err error) {
	endpoint := path.Join("guilds", gID, "roles", roleID)
	req := c.newRequestJSON("PATCH", endpoint, params)
	rateLimitPath := path.Join("guilds", gID, "roles", "*")
	return r, c.doUnmarshal(req, rateLimitPath, &r)
}

func (c *Client) DeleteGuildRole(gID, roleID string) error {
	endpoint := path.Join("guilds", gID, "roles", roleID)
	req := c.newRequest("DELETE", endpoint, nil)
	rateLimitPath := path.Join("guilds", gID, "roles", "*")
	return c.do(req, rateLimitPath)
}

func (c *Client) GetGuildPruneCount(gID string, days int) (pruned int, err error) {
	endpoint := path.Join("guilds", gID, "prune")
	req := c.newRequest("GET", endpoint, nil)
	if days > 0 {
		v := url.Values{}
		v.Set("days", strconv.Itoa(days))
		req.URL.RawQuery = v.Encode()
	}
	prunedStruct := struct {
		Pruned int `json:"pruned"`
	}{}
	return prunedStruct.Pruned, c.doUnmarshal(req, endpoint, &prunedStruct)
}

func (c *Client) BeginGuildPrune(gID string, days int) (pruned int, err error) {
	endpoint := path.Join("guilds", gID, "prune")
	req := c.newRequest("POST", endpoint, nil)
	if days > 0 {
		v := url.Values{}
		v.Set("days", strconv.Itoa(days))
		req.URL.RawQuery = v.Encode()
	}
	prunedStruct := struct {
		Pruned int `json:"pruned"`
	}{}
	return prunedStruct.Pruned, c.doUnmarshal(req, endpoint, &prunedStruct)
}

func (c *Client) Name(gID string) (voiceRegions []*VoiceRegion, err error) {
	endpoint := path.Join("guilds", gID, "regions")
	req := c.newRequest("GET", endpoint, nil)
	return voiceRegions, c.doUnmarshal(req, endpoint, &voiceRegions)
}

func (c *Client) GetGuildInvites(gID string) (invites []*Invite, err error) {
	endpoint := path.Join("guilds", gID, "invites")
	req := c.newRequest("GET", endpoint, nil)
	return invites, c.doUnmarshal(req, endpoint, &invites)
}

func (c *Client) GetGuildIntegrations(gID string) (integrations []*Integration, err error) {
	endpoint := path.Join("guilds", gID, "integrations")
	req := c.newRequest("GET", endpoint, nil)
	return integrations, c.doUnmarshal(req, endpoint, &integrations)
}

type CreateGuildIntegrationParams struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func (c *Client) CreateGuildIntegration(gID string, params *CreateGuildIntegrationParams) error {
	endpoint := path.Join("guilds", gID, "integrations")
	req := c.newRequestJSON("POST", endpoint, params)
	return c.do(req, endpoint)
}

type ModifyGuildIntegrationParams struct {
	// TODO impossible to not send or send 0 value :(
	ExpireBehaviour   int  `json:"expire_behaviour,omitempty"`
	ExpireGracePeriod int  `json:"expire_grace_period,omitempty"`
	EnableEmoticons   bool `json:"enable_emoticons,omitempty"`
}

func (c *Client) ModifyGuildIntegration(gID, integrationID string, params *ModifyGuildIntegrationParams) error {
	endpoint := path.Join("guilds", gID, "integrations", integrationID)
	req := c.newRequestJSON("PATCH", endpoint, params)
	rateLimitPath := path.Join("guilds", gID, "integrations", "*")
	return c.do(req, rateLimitPath)
}

func (c *Client) DeleteGuildIntegration(gID, integrationID string) error {
	endpoint := path.Join("guilds", gID, "integrations", integrationID)
	req := c.newRequest("DELETE", endpoint, nil)
	rateLimitPath := path.Join("guilds", gID, "integrations", "*")
	return c.do(req, rateLimitPath)
}

func (c *Client) SyncGuildIntegration(gID, integrationID string) error {
	endpoint := path.Join("guilds", gID, "integrations", integrationID, "sync")
	req := c.newRequest("POST", endpoint, nil)
	rateLimitPath := path.Join("guilds", gID, "integrations", "*", "sync")
	return c.do(req, rateLimitPath)
}

func (c *Client) GetGuildEmbed(gID string) (ge *GuildEmbed, err error) {
	endpoint := path.Join("guilds", gID, "embed")
	req := c.newRequest("GET", endpoint, nil)
	return ge, c.doUnmarshal(req, endpoint, &ge)
}

func (c *Client) ModifyGuildEmbed(gID string, ge *GuildEmbed) (newGE *GuildEmbed, err error) {
	endpoint := path.Join("guilds", gID, "embed")
	req := c.newRequestJSON("PATCH", endpoint, ge)
	return newGE, c.doUnmarshal(req, endpoint, &newGE)
}
