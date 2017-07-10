package discgo

import "github.com/bwmarrin/snowflake"

type VoiceState struct {
	GuildID   *snowflake.ID
	ChannelID snowflake.ID
	UserID    snowflake.ID
	SessionID string
	Deaf      bool
	Mute      bool
	SelfDeaf  bool
	SelfMute  bool
	Suppress  bool
}

type VoiceRegion struct {
	ID             string
	Name           string
	SampleHostname string
	SamplePort     int
	VIP            bool
	Optimal        bool
	Deprecated     bool
	Custom         bool
}
