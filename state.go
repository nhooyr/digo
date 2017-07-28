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

	// TODO seperate mutex for guilds map, channels map etc?
	mu         sync.RWMutex
	user       *ModelUser
	dmChannels map[string]*StateChannel
	guilds     map[string]*StateGuild
	channels   map[string]*StateChannel
}

func newState() *State {
	s := &State{
		eventMux:   newEventMux(),
		dmChannels: make(map[string]*StateChannel),
		guilds:     make(map[string]*StateGuild),
		channels:   make(map[string]*StateChannel),
	}

	s.eventMux.Register(s.ready)
	s.eventMux.Register(s.createChannel)

	return s
}

type StateGuild struct {
	mu sync.RWMutex
	g  *ModelGuild
	// TODO seperate mutexes for these?
	large       bool
	unavailable bool
	memberCount int
	voiceStates []*ModelVoiceState
	members     []*ModelGuildMember
	channels    []*StateChannel
	presences   []*ModelPresence
}

func (sg *StateGuild) ID() string {
	// It's immutable for sure but I'm doing this anyway because I'm gonna replace
	// the entire ModelGuild pointer with another on a GuildUpdate event.
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.g.ID
}

func (sg *StateGuild) Name() string {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.g.Name
}

func (sg *StateGuild) Icon() string {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.g.Icon
}

func (sg *StateGuild) Splash() string {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.g.Splash
}

func (sg *StateGuild) OwnerID() string {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.g.OwnerID
}

func (sg *StateGuild) Region() string {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.g.Region
}

func (sg *StateGuild) AFKChannelID() string {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.g.AFKChannelID
}

func (sg *StateGuild) AFKTimeout() int {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.g.AFKTimeout
}

func (sg *StateGuild) EmbedEnabled() bool {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.g.EmbedEnabled
}

func (sg *StateGuild) EmbedChannelID() string {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.g.EmbedChannelID
}

func (sg *StateGuild) VerificationLevel() int {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.g.VerificationLevel
}

func (sg *StateGuild) DefaultMessageNotificationLevel() int {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.g.DefaultMessageNotificationLevel
}

func (sg *StateGuild) Roles() []*ModelRole {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	roles := make([]*ModelRole, len(sg.g.Roles))
	copy(roles, sg.g.Roles)
	return roles
}

func (sg *StateGuild) Emojis() []*ModelGuildEmoji {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.g.Emojis
}

func (sg *StateGuild) Features() []string {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.g.Features
}

func (sg *StateGuild) MFALevel() int {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.g.MFALevel
}

func (sg *StateGuild) JoinedAt() time.Time {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.g.JoinedAt
}

func (sg *StateGuild) Large() bool {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.large
}

// No need for this to be exported, just a helper function. Couldn't think of a better name
func (sg *StateGuild) Unavailable() bool {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.unavailable
}

func (sg *StateGuild) MemberCount() int {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.memberCount
}

func (sg *StateGuild) VoiceStates() []*ModelVoiceState {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	voiceStates := make([]*ModelVoiceState, len(sg.voiceStates))
	copy(voiceStates, sg.voiceStates)
	return voiceStates
}

func (sg *StateGuild) Members() []*ModelGuildMember {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	members := make([]*ModelGuildMember, len(sg.members))
	copy(members, sg.members)
	return members
}

func (sg *StateGuild) Channels() []*StateChannel {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	channels := make([]*StateChannel, len(sg.channels))
	copy(channels, sg.channels)
	return channels
}

func (sg *StateGuild) Presences() []*ModelPresence {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	presences := make([]*ModelPresence, len(sg.presences))
	copy(presences, sg.presences)
	return presences
}

type StateChannel struct {
	mu    sync.RWMutex
	c     *ModelChannel
	guild *StateGuild
	// TODO another mutex for this?
	messages []*ModelMessage
}

func (sc *StateChannel) ID() string {
	// It's immutable for sure but I'm doing this anyway because I'm gonna replace
	// the entire ModelChannel pointer with another on a ChannelUpdate event.
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.c.ID
}

func (sc *StateChannel) Type() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.c.Type
}

func (sc *StateChannel) Guild() *StateGuild {
	// Guaranteed to never change.
	return sc.guild
}

func (sc *StateChannel) Position() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.c.Position
}

func (sc *StateChannel) PermissionOverwrites() []*ModelPermissionOverwrite {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	permissionOverwrites := make([]*ModelPermissionOverwrite, len(sc.c.PermissionOverwrites))
	copy(permissionOverwrites, sc.c.PermissionOverwrites)
	return permissionOverwrites
}

func (sc *StateChannel) Name() string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.c.Name
}

func (sc *StateChannel) Topic() string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.c.Topic
}

func (sc *StateChannel) LastMessageID() string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.c.LastMessageID
}

func (sc *StateChannel) Bitrate() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.c.Bitrate
}

func (sc *StateChannel) UserLimit() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.c.UserLimit
}

func (sc *StateChannel) Recipients() []*ModelUser {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	recipients := make([]*ModelUser, len(sc.c.Recipients))
	copy(recipients, sc.c.Recipients)
	return recipients
}

func (sc *StateChannel) Icon() string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.c.Icon
}

func (sc *StateChannel) OwnerID() string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.c.OwnerID
}

