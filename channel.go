package discgo

import (
	"time"

	"encoding/json"

	"net/url"
	"strconv"

	"bytes"
	"fmt"
	"io"
	"mime/multipart"

	"gopkg.in/guregu/null.v3"
)

// Channel represents a channel in Discord.
type Channel struct {
	ID   string      `json:"id"`
	Type channelType `json:"type"`

	// All of these may be null. They are pointers so that consumer's
	// of the package will be reminded that they may be null.
	GuildID              *string                 `json:"guild_id"`
	Position             *int                    `json:"position"`
	PermissionOverwrites *[]*PermissionOverwrite `json:"permission_overwrites"`
	Name                 *string                 `json:"name"`
	Topic                *string                 `json:"topic"`
	LastMessageID        *string                 `json:"last_message_id"`
	Bitrate              *int                    `json:"bitrate"`
	UserLimit            *int                    `json:"user_limit"`
	Recipients           *[]*User                `json:"recipients"`
	Icon                 *string                 `json:"icon"`
	OwnerID              *string                 `json:"owner_id"`
	ApplicationID        *string                 `json:"application_id"`
}

type channelType int

const (
	GuildTextChannel = iota
	DMChannel
	GuildVoiceChannel
	GroupDMChannel
	GuildCategoryChannel
)

type Message struct {
	ID              string        `json:"id"`
	ChannelID       string        `json:"channel_id"`
	Author          *User         `json:"author"`
	Content         string        `json:"content"`
	Timestamp       time.Time     `json:"timestamp"`
	EditedTimestamp *time.Time    `json:"edited_timestamp"`
	TTS             bool          `json:"tts"`
	MentionEveryone bool          `json:"mention_everyone"`
	Mentions        []*User       `json:"mentions"`
	MentionRoles    []string      `json:"mention_roles"`
	Attachments     []*Attachment `json:"attachments"`
	Embeds          []*Embed      `json:"embeds"`
	Reactions       *[]*Reaction  `json:"reactions"`
	Nonce           *string       `json:"nonce"`
	Pinned          bool          `json:"pinned"`
	WebhookID       *string       `json:"webhook_id"`
	Type            messageType   `json:"type"`
}

type messageType int

const (
	DefaultMessage = iota
	RecipientAdd
	RecipientRemove
	Call
	ChannelNameChange
	ChannelIconChange
	ChannelPinnedMessage
	GuildMemberJoin
)

type Reaction struct {
	Count int            `json:"count"`
	Me    bool           `json:"me"`
	Emoji *ReactionEmoji `json:"emoji"`
}

type ReactionEmoji struct {
	ID   *string `json:"id"`
	Name string  `json:"name"`
}

type PermissionOverwrite struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Allow int    `json:"allow"`
	Deny  int    `json:"deny"`
}

type Embed struct {
	Title       string          `json:"title,omitempty"`
	Type        string          `json:"type,omitempty"`
	Description string          `json:"description,omitempty"`
	URL         string          `json:"url,omitempty"`
	Timestamp   *time.Time      `json:"timestamp,omitempty"`
	Color       int             `json:"color,omitempty"`
	Footer      *EmbedFooter    `json:"footer,omitempty"`
	Image       *EmbedImage     `json:"image,omitempty"`
	Thumbnail   *EmbedThumbnail `json:"thumbnail,omitempty"`
	Video       *EmbedVideo     `json:"video,omitempty"`
	Provider    *EmbedProvider  `json:"provider,omitempty"`
	Author      *EmbedAuthor    `json:"author,omitempty"`
	Fields      []*EmbedField   `json:"fields,omitempty"`
}

type EmbedThumbnail struct {
	URL      string `json:"url,omitempty"`
	ProxyURL string `json:"proxy_url,omitempty"`
	Height   int    `json:"height,omitempty"`
	Width    int    `json:"width,omitempty"`
}

type EmbedVideo struct {
	URL    string `json:"url,omitempty"`
	Height int    `json:"height,omitempty"`
	Width  int    `json:"width,omitempty"`
}

type EmbedImage struct {
	URL      string `json:"url,omitempty"`
	ProxyURL string `json:"proxy_url,omitempty"`
	Height   int    `json:"height,omitempty"`
	Width    int    `json:"width,omitempty"`
}

type EmbedProvider struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

type EmbedAuthor struct {
	Name         string `json:"name,omitempty"`
	URL          string `json:"url,omitempty"`
	IconURL      string `json:"icon_url,omitempty"`
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}

type EmbedFooter struct {
	Text         string `json:"text,omitempty"`
	IconURL      string `json:"icon_url,omitempty"`
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}

type EmbedField struct {
	Name   string `json:"name,omitempty"`
	Value  string `json:"value,omitempty"`
	Inline bool   `json:"inline,omitempty"`
}

type Attachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Size     int    `json:"size"`
	URL      string `json:"url"`
	ProxyURL string `json:"proxy_url"`
	Height   int    `json:"height"`
	Width    int    `json:"width"`
}

func UserMention(uID string) string {
	return fmt.Sprintf("<@%v>", uID)
}

func UserNicknameMention(uID string) string {
	return fmt.Sprintf("<@!%v>", uID)
}

func ChannelMention(cID string) string {
	return fmt.Sprintf("<#%v>", cID)
}

func RoleMention(roleID string) string {
	return fmt.Sprintf("<@&%v>", roleID)
}

func CustomEmojiMessage(emojiName, emojiID string) string {
	return fmt.Sprintf("<:%v:%v>", emojiName, emojiID)
}

type EndpointChannel struct {
	*endpoint
}

func (c *Client) Channel(cID string) EndpointChannel {
	e2 := c.e.appendMajor("channels").appendMajor(cID)
	return EndpointChannel{e2}
}

func (e EndpointChannel) Get() (ch *Channel, err error) {
	return ch, e.doMethod("GET", nil, &ch)
}

type ParamsChannelModify struct {
	Name      string      `json:"name,omitempty"`
	Position  int         `json:"position,omitempty"`
	Topic     null.String `json:"topic"`
	Bitrate   int         `json:"bitrate,omitempty"`
	UserLimit null.Int    `json:"user_limit"`
}

func (e EndpointChannel) Modify(params *ParamsChannelModify) (ch *Channel, err error) {
	return ch, e.doMethod("PATCH", params, &ch)
}

func (e EndpointChannel) Delete() (ch *Channel, err error) {
	return ch, e.doMethod("DELETE", nil, &ch)
}

type EndpointMessages struct {
	*endpoint
}

func (e EndpointChannel) Messages() EndpointMessages {
	return EndpointMessages{e.appendMajor("messages")}
}

type ParamsMessagesBulkDelete struct {
	Messages []string `json:"messages"`
}

func (e EndpointMessages) BulkDelete(params *ParamsMessagesBulkDelete) error {
	e2 := e.appendMajor("bulk-delete")
	return e2.doMethod("POST", params, nil)
}

type ParamsMessagesGet struct {
	AroundID string
	BeforeID string
	AfterID  string
	Limit    int
}

func (params *ParamsMessagesGet) rawQuery() string {
	v := make(url.Values)
	if params.AroundID != "" {
		v.Set("around", params.AroundID)
	}
	if params.BeforeID != "" {
		v.Set("before", params.BeforeID)
	}
	if params.AfterID != "" {
		v.Set("after", params.AfterID)
	}
	if params.Limit > 0 {
		v.Set("limit", strconv.Itoa(params.Limit))
	}
	return v.Encode()
}

func (e EndpointMessages) Get(params *ParamsMessagesGet) (messages []*Message, err error) {
	req := e.newRequest("GET", nil)
	if params != nil {
		req.URL.RawQuery = params.rawQuery()
	}
	return messages, e.do(req, &messages)
}

type ParamsMessageCreate struct {
	Content string `json:"content,omitempty"`
	Nonce   string `json:"nonce,omitempty"`
	TTS     bool   `json:"tts,omitempty"`
	File    *File  `json:"-"`
	Embed   *Embed `json:"embed,omitempty"`
}

// TODO perhaps make an interface? like http.File?
type File struct {
	Name    string
	Content io.Reader
}

func (e EndpointMessages) Create(params *ParamsMessageCreate) (m *Message, err error) {
	reqBody := &bytes.Buffer{}
	reqBodyWriter := multipart.NewWriter(reqBody)

	payloadJSON, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	w, err := reqBodyWriter.CreateFormField("payload_json")
	if err != nil {
		return nil, err
	}
	_, err = w.Write(payloadJSON)
	if err != nil {
		return nil, err
	}

	if params.File != nil {
		w, err := reqBodyWriter.CreateFormFile("file", params.File.Name)
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(w, params.File.Content)
		if err != nil {
			return nil, err
		}
	}

	err = reqBodyWriter.Close()
	if err != nil {
		return nil, err
	}

	req := e.newRequest("POST", reqBody)
	req.Header.Set("Content-Type", reqBodyWriter.FormDataContentType())
	return m, e.do(req, &m)
}

type EndpointMessage struct {
	*endpoint
}

func (e EndpointChannel) Message(mID string) EndpointMessage {
	e2 := e.Messages().appendMinor(mID)
	return EndpointMessage{e2}
}

func (e EndpointMessage) Get() (m *Message, err error) {
	return m, e.doMethod("GET", nil, &m)
}

