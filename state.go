package discgo

import (
	"context"
	"errors"
	"sync"
	"time"
)

// State stored from websocket events.
type State struct {
	sessionID string
	eventMux  EventMux

	sync.RWMutex
	user       *ModelUser
	dmChannels map[string]*ModelChannel
	guilds     map[string]*ModelGuild
	channels   map[string]*ModelChannel
}

func newState() *State {
	s := &State{
		eventMux:   newEventMux(),
		dmChannels: make(map[string]*ModelChannel),
		guilds:     make(map[string]*ModelGuild),
		channels:   make(map[string]*ModelChannel),
	}

	s.eventMux.Register(s.ready)
	s.eventMux.Register(s.createChannel)

	return s
}

type StateGuild struct {
	mu sync.RWMutex
	// Odd struct indeed but oh well.
	g        *EventGuildCreate
	channels map[string]*StateChannel
}

func (s *StateGuild) ID() string {
	// It's immutable for sure but I'm doing this anyway because I'm gonna replace
	// the entire ModelGuild pointer with another on a GuildUpdate event.
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.g.ID
}

func (s *StateGuild) Name() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.g.Name
}

func (s *StateGuild) Icon() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.g.Icon
}

func (s *StateGuild) Splash() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.g.Splash
}

func (s *StateGuild) OwnerID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.g.OwnerID
}

func (s *StateGuild) Region() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.g.Region
}

func (s *StateGuild) AFKChannelID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.g.AFKChannelID
}

func (s *StateGuild) AFKTimeout() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.g.AFKTimeout
}

func (s *StateGuild) EmbedEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.g.EmbedEnabled
}

func (s *StateGuild) EmbedChannelID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.g.EmbedChannelID
}

func (s *StateGuild) VerificationLevel() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.g.VerificationLevel
}

func (s *StateGuild) DefaultMessageNotificationLevel() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.g.DefaultMessageNotificationLevel
}

func (s *StateGuild) Roles() []*ModelRole {
	s.mu.RLock()
	defer s.mu.RUnlock()
	roles := make([]*ModelRole, len(s.g.Roles))
	copy(roles, s.g.Roles)
	return roles
}

func (s *StateGuild) Emojis() []*ModelGuildEmoji {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.g.Emojis
}

func (s *StateGuild) Features() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.g.Features
}

func (s *StateGuild) MFALevel() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.g.MFALevel
}

func (s *StateGuild) JoinedAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.g.JoinedAt
}

func (s *StateGuild) Large() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.g.Large
}

func (s *StateGuild) MemberCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.g.MemberCount
}

func (s *StateGuild) VoiceStates() []*ModelVoiceState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	voiceStates := make([]*ModelVoiceState, len(s.g.VoiceStates))
	copy(voiceStates, s.g.VoiceStates)
	return voiceStates
}

func (s *StateGuild) Members() []*ModelGuildMember {
	s.mu.RLock()
	defer s.mu.RUnlock()
	members := make([]*ModelGuildMember, len(s.g.Members))
	copy(members, s.g.Members)
	return members
}

func (s *StateGuild) Presences() []*ModelPresence {
	s.mu.RLock()
	defer s.mu.RUnlock()
	presences := make([]*ModelPresence, len(s.g.Presences))
	copy(presences, s.g.Presences)
	return presences
}

type StateChannel struct {
	mu       sync.RWMutex
	c        *ModelChannel
	messages []*ModelMessage
}

func (s *State) ready(ctx context.Context, conn *Conn, e *eventReady) error {
	// Access to this is serialized by the Conn goroutines so we don't need to protect it.
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
			return errors.New("a channel created for an unknown guild")
		}
		g.Channels = append(g.Channels, c)
	}
	return nil
}

func (s *State) deleteChannel(ctx context.Context, conn *Conn, e *EventChannelDelete) error {
	s.Lock()
	defer s.Unlock()

	if e.Type == ModelChannelTypeDM || e.Type == ModelChannelTypeGroupDM {
		delete(s.dmChannels, e.ID)
	} else {
		delete(s.channels, e.ID)

		g, ok := s.guilds[*e.GuildID]
		if !ok {
			return errors.New("a channel deleted for an unknown guild")
		}
		for i, c := range g.Channels {
			if c.ID == e.ID {
				g.Channels = append(g.Channels[:i], g.Channels[i+1:]...)
				break
			}
		}
	}
}

var errHandled = errors.New("no need to handle the event further")

func (s *State) createGuild(ctx context.Context, conn *Conn, e *EventGuildCreate) error {
	s.Lock()
	defer s.Unlock()
	return nil
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
