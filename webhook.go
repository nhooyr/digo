package discgo

// TODO not needed for now, but maybe later?
type ModelWebhook struct {
	ID        string     `json:"id"`
	GuildID   *string    `json:"guild_id"`
	ChannelID string     `json:"channel_id"`
	User      *ModelUser `json:"user"`
	Name      *string    `json:"name"`
	Avatar    *string    `json:"avatar"`
	Token     string     `json:"token"`
}
