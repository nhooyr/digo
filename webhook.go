package discgo

import "github.com/bwmarrin/snowflake"

type Webhook struct {
	ID        string
	GuildID   *string
	ChannelID string
	User      *User // TODO wtf why is question mark behind?
	Name      *string
	Avatar    *string
	Token     string
}
