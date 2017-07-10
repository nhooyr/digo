package discgo

import "github.com/bwmarrin/snowflake"

type User struct {
	ID            string
	Username      string
	Discriminator string
	Avatar        string
	Bot           bool
	MFAEnabled    bool
	Verified      bool
	Email         string
}

type UserGuild struct {
	ID          string
	Name        string
	Icon        string
	Owner       bool
	Permissions int
}

type Connection struct {
	ID           string
	Name         string
	Types        string
	Revoked      bool
	Integrations []*Integration
}
