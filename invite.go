package discgo

import (
	"time"
)

type Invite struct {
	Code    string         `json:"code"`
	Guild   *InviteGuild   `json:"guild"`
	Channel *InviteChannel `json:"channel"`

	// Invite metadata
	Inviter   *User     `json:"inviter"`
	Uses      int       `json:"uses"`
	MaxUses   int       `json:"max_uses"`
	MaxAge    int       `json:"max_age"`
	Temporary bool      `json:"temporary"`
	CreatedAt time.Time `json:"created_at"`
	Revoked   bool      `json:"revoked"`
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
