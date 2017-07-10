package discgo

import (
	"time"
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
	CreatedAt time.Time
	Revoked   bool
}

type InviteGuild struct {
	ID     string
	Name   string
	Splash *string
	Icon   *string
}

type InviteChannel struct {
	ID   string
	Name string
	Type string
}
