package digo

import (
	"time"

	"github.com/bwmarrin/snowflake"
)


type Invite struct {
	Code    string
	Guild   *InviteGuild
	Channel *InviteChannel
}

type InviteMetadata struct {
	Inviter   *User
	Uses      int
	MaxUses   int
	MaxAge    int
	Temporary bool
	CreatedAt time.Time // TODO type not date time
	Revoked   bool
}

type InviteGuild struct {
	ID     snowflake.ID
	Name   string
	Splash *string // TODO nullable
	Icon   *string // TODO nullable
}

type InviteChannel struct {
	ID   snowflake.ID
	Name string
	Type string
}
