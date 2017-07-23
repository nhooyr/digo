package discgo

import (
	"context"
	"encoding/json"
)

type eventReady struct {
	V               int        `json:"v"`
	User            *User      `json:"user"`
	PrivateChannels []*Channel `json:"private_channels"`
	SessionID       string     `json:"session_id"`
	Trace           []string   `json:"_trace"`
}

type EventChannelCreate struct {
	Channel `json:"-"`
}

type EventChannelUpdate struct {
	Channel `json:"-"`
}

type EventChannelDelete struct {
	Channel `json:"-"`
}

type EventGuildCreate struct {
	Guild `json:"-"`
}

type EventGuildUpdate struct {
	Guild `json:"-"`
}

type EventGuildDelete struct {
	ID          string `json:"id"`
	Unavailable bool   `json:"unavailable"`
}

type EventGuildBanAdd struct {
	User    `json:"-"`
	GuildID string `json:"guild_id"`
}

type EventGuildBanRemove struct {
	User    `json:"-"`
	GuildID string `json:"guild_id"`
}

type EventGuildEmojisUpdate struct {
	GuildID string        `json:"guild_id"`
	Emojis  []*GuildEmoji `json:"emojis"`
}

type EventGuildIntegrationsUpdate struct {
	GuildID string `json:"guild_id"`
}

type EventGuildMemberAdd struct {
	GuildID string `json:"guild_id"`
}

type EventGuildMemberRemove struct {
	User    *User  `json:"user"`
	GuildID string `json:"guild_id"`
}

type EventGuildMemberUpdate struct {
	GuildID string `json:"guild_id"`
	Roles   []string `json:"roles"`
	User    User `json:"user"`
	Nick    string `json:"nick"`
}

type EventGuildMembersChunk struct {
	GuildID string
	Members []*GuildMember
}

type EventGuildRoleCreate struct {
	GuildID string `json:"guild_id"`
	Role    Role `json:"role"`
}

type EventGuildRoleUpdate struct {
	GuildID string `json:"guild_id"`
	Role    Role `json:"role"`
}

type EventGuildRoleDelete struct {
	GuildID string `json:"guild_id"`
	Role    Role `json:"role"`
}

type EventMessageCreate struct {
	Message `json:"-"`
}

// May not be full message.
type EventMessageUpdate struct {
	Message `json:"-"`
}

type EventMessageDelete struct {
	ID        string `json:"id"`
	ChannelID string  `json:"channel_id"`
}

type EventMessageDeleteBulk struct {
	IDs       []string `json:"ids"`
	ChannelID string `json:"channel_id"`
}

type EventMessageReactionAdd struct {
	UserID    string `json:"user_id"`
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
	Emoji     GuildEmoji `json:"emoji"`
}

type EventMessageReactionRemove struct {
	UserID    string `json:"user_id"`
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
	Emoji     GuildMember `json:"emoji"`
}

type EventMessageReactionRemoveAll struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
}

type EventPresenceUpdate struct {
	User    User `json:"user"`
	Roles   []string `json:"roles"`
	Game    *Game `json:"game"`
	GuildID string `json:"guild_id"`
	Status  string `json:"status"`
}

type Game struct {
	Name string  `json:"name"`
	Type *int    `json:"type"`
	URL  *string `json:"url"`
}

const (
	// Yes this is actually what Discord calls it.
	GameTypeGame      = iota
	GameTypeStreaming

	StatusIdle    = "idle"
	StatusDND     = "dnd"
	StatusOnline  = "online"
	StatusOffline = "offline"
)

type EventTypingStart struct {
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
	Timestamp int `json:"timestamp"`
}

type EventUserUpdate struct {
	User `json:"-"`
}

type EventVoiceStateUpdate struct {
	VoiceState `json:"-"`
}

type eventVoiceServerUpdate struct {
	Token    string `json:"token"`
	GuildID  string `json:"guild_id"`
	Endpoint string `json:"endpoint"`
}

type eventMux map[string]interface{}

func (em eventMux) register(fn interface{}) {
	switch fn.(type) {
	case func(ctx context.Context, conn *Conn, e *eventReady):
		em["READY"] = fn
	}
}

func (em eventMux) route(ctx context.Context, conn *Conn, p *receivedPayload, sync bool) error {
	var fn func()
	switch p.Type {
	case "READY":
		{
			h, ok := em[p.Type]
			if ok {
				var e eventReady
				err := json.Unmarshal(p.Data, &e)
				if err != nil {
					return err
				}
				fn = func() {
					h.(func(context.Context, *Conn, *eventReady))(ctx, conn, &e)
				}
			}
		}
	}
	if sync {
		fn()
	} else {
		go fn()
	}
	return nil
}
