package discgo

// TODO not needed for now, but maybe later?
type Webhook struct {
	ID        string  `json:"id"`
	GuildID   *string `json:"guild_id"`
	ChannelID string  `json:"channel_id"`
	User      *User   `json:"user"`
	Name      *string `json:"name"`
	Avatar    *string `json:"avatar"`
	Token     string  `json:"token"`
}
