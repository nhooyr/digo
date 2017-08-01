package discgo

import (
	"context"
	"errors"
	"sync"
	"time"
)

type StateGuild struct {
	modelMu                         sync.RWMutex
	id                              string
	name                            string
	icon                            string
	splash                          string
	ownerID                         string
	region                          string
	afkChannelID                    string
	afkTimeout                      int
	embedEnabled                    bool
	embedChannelID                  string
	verificationLevel               int
	defaultMessageNotificationLevel int
	features                        []string
	mfaLevel                        int
	joinedAt                        time.Time

	rolesMu sync.RWMutex
	roles   map[string]*ModelRole

	emojisMu sync.RWMutex
	emojis   []*ModelGuildEmoji

	large       bool
	unavailable bool

	voiceStatesMu sync.RWMutex
	voiceStates   map[string]*ModelVoiceState

	membersMu   sync.RWMutex
	members     map[string]*ModelGuildMember
	memberCount int

	channelsMu sync.RWMutex
	channels   map[string]*StateChannel

	presencesMu sync.RWMutex
	// TODO how should I handle improperly typed updates?
	presences map[string]*ModelPresence
}

func (sg *StateGuild) updateFromModel(g *ModelGuild) {
	sg.modelMu.Lock()
	sg.rolesMu.Lock()
	sg.emojisMu.Lock()

	sg.id = g.ID
	sg.name = g.Name
	sg.icon = g.Icon
	sg.splash = g.Splash
	sg.ownerID = g.OwnerID
	sg.region = g.Region
	sg.afkChannelID = g.AFKChannelID
	sg.afkTimeout = g.AFKTimeout
	sg.embedEnabled = g.EmbedEnabled
	sg.embedChannelID = g.EmbedChannelID
	sg.verificationLevel = g.VerificationLevel
	sg.defaultMessageNotificationLevel = g.DefaultMessageNotificationLevel
	sg.roles = make(map[string]*ModelRole)
	for _, r := range g.Roles {
		sg.roles[r.ID] = r
	}
	sg.emojis = g.Emojis
	sg.features = g.Features
	sg.mfaLevel = g.MFALevel
	sg.joinedAt = g.JoinedAt

	sg.modelMu.Unlock()
	sg.rolesMu.Unlock()
	sg.emojisMu.Unlock()
}

func (sg *StateGuild) ID() string {
	// It's immutable for sure but I'm doing this anyway for consistency.
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	return sg.id
}

func (sg *StateGuild) Name() string {
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	return sg.name
}

func (sg *StateGuild) Icon() string {
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	return sg.icon
}

func (sg *StateGuild) Splash() string {
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	return sg.splash
}

func (sg *StateGuild) OwnerID() string {
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	return sg.ownerID
}

func (sg *StateGuild) Region() string {
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	return sg.region
}

func (sg *StateGuild) AFKChannelID() string {
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	return sg.afkChannelID
}

func (sg *StateGuild) AFKTimeout() int {
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	return sg.afkTimeout
}

func (sg *StateGuild) EmbedEnabled() bool {
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	return sg.embedEnabled
}

func (sg *StateGuild) EmbedChannelID() string {
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	return sg.embedChannelID
}

func (sg *StateGuild) VerificationLevel() int {
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	return sg.verificationLevel
}

func (sg *StateGuild) DefaultMessageNotificationLevel() int {
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	return sg.defaultMessageNotificationLevel
}

func (sg *StateGuild) Role(rID string) (*ModelRole, bool) {
	sg.rolesMu.RLock()
	r, ok := sg.roles[rID]
	sg.rolesMu.RUnlock()
	return r, ok
}

func (sg *StateGuild) Roles() []*ModelRole {
	sg.rolesMu.RLock()
	roles := make([]*ModelRole, 0, len(sg.roles))
	for _, r := range sg.roles {
		roles = append(roles, r)
	}
	sg.rolesMu.RUnlock()
	return roles
}

func (sg *StateGuild) Emojis() []*ModelGuildEmoji {
	sg.emojisMu.RLock()
	emojis := make([]*ModelGuildEmoji, len(sg.emojis))
	copy(emojis, sg.emojis)
	sg.emojisMu.RUnlock()
	return emojis
}

func (sg *StateGuild) Features() []string {
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	return sg.features
}

func (sg *StateGuild) MFALevel() int {
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	return sg.mfaLevel
}

func (sg *StateGuild) JoinedAt() time.Time {
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	return sg.joinedAt
}

func (sg *StateGuild) Large() bool {
	return sg.large
}

func (sg *StateGuild) VoiceStates() []*ModelVoiceState {
	sg.voiceStatesMu.RLock()
	voiceStates := make([]*ModelVoiceState, 0, len(sg.voiceStates))
	for _, vs := range sg.voiceStates {
		voiceStates = append(voiceStates, vs)
	}
	sg.voiceStatesMu.RUnlock()
	return voiceStates
}

