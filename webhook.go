package digo

import "github.com/bwmarrin/snowflake"

type Webhook struct {
	ID        snowflake.ID
	GuildID   *snowflake.ID
	ChannelID snowflake.ID
	User      *User // TODO wtf why is question mark behind?
	Name      *string
	Avatar    *string
	Token     string
}