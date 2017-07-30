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
	// TODO event handlers send new transformed data further? E.g. not the raw events but StateGuild etc?
	mu         sync.RWMutex
	user       *ModelUser
	dmChannels map[string]*StateChannel
	guilds     map[string]*StateGuild
	// I have a seperate map for this because messages only store Channel IDs, no Guild IDs.
	// So if someone wanted to find the Channel in which a message was sent, they would have to search
	// all guilds.
	channels map[string]*StateChannel
	// TOOD users      map[string]*StateUser
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

// The maps are reason I went for a modelMu instead of using the map as a sync point. E.g. instead of
// updating the values up above in the struct itself, I create a new struct and copy the values in
// and then update the map with that new struct.
// Issue with this is that if I would have to lock/unlock all the map mutexes before copying the
// struct to make any little change. Sounds dumb to me. Reason I always have to lock is incase
// the map reference is written to, e.g. when I want to dump one of the maps when I get a guild create
// after a guild has gone unavailable.
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
	roles  map[string]*ModelRole

	emojisMu sync.RWMutex
	emojis map[string]*ModelGuildEmoji

	large       bool

	unavailableMu sync.RWMutex
	unavailable bool

	memberCountMu sync.RWMutex
	memberCount int

	voiceStatesMu sync.RWMutex
	voiceStates map[string]*ModelVoiceState

	membersMu sync.RWMutex
	members     map[string]*ModelGuildMember

	channelsMu sync.RWMutex
	channels    map[string]*StateChannel

	presencesMu sync.RWMutex
	// TODO how should I handle improperly typed updates?
	presences map[string]*StatePresence
}

type StateGuildMember struct {
	StateUser
}

type StatePresence struct {
	User   *StateUser `json:"user"`
	Game   *ModelGame `json:"game"`
	Status string     `json:"status"`
}