// Only cause discord API sends it.
func (sg *StateGuild) MemberCount() int {
	sg.membersMu.RLock()
	defer sg.membersMu.RUnlock()
	return sg.memberCount
}

func (sg *StateGuild) Member(uID string) (*ModelGuildMember, bool) {
	sg.membersMu.RLock()
	sgm, ok := sg.members[uID]
	sg.membersMu.RUnlock()
	return sgm, ok
}

func (sg *StateGuild) Members() []*ModelGuildMember {
	sg.membersMu.RLock()
	members := make([]*ModelGuildMember, 0, len(sg.members))
	for _, gm := range sg.members {
		members = append(members, gm)
	}
	sg.membersMu.RUnlock()
	return members
}

func (sg *StateGuild) Channel(cID string) (*StateChannel, bool) {
	sg.channelsMu.RLock()
	sc, ok := sg.channels[cID]
	sg.channelsMu.RUnlock()
	return sc, ok
}

func (sg *StateGuild) Channels() []*StateChannel {
	sg.channelsMu.RLock()
	channels := make([]*StateChannel, 0, len(sg.channels))
	for _, sc := range sg.channels {
		channels = append(channels, sc)
	}
	sg.channelsMu.RUnlock()
	return channels
}

func (sg *StateGuild) Presence(uID string) (*ModelPresence, bool) {
	sg.presencesMu.RLock()
	sp, ok := sg.presences[uID]
	sg.presencesMu.RUnlock()
	return sp, ok
}

func (sg *StateGuild) Presences() []*ModelPresence {
	sg.presencesMu.RLock()
	presences := make([]*ModelPresence, 0, len(sg.presences))
	for _, p := range sg.presences {
		presences = append(presences, p)
	}
	sg.presencesMu.RUnlock()
	return presences
}

type StateChannel struct {
	guild *StateGuild

	mu                   sync.RWMutex
	id                   string
	chanType             int
	position             int
	permissionOverwrites []*ModelPermissionOverwrite
	name                 string
	topic                string
	lastMessageID        string
	bitrate              int
	userLimit            int
	recipients           []*ModelUser
	icon                 string
	ownerID              string
	applicationID        string

	messagesMu sync.RWMutex
	messages   []*ModelMessage
}

func (sc *StateChannel) ID() string {
	// Immutable but doing this for consistency.
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.id
}

func (sc *StateChannel) Type() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.chanType
}

func (sc *StateChannel) Guild() *StateGuild {
	// Guaranteed to never change.
	return sc.guild
}

func (sc *StateChannel) Position() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.position
}

func (sc *StateChannel) PermissionOverwrites() []*ModelPermissionOverwrite {
	sc.mu.RLock()
	permissionOverwrites := make([]*ModelPermissionOverwrite, len(sc.permissionOverwrites))
	copy(permissionOverwrites, sc.permissionOverwrites)
	sc.mu.RUnlock()
	return permissionOverwrites
}

func (sc *StateChannel) Name() string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.name
}

func (sc *StateChannel) Topic() string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.topic
}

func (sc *StateChannel) LastMessageID() string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.lastMessageID
}

func (sc *StateChannel) Bitrate() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.bitrate
}

func (sc *StateChannel) UserLimit() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.userLimit
}

func (sc *StateChannel) Recipients() []*ModelUser {
	sc.mu.RLock()
	recipients := make([]*ModelUser, len(sc.recipients))
	copy(recipients, sc.recipients)
	sc.mu.RUnlock()
	return recipients
}

func (sc *StateChannel) Icon() string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.icon
}

func (sc *StateChannel) OwnerID() string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.ownerID
}

func (sc *StateChannel) ApplicationID() string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.applicationID
}

// Messages returns a copy of the current messages. The last message is the most recent.
func (sc *StateChannel) Messages() []*ModelMessage {
	sc.messagesMu.RLock()
	messages := make([]*ModelMessage, len(sc.messages))
	copy(messages, sc.messages)
	sc.messagesMu.RUnlock()
	return messages
}

func (sc *StateChannel) updateFromModel(c *ModelChannel) {
	sc.mu.Lock()
	sc.id = c.ID
	sc.chanType = c.Type
	sc.position = c.Position
	sc.permissionOverwrites = c.PermissionOverwrites
	sc.name = c.Name
	sc.topic = c.Topic
	sc.lastMessageID = c.LastMessageID
	sc.bitrate = c.Bitrate
	sc.userLimit = c.UserLimit
	sc.recipients = c.Recipients
	sc.icon = c.Icon
	sc.ownerID = c.OwnerID
	sc.applicationID = c.ApplicationID
	sc.mu.Unlock()
}

func (sc *StateChannel) appendMessage() {

}

