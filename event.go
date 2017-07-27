package discgo

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
)

type eventReady struct {
	V               int             `json:"v"`
	User            *ModelUser      `json:"user"`
	PrivateChannels []*ModelChannel `json:"private_channels"`
	Guilds          []*ModelGuild   `json:"guilds"`
	SessionID       string          `json:"session_id"`
	Trace           []string        `json:"_trace"`
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
	GuildID string `json:"guild_id"`
}

type EventGuildMemberRemove struct {
	User    *ModelUser `json:"user"`
	GuildID string     `json:"guild_id"`
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
	ModelGameTypeGame      = iota
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

type EventMux map[string]interface{}

func NewEventMux() EventMux {
	return make(map[string]interface{})
}

func (em EventMux) Register(fn interface{}) {
	switch fn.(type) {
	case func(ctx context.Context, conn *Conn, e *eventReady) error:
		em["READY"] = fn
	case func(ctx context.Context, conn *Conn, e *eventResumed) error:
		em["RESUMED"] = fn
	case func(ctx context.Context, conn *Conn, e *EventChannelCreate) error:
		em["CHANNEL_CREATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventChannelUpdate) error:
		em["CHANNEL_UPDATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventChannelDelete) error:
		em["CHANNEL_DELETE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildCreate) error:
		em["GUILD_CREATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildUpdate) error:
		em["GUILD_UPDATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildDelete) error:
		em["GUILD_DELETE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildBanAdd) error:
		em["GUILD_BAN_ADD"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildBanRemove) error:
		em["GUILD_BAN_REMOVE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildEmojisUpdate) error:
		em["GUILD_EMOJIS_UPDATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildIntegrationsUpdate) error:
		em["GUILD_INTEGRATIONS_UPDATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildMemberAdd) error:
		em["GUILD_MEMBER_ADD"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildMemberRemove) error:
		em["GUILD_MEMBER_REMOVE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildMemberUpdate) error:
		em["GUILD_MEMBER_UPDATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildMembersChunk) error:
		em["GUILD_MEMBERS_CHUNK"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildRoleCreate) error:
		em["GUILD_ROLE_CREATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildRoleUpdate) error:
		em["GUILD_ROLE_UPDATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventGuildRoleDelete) error:
		em["GUILD_ROLE_DELETE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventMessageCreate) error:
		em["MESSAGE_CREATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventMessageUpdate) error:
		em["MESSAGE_UPDATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventMessageDelete) error:
		em["MESSAGE_DELETE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventMessageDeleteBulk) error:
		em["MESSAGE_DELETE_BULK"] = fn
	case func(ctx context.Context, conn *Conn, e *EventMessageReactionAdd) error:
		em["MESSAGE_REACTION_ADD"] = fn
	case func(ctx context.Context, conn *Conn, e *EventMessageReactionRemove) error:
		em["MESSAGE_REACTION_REMOVE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventMessageReactionRemoveAll) error:
		em["MESSAGE_REACTION_REMOVE_ALL"] = fn
	case func(ctx context.Context, conn *Conn, e *EventPresenceUpdate) error:
		em["PRESENCE_UPDATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventTypingStart) error:
		em["TYPING_START"] = fn
	case func(ctx context.Context, conn *Conn, e *EventUserUpdate) error:
		em["USER_UPDATE"] = fn
	case func(ctx context.Context, conn *Conn, e *EventVoiceStateUpdate) error:
		em["VOICE_STATE_UPDATE"] = fn
	case func(ctx context.Context, conn *Conn, e *eventVoiceServerUpdate) error:
		em["VOICE_SERVER_UPDATE"] = fn
	default:
		panic("unknown event handler signature")
	}
}

func (em EventMux) getHandler(ctx context.Context, conn *Conn, p *receivedPayload) (func() error, error) {
	h, ok := em[p.Type]
	if !ok {
		// Discord better not be sending unknown events.
		return nil, nil
	}
	e := reflect.New(reflect.TypeOf(h).In(2).Elem())
	err := json.Unmarshal(p.Data, e.Interface())
	if err != nil {
		return nil, err
	}
	return func() error {
		args := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(conn), e}
		err = reflect.ValueOf(h).Call(args)[0].Interface().(error)
		if err != nil {
			return &EventHandlerError{
				EventName: p.Type,
				Event:     e.Interface(),
				Err:       err,
			}
		}
		return nil
	}, nil
}

// State stored from websocket events.
type State struct {
	sessionID string
	eventMux EventMux

	sync.RWMutex
	user       *ModelUser
	dmChannels map[string]*ModelChannel
	guilds   map[string]*ModelGuild
	channels map[string]*ModelChannel
}

func newState() *State {
	s := &State{
		eventMux: NewEventMux(),
		dmChannels: make(map[string]*ModelChannel),
		guilds: make(map[string]*ModelGuild),
		channels: make(map[string]*ModelChannel),
	}

	s.eventMux.Register(s.ready)
	s.eventMux.Register(s.createChannel)

	return s
}

func (s *State) ready(ctx context.Context, conn *Conn, e *eventReady) error {
	s.sessionID = e.SessionID
	s.Lock()
	defer s.Unlock()
	s.user = e.User
	for _, c := range e.PrivateChannels {
		s.dmChannels[c.ID] = c
	}
	for _, g := range e.Guilds {
		s.guilds[g.ID] = g
	}
	return nil
}

func (s *State) createChannel(ctx context.Context, conn *Conn, e *EventChannelCreate) error {
	return s.insertChannel(&e.ModelChannel)
}

func (s *State) updateChannel(ctx context.Context, conn *Conn, e *EventChannelUpdate) error {
	return s.insertChannel(&e.ModelChannel)
}

func (s *State) insertChannel(c *ModelChannel) error {
	s.Lock()
	defer s.Unlock()

	if c.Type == ModelChannelTypeDM || c.Type == ModelChannelTypeGroupDM {
		s.dmChannels[c.ID] = c
	} else {
		s.channels[c.ID] = c

		g, ok := s.guilds[*c.GuildID]
		if !ok {
			// TODO on panics, print out the associated event.
			panic("a channel created for an unknown guild")
		}
		g.Channels = append(g.Channels, c)
	}
	return nil
}

func (s *State) deleteChannel(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()

	if e.Type == ModelChannelTypeDM || e.Type == ModelChannelTypeGroupDM {
		delete(s.dmChannels, e.ID)
	} else {
		delete(s.channels, e.ID)

		g, ok := s.guilds[*e.GuildID]
		if !ok {
			// TODO on panics, print out the associated event.
			panic("a channel created for an unknown guild")
		}
		for i, c := range g.Channels {
			if c.ID == e.ID {
				g.Channels = append(g.Channels[:i], g.Channels[i+1:]...)
				break
			}
		}
	}
}

func (s *State) Channel(cID string) (*ModelChannel, bool) {
	s.RLock()
	defer s.RUnlock()

	c, ok := s.dmChannels[cID]
	if ok {

	}

	c, ok = s.channels[cID]
	if ok {

	}
}

// TODO how do I handle a createGuild in response to a guild being unavailable?
// Should the error event even run?
func (s *State) createGuild(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) updateGuild(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) deleteGuild(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) addGuildBan(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) updateGuildEmojis(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) addGuildMember(ctx context.Context, conn *Conn, e *EventGuildMemberAdd) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) removeGuildMember(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) updateGuildMember(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) chunkGuildMembers(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) createGuildRole(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) updateGuildRole(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) deleteGuildRole(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) createMessage(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) updateMessage(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) deleteMessage(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) bulkDeleteMessages(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) addMessageReaction(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) removeMessageReaction(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) updatePresence(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) startTyping(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) updateUser(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) updateVoiceState(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}

func (s *State) updateVoiceServer(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.Lock()
	defer s.Unlock()
}