func (sg *StateGuild) ID() string {
	// It's immutable for sure but I'm doing this anyway because I'm gonna replace
	// the entire ModelGuild pointer with another on a GuildUpdate event.
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

func (sg *StateGuild) Roles() []*ModelRole {
	sg.rolesMu.RLock()
	defer sg.rolesMu.RUnlock()
	roles := make([]*ModelRole, 0, len(sg.roles))
	for _, r := range sg.roles {
		roles = append(roles, r)
	}
	return roles
}

func (sg *StateGuild) Emojis() []*ModelGuildEmoji {
	sg.emojisMu.RLock()
	defer sg.emojisMu.RUnlock()
	emojis := make([]*ModelGuildEmoji, 0, len(sg.emojis))
	for _, e := range sg.emojis {
		emojis = append(emojis, e)
	}
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

// No need for this to be exported, just a helper function. Couldn't think of a better name
func (sg *StateGuild) Unavailable() bool {
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	return sg.unavailable
}

func (sg *StateGuild) MemberCount() int {
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	return sg.memberCount
}

func (sg *StateGuild) VoiceStates() []*ModelVoiceState {
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	voiceStates := make([]*ModelVoiceState, 0, len(sg.voiceStates))
	for _, vs := range sg.voiceStates {
		voiceStates = append(voiceStates, vs)
	}
	return voiceStates
}

func (sg *StateGuild) Members() []*ModelGuildMember {
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	members := make([]*ModelGuildMember, 0, len(sg.members))
	for _, gm := range sg.members {
		members = append(members, gm)
	}
	return members
}

func (sg *StateGuild) Channels() []*StateChannel {
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	channels := make([]*StateChannel, 0, len(sg.channels))
	for _, sc := range sg.channels {
		channels = append(channels, sc)
	}
	return channels
}

func (sg *StateGuild) Presences() []*StatePresence {
	sg.modelMu.RLock()
	defer sg.modelMu.RUnlock()
	presences := make([]*StatePresence, 0, len(sg.presences))
	for _, p := range sg.presences {
		presences = append(presences, p)
	}
	return presences
}

type StateChannel struct {
	mu       sync.RWMutex
	c        *ModelChannel
	guild    *StateGuild
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

type StateUser struct {
	mu sync.RWMutex
	u  *ModelUser
}

func (su *StateUser) ID() string {
	su.mu.RLock()
	defer su.mu.RUnlock()
	return su.u.ID
}

func (su *StateUser) Username() string {
	su.mu.RLock()
	defer su.mu.RUnlock()
	return su.u.Username
}

func (su *StateUser) Discriminator() string {
	su.mu.RLock()
	defer su.mu.RUnlock()
	return su.u.Discriminator
}

func (su *StateUser) Avatar() string {
	su.mu.RLock()
	defer su.mu.RUnlock()
	return su.u.Avatar
}

func (su *StateUser) Bot() bool {
	su.mu.RLock()
	defer su.mu.RUnlock()
	return su.u.Bot
}

func (su *StateUser) MFAEnabled() bool {
	su.mu.RLock()
	defer su.mu.RUnlock()
	return su.u.MFAEnabled
}

func (su *StateUser) Verified() bool {
	su.mu.RLock()
	defer su.mu.RUnlock()
	return su.u.Verified
}

func (su *StateUser) Email() string {
	su.mu.RLock()
	defer su.mu.RUnlock()
	return su.u.Email
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

	sg.modelMu.Lock()
	sg.channels = append(sg.channels, sc)
	sg.modelMu.Unlock()

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
			sg.modelMu.Lock()
			sg.channels = append(sg.channels[:i], sg.channels[i+1:]...)
			sg.modelMu.Unlock()
			return nil
		}
	}
	return errors.New("channel removed from a guild in which it never existed?")
}

var errHandled = errors.New("no need to handle the event further")

func (s *State) createGuild(ctx context.Context, conn *Conn, e *EventGuildCreate) error {
	sg := &StateGuild{
		g:           &e.ModelGuild,
		large:       e.Large,
		unavailable: e.Unavailable,
		memberCount: e.MemberCount,
		voiceStates: e.VoiceStates,
		members:     e.Members,
		presences:   e.Presences,
	}

	s.mu.Lock()
	for _, c := range e.Channels {
		sc := &StateChannel{
			c:     c,
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
	sg, ok := s.guilds[e.ID]
	if !ok {
		return errors.New("non existing guild updated?")
	}
	s.mu.Unlock()

	sg.modelMu.Lock()
	sg.g = &e.ModelGuild
	sg.modelMu.Unlock()
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
		sg.modelMu.Lock()
		sg.unavailable = true
		sg.modelMu.Lock()
	} else {
		delete(s.guilds, e.ID)
	}
	for _, sc := range sg.channels {
		// I don't think ID() helper is necessary, but cannot be too safe.
		delete(s.channels, sc.ID())
	}
	return nil
}

// TODO maybe guild ban add and guild ban remove? Not sure....

func (s *State) Guild(id string) (*StateGuild, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sg, ok := s.guilds[id]
	return sg, ok
}

func (s *State) updateGuildEmojis(ctx context.Context, conn *Conn, e *EventGuildEmojisUpdate) error {
	sg, ok := s.Guild(e.GuildID)
	if !ok {
		return errors.New("guild emojis updated for non existing guild")
	}
	sg.modelMu.Lock()
	sg.g.Emojis = e.Emojis
	sg.modelMu.Unlock()
	return nil
}

func (s *State) addGuildMember(ctx context.Context, conn *Conn, e *EventGuildMemberAdd) error {
	sg, ok := s.Guild(e.GuildID)
	if !ok {
		return errors.New("guild member added in non existing guild")
	}
	sg.modelMu.Lock()
	sg.memberCount++
	sg.members = append(sg.members, &e.ModelGuildMember)
	sg.modelMu.Unlock()
	return nil
}

func (s *State) removeGuildMember(ctx context.Context, conn *Conn, e *EventGuildMemberRemove) error {
	sg, ok := s.Guild(e.GuildID)
	if !ok {
		return errors.New("guild member removed in non existing guild")
	}
	for i, gm := range sg.Members() {
		if gm.User.ID == e.User.ID {
			sg.modelMu.Lock()
			sg.memberCount--
			sg.members = append(sg.members[:i], sg.members[i+1:]...)
			sg.modelMu.Unlock()
			return nil
		}
	}
	return errors.New("guild member removed in a guild in which it never joined?")
}

func (s *State) updateGuildMember(ctx context.Context, conn *Conn, e *EventGuildMemberUpdate) error {
	sg, ok := s.Guild(e.GuildID)
	if !ok {
		return errors.New("guild member updated in non existing guild")
	}
	for i, gm := range sg.Members() {
		if gm.User.ID == e.User.ID {
			gm2 := &ModelGuildMember{
				User:     gm.User,
				Roles:    e.Roles,
				JoinedAt: gm.JoinedAt,
				Deaf:     gm.Deaf,
				Mute:     gm.Mute,
				Nick:     gm.Nick,
			}
			sg.modelMu.Lock()
			sg.members[i] = gm2
			sg.modelMu.Unlock()
			return nil
		}
	}
	return errors.New("guild member updated in a guild in which it never joined?")
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
