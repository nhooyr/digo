package discgo

import (
	"net/url"
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
	Unavailable bool   `json:"unavailable"`
}

type GuildEmbed struct {
	Enabled   bool   `json:"enabled,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`
}

type GuildMember struct {
	User     *User     `json:"user"`
	Nick     *string   `json:"nick"`
	Roles    []string  `json:"roles"`
	JoinedAt time.Time `json:"joined_at"`
	Deaf     bool      `json:"deaf"`
	Mute     bool      `json:"mute"`
}

type Integration struct {
	ID                string              `json:"id"`
	Name              string              `json:"name"`
	Type              string              `json:"type"`
	Enabled           bool                `json:"enabled"`
	Syncing           bool                `json:"syncing"`
	RoleID            string              `json:"role_id"`
	ExpireBehaviour   int                 `json:"expire_behaviour"`
	ExpireGracePeriod int                 `json:"expire_grace_period"`
	User              *User               `json:"user"`
	Account           *IntegrationAccount `json:"account"`
	SyncedAt          time.Time           `json:"synced_at"`
}

type IntegrationAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GuildEmoji struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Roles         []string `json:"roles"`
	RequireColons bool     `json:"require_colons"`
	Managed       bool     `json:"managed"`
}

type GuildsEndpoint struct {
	*endpoint
}

func (c *Client) Guilds() GuildsEndpoint {
	e2 := c.e.appendMajor("guilds")
	return GuildsEndpoint{e2}
}

