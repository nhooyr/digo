package discgo

import (
	"context"
	"encoding/json"
	"reflect"
)

type eventReady struct {
	V               int        `json:"v"`
	User            *User      `json:"user"`
	PrivateChannels []*Channel `json:"private_channels"`
	SessionID       string     `json:"session_id"`
	Trace           []string   `json:"_trace"`
}

type eventResumed struct {
	Trace []string `json:"_trace"`
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
	GuildID string   `json:"guild_id"`
	Roles   []string `json:"roles"`
	User    User     `json:"user"`
	Nick    string   `json:"nick"`
}

type EventGuildMembersChunk struct {
	GuildID string
	Members []*GuildMember
}

type EventGuildRoleCreate struct {
	GuildID string `json:"guild_id"`
	Role    Role   `json:"role"`
}

type EventGuildRoleUpdate struct {
	GuildID string `json:"guild_id"`
	Role    Role   `json:"role"`
}

type EventGuildRoleDelete struct {
	GuildID string `json:"guild_id"`
	Role    Role   `json:"role"`
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
	ChannelID string `json:"channel_id"`
}

type EventMessageDeleteBulk struct {
	IDs       []string `json:"ids"`
	ChannelID string   `json:"channel_id"`
}

type EventMessageReactionAdd struct {
	UserID    string     `json:"user_id"`
	ChannelID string     `json:"channel_id"`
	MessageID string     `json:"message_id"`
	Emoji     GuildEmoji `json:"emoji"`
}

type EventMessageReactionRemove struct {
	UserID    string      `json:"user_id"`
	ChannelID string      `json:"channel_id"`
	MessageID string      `json:"message_id"`
	Emoji     GuildMember `json:"emoji"`
}

type EventMessageReactionRemoveAll struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
}

type EventPresenceUpdate struct {
	User    User     `json:"user"`
	Roles   []string `json:"roles"`
	Game    *Game    `json:"game"`
	GuildID string   `json:"guild_id"`
	Status  string   `json:"status"`
}

type Game struct {
	Name string  `json:"name"`
	Type *int    `json:"type"`
	URL  *string `json:"url"`
}

const (
	// Yes this is actually what Discord calls it.
	GameTypeGame = iota
	GameTypeStreaming

	StatusIdle    = "idle"
	StatusDND     = "dnd"
	StatusOnline  = "online"
	StatusOffline = "offline"
)

type EventTypingStart struct {
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
	Timestamp int    `json:"timestamp"`
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
	case func(ctx context.Context, conn *Conn, e *eventResumed):
		em["RESUMED"] = fn
	case func(ctx context.Context, conn *Conn, e *EventChannelCreate):
		em["CHANNEL_CREATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventChannelUpdate):
		em["CHANNEL_UPDATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventChannelDelete):
		em["CHANNEL_DELETE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildCreate):
		em["GUILD_CREATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildUpdate):
		em["GUILD_UPDATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildDelete):
		em["GUILD_DELETE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildBanAdd):
		em["GUILD_BAN_ADD"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildBanRemove):
		em["GUILD_BAN_REMOVE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildEmojisUpdate):
		em["GUILD_EMOJIS_UPDATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildIntegrationsUpdate):
		em["GUILD_INTEGRATIONS_UPDATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildMemberAdd):
		em["GUILD_MEMBER_ADD"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildMemberRemove):
		em["GUILD_MEMBER_REMOVE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildMemberUpdate):
		em["GUILD_MEMBER_UPDATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildMembersChunk):
		em["GUILD_MEMBERS_CHUNK"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildRoleCreate):
		em["GUILD_ROLE_CREATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildRoleUpdate):
		em["GUILD_ROLE_UPDATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildRoleDelete):
		em["GUILD_ROLE_DELETE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventMessageCreate):
		em["MESSAGE_CREATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventMessageUpdate):
		em["MESSAGE_UPDATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventMessageDelete):
		em["MESSAGE_DELETE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventMessageDeleteBulk):
		em["MESSAGE_DELETE_BULK"] = fn
	case func(ctx context.Context, conn *Conn, e *EventMessageReactionAdd):
		em["MESSAGE_REACTION_ADD"] = fn
	case func(ctx context.Context, conn *Conn, e *EventMessageReactionRemove):
		em["MESSAGE_REACTION_REMOVE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventMessageReactionRemoveAll):
		em["MESSAGE_REACTION_REMOVE_ALL"] = fn
	case func(ctx context.Context, conn *Conn, e *EventPresenceUpdate):
		em["PRESENCE_UPDATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventTypingStart):
		em["TYPING_START"] = fn
	case func(ctx context.Context, conn *Conn, e *EventUserUpdate):
		em["USER_UPDATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventVoiceStateUpdate):
		em["VOICE_STATE_UPDATE"] = fn
	case func(ctx context.Context, conn *Conn, e *eventVoiceServerUpdate):
		em["VOICE_SERVER_UPDATE"] = fn
	default:
		panic("unknown event handler signature")
	}
}

func (em eventMux) route(ctx context.Context, conn *Conn, p *receivedPayload, sync bool) error {
	h, ok := em[p.Type]
	if !ok {
		// Discord better not be sending unknown events.
		return nil
	}
	e := reflect.New(reflect.TypeOf(h).In(2))
	err := json.Unmarshal(p.Data, e.Interface())
	if err != nil {
		return err
	}
	args := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(conn), e}
	fn := func() {
		reflect.ValueOf(h).Call(args)
	}
	if sync {
		fn()
	} else {
		go fn()
	}
	return nil
}
