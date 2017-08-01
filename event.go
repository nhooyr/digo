package discgo

import (
	"encoding/json"
	"fmt"
)

type eventReady struct {
	V               int                 `json:"v"`
	User            *ModelUser          `json:"user"`
	PrivateChannels []*ModelChannel     `json:"private_channels"`
	Guilds          []*EventGuildCreate `json:"guilds"`
	SessionID       string              `json:"session_id"`
	Trace           []string            `json:"_trace"`
}

type eventResumed struct {
	Trace []string `json:"_trace"`
}

type EventChannelCreate struct {
	ModelChannel
}

type EventChannelUpdate struct {
	ModelChannel
}

type EventChannelDelete struct {
	ModelChannel
}

type EventGuildCreate struct {
	ModelGuild
	Large       bool                `json:"large"`
	Unavailable bool                `json:"unavailable"`
	MemberCount int                 `json:"member_count"`
	VoiceStates []*ModelVoiceState  `json:"voice_states"` // without guild_id key
	Members     []*ModelGuildMember `json:"members"`
	Channels    []*ModelChannel     `json:"channels"`
	Presences   []*ModelPresence    `json:"presences"`
}

type ModelPresence struct {
	User   ModelUser  `json:"user"`
	Game   *ModelGame `json:"game"`
	Status string     `json:"status"`
}

type EventGuildUpdate struct {
	ModelGuild
}

type EventGuildDelete struct {
	ID          string `json:"id"`
	Unavailable bool   `json:"unavailable"`
}

type EventGuildBanAdd struct {
	ModelUser
	GuildID string `json:"guild_id"`
}

type EventGuildBanRemove struct {
	ModelUser
	GuildID string `json:"guild_id"`
}

type EventGuildEmojisUpdate struct {
	GuildID string             `json:"guild_id"`
	Emojis  []*ModelGuildEmoji `json:"emojis"`
}

type EventGuildIntegrationsUpdate struct {
	GuildID string `json:"guild_id"`
}

type EventGuildMemberAdd struct {
	ModelGuildMember
	GuildID string `json:"guild_id"`
}

type EventGuildMemberRemove struct {
	User    ModelUser `json:"user"`
	GuildID string    `json:"guild_id"`
}

type EventGuildMemberUpdate struct {
	GuildID string    `json:"guild_id"`
	Roles   []string  `json:"roles"`
	User    ModelUser `json:"user"`
	Nick    string    `json:"nick"`
}

type EventGuildMembersChunk struct {
	GuildID string              `json:"guild_id"`
	Members []*ModelGuildMember `json:"members"`
}

type EventGuildRoleCreate struct {
	GuildID string    `json:"guild_id"`
	Role    ModelRole `json:"role"`
}

type EventGuildRoleUpdate struct {
	GuildID string    `json:"guild_id"`
	Role    ModelRole `json:"role"`
}

type EventGuildRoleDelete struct {
	GuildID string    `json:"guild_id"`
	Role    ModelRole `json:"role"`
}

type EventMessageCreate struct {
	ModelMessage
}

// May not be full message.
type EventMessageUpdate struct {
	ModelMessage
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
	UserID    string          `json:"user_id"`
	ChannelID string          `json:"channel_id"`
	MessageID string          `json:"message_id"`
	Emoji     ModelGuildEmoji `json:"emoji"`
}

type EventMessageReactionRemove struct {
	UserID    string           `json:"user_id"`
	ChannelID string           `json:"channel_id"`
	MessageID string           `json:"message_id"`
	Emoji     ModelGuildMember `json:"emoji"`
}

type EventMessageReactionRemoveAll struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
}

type EventPresenceUpdate struct {
	// TODO why is there even a user here?
	User    ModelUser  `json:"user"`
	Roles   []string   `json:"roles"`
	Game    *ModelGame `json:"game"`
	GuildID string     `json:"guild_id"`
	Status  string     `json:"status"`
}

