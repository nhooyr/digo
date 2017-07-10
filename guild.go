package discgo

import (
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
	// Presences                   []*Presence // TODO like presence update event sans a roles or guild_id key
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
