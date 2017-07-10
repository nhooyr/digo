package discgo

import "github.com/bwmarrin/snowflake"

type VoiceState struct {
	GuildID   *string
	ChannelID string
	UserID    string
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