const (
	StatusIdle    = "idle"
	StatusDND     = "dnd"
	StatusOnline  = "online"
	StatusOffline = "offline"
)

type ModelGame struct {
	Name string  `json:"name"`
	Type *int    `json:"type"`
	URL  *string `json:"url"`
}

const (
	// Yes this is actually what Discord calls it.
	ModelGameTypeGame = iota
	ModelGameTypeStreaming
)

type EventTypingStart struct {
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
	Timestamp int    `json:"timestamp"`
}

type EventUserUpdate struct {
	ModelUser
}

type EventVoiceStateUpdate struct {
	ModelVoiceState
}

type eventVoiceServerUpdate struct {
	Token    string `json:"token"`
	GuildID  string `json:"guild_id"`
	Endpoint string `json:"endpoint"`
}

type EventHandlerError struct {
	Err       error
	Event     interface{}
	EventName string
}

func (e *EventHandlerError) Error() string {
	eventJSON, err := json.MarshalIndent(e.Event, "", "    ")
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%v handler error: %v\nevent: %v", e.EventName, e.Err, eventJSON)
}

func getEventStruct(eventType string) interface{} {
	switch eventType {
	case "READY":
		return new(eventReady)
	case "RESUMED":
		return new(eventResumed)
	case "CHANNEL_CREATE":
		return new(EventChannelCreate)
	case "CHANNEL_UPDATE":
		return new(*EventChannelUpdate)
	case "CHANNEL_DELETE":
		return new(*EventChannelDelete)
	case "GUILD_CREATE":
		return new(*EventGuildCreate)
	case "GUILD_UPDATE":
		return new(*EventGuildUpdate)
	case "GUILD_DELETE":
		return new(*EventGuildDelete)
	case "GUILD_BAN_ADD":
		return new(*EventGuildBanAdd)
	case "GUILD_BAN_REMOVE":
		return new(*EventGuildBanRemove)
	case "GUILD_EMOJIS_UPDATE":
		return new(*EventGuildEmojisUpdate)
	case "GUILD_INTEGRATIONS_UPDATE":
		return new(*EventGuildIntegrationsUpdate)
	case "GUILD_MEMBER_ADD":
		return new(*EventGuildMemberAdd)
	case "GUILD_MEMBER_REMOVE":
		return new(*EventGuildMemberRemove)
	case "GUILD_MEMBER_UPDATE":
		return new(*EventGuildMemberUpdate)
	case "GUILD_MEMBERS_CHUNK":
		return new(*EventGuildMembersChunk)
	case "GUILD_ROLE_CREATE":
		return new(*EventGuildRoleCreate)
	case "GUILD_ROLE_UPDATE":
		return new(*EventGuildRoleUpdate)
	case "GUILD_ROLE_DELETE":
		return new(*EventGuildRoleDelete)
	case "MESSAGE_CREATE":
		return new(*EventMessageCreate)
	case "MESSAGE_UPDATE":
		return new(*EventMessageUpdate)
	case "MESSAGE_DELETE":
		return new(*EventMessageDelete)
	case "MESSAGE_DELETE_BULK":
		return new(*EventMessageDeleteBulk)
	case "MESSAGE_REACTION_ADD":
		return new(*EventMessageReactionAdd)
	case "MESSAGE_REACTION_REMOVE":
		return new(*EventMessageReactionRemove)
	case "MESSAGE_REACTION_REMOVE_ALL":
		return new(*EventMessageReactionRemoveAll)
	case "PRESENCE_UPDATE":
		return new(*EventPresenceUpdate)
	case "TYPING_START":
		return new(*EventTypingStart)
	case "USER_UPDATE":
		return new(*EventUserUpdate)
	case "VOICE_STATE_UPDATE":
		return new(*EventVoiceStateUpdate)
	case "VOICE_SERVER_UPDATE":
		return new(*eventVoiceServerUpdate)
	default:
		panic("unknown event")
	}
}