type ParamsMessageEdit struct {
	// TODO should I allow setting the content to ""?
	Content string `json:"content,omitempty"`
	Embed   *Embed `json:"embed,omitempty"`
}

func (e EndpointMessage) Edit(params *ParamsMessageEdit) (m *Message, err error) {
	return m, e.doMethod("PATCH", params, &m)
}

func (e EndpointMessage) Delete() error {
	return e.doMethod("DELETE", nil, nil)
}

// TODO not a fan of this API design, revisit later maybe
type EndpointReactions struct {
	*endpoint
}

func (e EndpointMessage) Reactions() EndpointReactions {
	return EndpointReactions{e.appendMajor("reactions")}
}

func (e EndpointReactions) Delete() error {
	return e.doMethod("DELETE", nil, nil)
}

func (e EndpointReactions) Get(emoji string) (users []*User, err error) {
	e2 := e.appendMinor(emoji)
	return users, e2.doMethod("GET", nil, &users)
}

func (e EndpointReactions) Create(emoji string) error {
	e2 := e.appendMinor(emoji).appendMinor("@me")
	return e2.doMethod("PUT", nil, nil)
}

type EndpointReaction struct {
	*endpoint
}

// uID = @me to delete your reaction.
func (e EndpointMessage) Reaction(emoji, uID string) EndpointReaction {
	e2 := e.Reactions().appendMinor(emoji).appendMinor(uID)
	return EndpointReaction{e2}
}

func (e EndpointReaction) Delete() error {
	return e.doMethod("DELETE", nil, nil)
}

type EndpointPermissionOverwrite struct {
	*endpoint
}

func (e EndpointChannel) PermissionOverwrite(overwriteID string) EndpointPermissionOverwrite {
	e2 := e.appendMajor("permissions").appendMinor(overwriteID)
	return EndpointPermissionOverwrite{e2}
}

type ParamsPermissionOverwriteEdit struct {
	Allow int    `json:"allow"`
	Deny  int    `json:"deny"`
	Type  string `json:"type"`
}

func (e EndpointPermissionOverwrite) Edit(params *ParamsPermissionOverwriteEdit) error {
	return e.doMethod("PUT", params, nil)
}

func (e EndpointPermissionOverwrite) Delete() error {
	return e.doMethod("DELETE", nil, nil)
}

// TODO move somewhere where it can be shared between guild.go and channel.go
type EndpointInvites struct {
	*endpoint
}

func (e EndpointChannel) Invites() EndpointInvites {
	e2 := e.appendMajor("invites")
	return EndpointInvites{e2}
}

func (e EndpointInvites) Get() (invites []*Invite, err error) {
	return invites, e.doMethod("GET", nil, &invites)
}

type ParamsInviteCreate struct {
	MaxAge    null.Int `json:"max_age"`
	MaxUses   null.Int `json:"max_uses"`
	Temporary bool     `json:"temporary,omitempty"`
	Unique    bool     `json:"unique,omitempty"`
}

func (e EndpointInvites) Create(params *ParamsInviteCreate) (invite *Invite, err error) {
	return invite, e.doMethod("POST", params, &invite)
}

type EndpointTypingIndicator struct {
	*endpoint
}

func (e EndpointChannel) TypingIndicator() EndpointTypingIndicator {
	e2 := e.appendMajor("typing")
	return EndpointTypingIndicator{e2}
}

func (e EndpointTypingIndicator) Trigger() error {
	return e.doMethod("POST", nil, nil)
}

type EndpointPins struct {
	*endpoint
}

func (e EndpointChannel) Pins() EndpointPins {
	e2 := e.appendMajor("pins")
	return EndpointPins{e2}
}

func (e EndpointPins) Get() (messages []*Message, err error) {
	return messages, e.doMethod("GET", nil, &messages)
}

type EndpointPin struct {
	*endpoint
}

func (e EndpointChannel) Pin(mID string) EndpointPin {
	e2 := e.Pins().appendMinor(mID)
	return EndpointPin{e2}
}

func (e EndpointPin) Add() error {
	return e.doMethod("PUT", nil, nil)
}

func (e EndpointPin) Delete() error {
	return e.doMethod("DELETE", nil, nil)
}

type EndpointRecipient struct {
	*endpoint
}

func (e EndpointChannel) Recipient(uID string) EndpointRecipient {
	e2 := e.appendMajor("recipients").appendMinor(uID)
	return EndpointRecipient{e2}
}

type ParamsRecipientAdd struct {
	AccessToken string `json:"access_token"`
	Nick        string `json:"nick"`
}

func (e EndpointRecipient) Add(params *ParamsRecipientAdd) error {
	return e.doMethod("PUT", params, nil)
}

func (e EndpointRecipient) Delete() error {
	return e.doMethod("DELETE", nil, nil)
}
