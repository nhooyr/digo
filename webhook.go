package discgo

type Webhook struct {
	ID        string
	GuildID   *string
	ChannelID string
	User      *User
	Name      *string
	Avatar    *string
	Token     string
}