// State stored from websocket events.
// TODO should I just expose a single RWMutex for state and allow caller to lock/unlock? It would allow them to have better control. E.g. they would never get inconsistent data because they can hold RLock for longer than a single call.
type State struct {
	sessionID string

	// TODO event handlers send new transformed data further? E.g. not the raw events but StateGuild etc?
	// TODO does the user update event only apply to the current user or all users?
	userMu sync.RWMutex
	user   *ModelUser

	dmChannelsMu sync.RWMutex
	dmChannels   map[string]*StateChannel

	guildsMu sync.RWMutex
	guilds   map[string]*StateGuild
	// I have a separate map for guild channels because messages only store Channel IDs, no Guild IDs.
	// So if someone wanted to find the guild channel in which a message was sent, they would have to search
	// all guilds and their channels.
	guildChannelsMu sync.RWMutex
	guildChannels   map[string]*StateChannel
}

func (s *State) Guild(gID string) (*StateGuild, bool) {
	s.guildsMu.RLock()
	sg, ok := s.guilds[gID]
	s.guildsMu.RUnlock()
	if ok && sg.unavailable {
		return nil, false
	}
	return sg, ok
}

func (s *State) Channel(cID string) (*StateChannel, bool) {
	// Check Guild Channels first, DMChannels are used less often.
	s.guildChannelsMu.RLock()
	sc, ok := s.guildChannels[cID]
	s.guildChannelsMu.RUnlock()
	if !ok {
		s.dmChannelsMu.RLock()
		sc, ok = s.dmChannels[cID]
		s.dmChannelsMu.RUnlock()
	}
	return sc, ok
}

func (s *State) handle(e interface{}, handler EventHandler) {
	switch e.(type) {
	// TODO
	}
}

func (s *State) ready(ctx context.Context, conn *Conn, e *eventReady) error {
	// Access to this is serialized by the Conn goroutines so we don't need to protect it.
	s.sessionID = e.SessionID

	s.userMu.Lock()
	s.dmChannelsMu.Lock()
	s.guildsMu.Lock()
	s.guildChannelsMu.Lock()

	s.user = e.User

	s.dmChannels = make(map[string]*StateChannel)
	for _, c := range e.PrivateChannels {
		sc := new(StateChannel)
		sc.updateFromModel(c)
		s.dmChannels[c.ID] = sc
	}

	s.guilds = make(map[string]*StateGuild)
	for _, ee := range e.Guilds {
		s.guilds[ee.ID] = &StateGuild{unavailable: true}
	}

	s.guildChannels = make(map[string]*StateChannel)

	s.userMu.Unlock()
	s.dmChannelsMu.Unlock()
	s.guildsMu.Unlock()
	s.guildChannelsMu.Unlock()

	return nil
}

func (s *State) createChannel(ctx context.Context, conn *Conn, e *EventChannelCreate) error {
	return s.insertChannel(&e.ModelChannel)
}

func (s *State) updateChannel(ctx context.Context, conn *Conn, e *EventChannelUpdate) error {
	return s.insertChannel(&e.ModelChannel)
}

func (s *State) insertChannel(c *ModelChannel) error {
	if c.Type == ModelChannelTypeDM || c.Type == ModelChannelTypeGroupDM {
		s.dmChannelsMu.RLock()
		sc, ok := s.dmChannels[c.ID]
		s.dmChannelsMu.RUnlock()
		if !ok {
			sc = new(StateChannel)
		}
		sc.updateFromModel(c)
		if !ok {
			s.dmChannelsMu.Lock()
			s.dmChannels[c.ID] = sc
			s.dmChannelsMu.Unlock()
		}
		return nil
	}

	s.guildChannelsMu.RLock()
	sc, ok := s.guildChannels[c.ID]
	s.guildChannelsMu.RUnlock()

	if ok {
		sc.updateFromModel(c)
		return nil
	}

	sg, ok := s.Guild(c.GuildID)
	if !ok {
		return errors.New("a channel created/updated for an unknown guild")
	}

	sc = new(StateChannel)
	sc.updateFromModel(c)

	s.guildChannelsMu.Lock()
	sg.channelsMu.Lock()

	s.guildChannels[sc.id] = sc
	sg.channels[c.ID] = sc

	s.guildChannelsMu.Unlock()
	sg.channelsMu.Unlock()
	return nil
}

func (s *State) deleteChannel(ctx context.Context, conn *Conn, e *EventChannelDelete) error {
	if e.Type == ModelChannelTypeDM || e.Type == ModelChannelTypeGroupDM {
		s.dmChannelsMu.Lock()
		delete(s.dmChannels, e.ID)
		s.dmChannelsMu.Unlock()
		return nil
	}

	sg, ok := s.Guild(e.GuildID)
	if !ok {
		return errors.New("a channel deleted for an unknown guild")
	}

	s.guildChannelsMu.Lock()
	sg.channelsMu.Lock()

	delete(s.guildChannels, e.ID)
	delete(sg.channels, e.ID)

	s.guildChannelsMu.Unlock()
	sg.channelsMu.Unlock()

	return nil
}

