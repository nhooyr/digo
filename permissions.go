package discgo

import "github.com/bwmarrin/snowflake"

type Role struct {
	ID          snowflake.ID
	Name        string
	Color       int
	Hoist       bool
	Position    int
	Permissions int
	Managed     bool
	Mentionable bool
}