func (sc *StateChannel) ApplicationID() string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.c.ApplicationID
}

// Messages returns a copy of the current messages. The last message is the most recent.
func (sc *StateChannel) Messages() []*ModelMessage {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	messages := make([]*ModelMessage, len(sc.messages))
	copy(messages, sc.messages)
	return messages
}

func (s *State) ready(ctx context.Context, conn *Conn, e *eventReady) error {
	// Access to this is serialized by the Conn goroutines so we don't need to protect it.
	s.sessionID = e.SessionID

	s.mu.Lock()
	defer s.mu.Unlock()
	s.user = e.User
	for _, c := range e.PrivateChannels {
		s.dmChannels[c.ID] = &StateChannel{c: c}
	}
	for _, ee := range e.Guilds {
		s.guilds[ee.ID] = &StateGuild{g: &ee.ModelGuild, unavailable: ee.Unavailable}
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
	sc := &StateChannel{c: c}

	s.mu.Lock()

	if c.Type == ModelChannelTypeDM || c.Type == ModelChannelTypeGroupDM {
		s.dmChannels[c.ID] = sc
		s.mu.Unlock()
		return nil
	}

	s.channels[c.ID] = sc
	sg, ok := s.guilds[c.GuildID]
	s.mu.Unlock()

	if !ok {
		return errors.New("a channel created/updated for an unknown guild")
	}

	sc.mu.Lock()
	sc.guild = sg
	sc.mu.Unlock()

	sg.mu.Lock()
	sg.channels = append(sg.channels, sc)
	sg.mu.Unlock()

	return nil
}

func (s *State) deleteChannel(ctx context.Context, conn *Conn, e *EventChannelDelete) error {
	s.mu.Lock()

	if e.Type == ModelChannelTypeDM || e.Type == ModelChannelTypeGroupDM {
		delete(s.dmChannels, e.ID)
		s.mu.Unlock()
		return nil
	}

	delete(s.channels, e.ID)
	sg, ok := s.guilds[e.GuildID]
	s.mu.Unlock()

	if !ok {
		return errors.New("a channel deleted for an unknown guild")
	}

	for i, sc := range sg.Channels() {
		// I don't think ID() helper is necessary, but cannot be too safe.
		if sc.ID() == e.ID {
			sg.mu.Lock()
			sg.channels = append(sg.channels[:i], sg.channels[i+1:]...)
			sg.mu.Unlock()
			return nil
		}
	}
	return errors.New("channel removed from a guild in which it never existed?")
}

var errHandled = errors.New("no need to handle the event further")

func (s *State) createGuild(ctx context.Context, conn *Conn, e *EventGuildCreate) error {
	sg := &StateGuild{
		g: &e.ModelGuild,
		large: e.Large,
		unavailable: e.Unavailable,
		memberCount: e.MemberCount,
		voiceStates: e.VoiceStates,
		members: e.Members,
		presences: e.Presences,
	}

	s.mu.Lock()
	for _, c := range e.Channels {
		sc := &StateChannel{
			c: c,
			guild: sg,
		}
		sg.channels = append(sg.channels, sc)
		s.channels[c.ID] = sc
	}
	sgOld, ok := s.guilds[e.ID]
	s.guilds[e.ID] = sg
	defer s.mu.Unlock()

	if ok {
		if sgOld.Unavailable() {
			// Guild is available again or was lazily loaded.
			// Either way, don't run any GuildCreate event handlers.
			return errHandled
		}
		// We updated the guild map even though the state is now corrupt.
		// Shouldn't really be an issue though.
		return errors.New("guild already exists?")
	}
	return nil
}

func (s *State) updateGuild(ctx context.Context, conn *Conn, e *EventGuildUpdate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sg, ok := s.guilds[e.ID]
	if !ok {
		return errors.New("non existing guild updated?")
	}
	sg.g = &e.ModelGuild
	return nil
}

func (s *State) deleteGuild(ctx context.Context, conn *Conn, e *EventGuildDelete) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sg, ok := s.guilds[e.ID]
	if !ok {
		return errors.New("non existing guild deleted?")
	}
	if e.Unavailable {
		sg.mu.Lock()
		sg.unavailable = true
		sg.mu.Lock()
	} else {
		delete(s.guilds, e.ID)
	}
	for _, sc := range sg.channels {
		// I don't think ID() helper is necessary, but cannot be too safe.
		delete(s.channels, sc.ID())
	}
}

func (s *State) addGuildBan(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *State) updateGuildEmojis(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *State) addGuildMember(ctx context.Context, conn *Conn, e *EventGuildMemberAdd) {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *State) removeGuildMember(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *State) updateGuildMember(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *State) chunkGuildMembers(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *State) createGuildRole(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *State) updateGuildRole(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *State) deleteGuildRole(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *State) createMessage(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *State) updateMessage(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *State) deleteMessage(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *State) bulkDeleteMessages(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *State) addMessageReaction(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *State) removeMessageReaction(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *State) updatePresence(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *State) startTyping(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *State) updateUser(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *State) updateVoiceState(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *State) updateVoiceServer(ctx context.Context, conn *Conn, e *EventChannelDelete) {
	s.mu.Lock()
	defer s.mu.Unlock()
}