var errHandled = errors.New("no need to handle the event further")

func (s *State) createGuild(ctx context.Context, conn *Conn, e *EventGuildCreate) error {
	sg := new(StateGuild)
	sg.updateFromModel(&e.ModelGuild)

	sg.large = e.Large

	sg.voiceStates = make(map[string]*ModelVoiceState)
	for _, vs := range e.VoiceStates {
		sg.voiceStates[vs.UserID] = vs
	}

	sg.memberCount = e.MemberCount
	sg.members = make(map[string]*ModelGuildMember)
	for _, gm := range e.Members {
		sg.members[gm.User.ID] = gm
	}

	sg.presences = make(map[string]*ModelPresence)
	for _, p := range e.Presences {
		sg.presences[p.User.ID] = p
	}

	s.guildsMu.Lock()
	s.guildChannelsMu.Lock()

	for _, c := range e.Channels {
		sc := new(StateChannel)
		// TODO updateFromModel locks but this is a brand new StateChannel so no locking necessary. maybe change later.
		sc.updateFromModel(c)

		sg.channels[c.ID] = sc
		s.guildChannels[c.ID] = sc
	}

	sgOld, ok := s.guilds[e.ID]
	s.guilds[e.ID] = sg

	s.guildsMu.Unlock()
	s.guildChannelsMu.Unlock()

	if ok {
		if sgOld.unavailable {
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
	sg, ok := s.Guild(e.ID)
	if !ok {
		return errors.New("non existing guild updated?")
	}
	sg.updateFromModel(&e.ModelGuild)
	return nil
}

func (s *State) deleteGuild(ctx context.Context, conn *Conn, e *EventGuildDelete) error {
	s.guildsMu.Lock()
	sg, ok := s.guilds[e.ID]
	if !ok {
		s.guildsMu.Unlock()
		return errors.New("non existing guild deleted")
	}
	if e.Unavailable {
		s.guilds[e.ID] = &StateGuild{
			unavailable: true,
		}
	} else {
		delete(s.guilds, e.ID)
	}
	s.guildsMu.Unlock()

	s.guildChannelsMu.Lock()
	for id := range sg.channels {
		delete(s.guildChannels, id)
	}
	s.guildChannelsMu.Unlock()

	return nil
}

// TODO maybe guild ban add and guild ban remove? Not sure....

func (s *State) updateGuildEmojis(ctx context.Context, conn *Conn, e *EventGuildEmojisUpdate) error {
	sg, ok := s.Guild(e.GuildID)
	if !ok {
		return errors.New("guild emojis updated for non existing guild")
	}
	sg.emojisMu.Lock()
	sg.emojis = e.Emojis
	sg.emojisMu.Unlock()
	return nil
}

func (s *State) addGuildMember(ctx context.Context, conn *Conn, e *EventGuildMemberAdd) error {
	sg, ok := s.Guild(e.GuildID)
	if !ok {
		return errors.New("guild member added in non existing guild")
	}
	sg.membersMu.Lock()
	sg.memberCount++
	sg.members[e.User.ID] = &e.ModelGuildMember
	sg.membersMu.Unlock()
	return nil
}

func (s *State) removeGuildMember(ctx context.Context, conn *Conn, e *EventGuildMemberRemove) error {
	sg, ok := s.Guild(e.GuildID)
	if !ok {
		return errors.New("guild member removed in non existing guild")
	}
	sg.membersMu.Lock()
	sg.memberCount--
	delete(sg.members, e.User.ID)
	sg.membersMu.Unlock()
	return nil
}

func (s *State) updateGuildMember(ctx context.Context, conn *Conn, e *EventGuildMemberUpdate) error {
	sg, ok := s.Guild(e.GuildID)
	if !ok {
		return errors.New("guild member updated in non existing guild")
	}
	sg.membersMu.Lock()
	defer sg.membersMu.Unlock()
	gm, ok := sg.members[e.User.ID]
	if !ok {
		return errors.New("guild member updated in a guild in which it never joined?")
	}
	sg.members[e.User.ID] = &ModelGuildMember{
		User:     &e.User,
		Roles:    e.Roles,
		JoinedAt: gm.JoinedAt,
		Deaf:     gm.Deaf,
		Mute:     gm.Mute,
		Nick:     &e.Nick,
	}
	return nil
}

// TODO not sure how to handle this stuff
func (s *State) chunkGuildMembers(ctx context.Context, conn *Conn, e *EventGuildMembersChunk) error {
	panic("TODO")
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