type GuildsCreateParams struct {
	Name                        string                      `json:"name,omitempty"`
	Region                      string                      `json:"region,omitempty"`
	Icon                        string                      `json:"icon,omitempty"`
	VerificationLevel           int                         `json:"verification_level,omitempty"`
	DefaultMessageNotifications int                         `json:"default_message_notifications,omitempty"`
	Roles                       []*Role                     `json:"roles,omitempty"`
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
func (e GuildsEndpoint) Create(params *GuildsCreateParams) (g *Guild, err error) {
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
// TODO i also don't really like the name, doesn't fit in
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

func (e GuildChannelsEndpoint) Get() (channels []*Channel, err error) {
	return channels, e.doMethod("GET", nil, &channels)
}

func (e GuildChannelsEndpoint) Create(params *GuildChannelCreateParams) (ch *Channel, err error) {
	return ch, e.doMethod("POST", params, &ch)
}

type GuildChannelsModifyPositionsParams struct {
	ID       string `json:"id"`
	Position int    `json:"position"`
}

func (e GuildChannelsEndpoint) ModifyPositions(params []*GuildChannelsModifyPositionsParams) (channels *Channel, err error) {
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

func (e GuildMemberEndpoint) Modify(params *GuildMemberModifyParams) error {
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

// TODO maybe guildrole instead?
type RolesEndpoint struct {
	*endpoint
}

func (e GuildEndpoint) Roles() RolesEndpoint {
	e2 := e.appendMajor("roles")
	return RolesEndpoint{e2}
}

func (e RolesEndpoint) Get() (roles []*Role, err error) {
	return roles, e.doMethod("GET", nil, &roles)
}

type RoleCreateParams struct {
	Name string `json:"name,omitempty"`
	// TODO should be null?
	Permissions int  `json:"permissions,omitempty"`
	Color       int  `json:"color,omitempty"`
	Hoist       bool `json:"hoist,omitempty"`
	Mentionable bool `json:"mentionable,omitempty"`
}

func (e RolesEndpoint) Create(params *RoleCreateParams) (r *Role, err error) {
	return r, e.doMethod("GET", params, &r)
}

type RolesModifyPositionsParams struct {
	ID       string `json:"id"`
	Position int    `json:"position"`
}

func (e RolesEndpoint) ModifyPositions(params *RolesModifyPositionsParams) (roles []*Role, err error) {
	return roles, e.doMethod("PATCH", params, &roles)
}

type RoleEndpoint struct {
	*endpoint
}

func (e GuildEndpoint) Role(roleID string) RoleEndpoint {
	e2 := e.Roles().appendMinor(roleID)
	return RoleEndpoint{e2}
}

// TODO nulls
type RoleModifyParams struct {
	Name        string `json:"name,omitempty"`
	Permissions int    `json:"permissions,omitempty"`
	Color       int    `json:"color,omitempty"`
	Hoist       bool   `json:"hoist,omitempty"`
	Mentionable bool   `json:"mentionable,omitempty"`
}

func (e RoleEndpoint) Modify(params *RoleModifyParams) (r *Role, err error) {
	return r, e.doMethod("PATCH", params, &r)
}

func (e RoleEndpoint) Delete() error {
	return e.doMethod("DELETE", nil, nil)
}

// TODO i don't like the api because prune is a verb :(. same as bulk-delete ****
type PruneEndpoint struct {
	*endpoint
}

func (e GuildEndpoint) Prune() PruneEndpoint {
	e2 := e.appendMajor("prune")
	return PruneEndpoint{e2}
}

func (e PruneEndpoint) GetCount(days int) (count int, err error) {
	req := e.newRequest("GET", nil)
	if days > 0 {
		v := url.Values{}
		v.Set("days", strconv.Itoa(days))
		req.URL.RawQuery = v.Encode()
	}
	countStruct := struct {
		// TODO should I stick with discord's naming?
		Count int `json:"pruned"`
	}{}
	return countStruct.Count, e.do(req, &countStruct)
}

func (e PruneEndpoint) Begin(days int) (pruned int, err error) {
	req := e.newRequest("POST", nil)
	if days > 0 {
		v := url.Values{}
		v.Set("days", strconv.Itoa(days))
		req.URL.RawQuery = v.Encode()
	}
	prunedStruct := struct {
		Pruned int `json:"pruned"`
	}{}
	return prunedStruct.Pruned, e.do(req, &prunedStruct)
}

func (e GuildEndpoint) VoiceRegions() VoiceRegionsEndpoint {
	e2 := e.appendMajor("regions")
	return VoiceRegionsEndpoint{e2}
}

func (e GuildEndpoint) Invites() InvitesEndpoint {
	e2 := e.appendMajor("invites")
	return InvitesEndpoint{e2}
}

type IntegrationsEndpoint struct {
	*endpoint
}

func (e GuildEndpoint) Integrations() IntegrationsEndpoint {
	e2 := e.appendMajor("integrations")
	return IntegrationsEndpoint{e2}
}

func (e IntegrationsEndpoint) Get() (integrations []*Integration, err error) {
	return integrations, e.doMethod("GET", nil, &integrations)
}

type IntegrationsCreateParams struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func (e IntegrationsEndpoint) Create(params *IntegrationsCreateParams) error {
	return e.doMethod("POST", params, nil)
}

type IntegrationEndpoint struct {
	*endpoint
}

func (e GuildEndpoint) Integration(integrationID string) IntegrationEndpoint {
	e2 := e.Integrations().appendMinor(integrationID)
	return IntegrationEndpoint{e2}
}

type IntegrationModifyParams struct {
	// TODO impossible to not send or send 0 value :(
	ExpireBehaviour   int  `json:"expire_behaviour,omitempty"`
	ExpireGracePeriod int  `json:"expire_grace_period,omitempty"`
	EnableEmoticons   bool `json:"enable_emoticons,omitempty"`
}

func (e IntegrationEndpoint) Modify(params *IntegrationModifyParams) error {
	return e.doMethod("PATCH", params, nil)
}

func (e IntegrationEndpoint) Delete(gID, integrationID string) error {
	return e.doMethod("DELETE", nil, nil)
}

func (e IntegrationEndpoint) Sync() error {
	e2 := e.appendMajor("sync")
	return e2.doMethod("POST", nil, nil)
}

type GuildEmbedEndpoint struct {
	*endpoint
}

func (e GuildEndpoint) Embed() GuildEmbedEndpoint {
	e2 := e.appendMajor("embed")
	return GuildEmbedEndpoint{e2}
}

func (e GuildEmbedEndpoint) Get() (ge *GuildEmbed, err error) {
	return ge, e.doMethod("GET", nil, &ge)
}

func (e GuildEmbedEndpoint) Modify(ge *GuildEmbed) (newGE *GuildEmbed, err error) {
	return newGE, e.doMethod("PATCH", ge, &newGE)
}
