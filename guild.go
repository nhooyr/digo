package discgo

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

type Guild struct {
	ID                          snowflake.ID
	Name                        string
	Icon                        string
	Splash                      string
	OwnerID                     snowflake.ID
	Region                      string
	AFKChannelID                snowflake.ID
	AFKTimeout                  int
	EmbedEnabled                bool
	EmbedChannelID              snowflake.ID
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
	// Presences                   []*Presence // TODO like presence update event sans a roles or guild_id key
}

type UnavailableGuild struct {
	ID          snowflake.ID
	Unavailable bool
}

type GuildEmbed struct {
	Enabled   bool
	ChannelID snowflake.ID
}

type GuildMember struct {
	User     *User
	Nick     *string
	Roles    []snowflake.ID
	JoinedAt time.Time
	Deaf     bool
	Mute     bool
}

type Integration struct {
	ID                snowflake.ID
	Name              string
	Type              string
	Enabled           bool
	Syncing           bool
	RoleID            snowflake.ID
	ExpireBehaviour   int
	ExpireGracePeriod int
	User              *User
	Account           *IntegrationAccount
	SyncedAt          time.Time
}

type IntegrationAccount struct {
	ID   snowflake.ID
	Name string
}

type GuildEmoji struct {
	ID            snowflake.ID
	Name          string
	Roles         []snowflake.ID
	RequireColons bool
	Managed       bool
}
