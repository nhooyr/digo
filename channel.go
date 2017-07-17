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
)

// Channel represents a channel in Discord.
// If IsPrivate is set this is a DM Channel otherwise this is a Guild Channel.
type Channel struct {
	ID                   string                 `json:"id"`
	GuildID              string                 `json:"guild_id"`
	Name                 string                 `json:"name"`
	Type                 string                 `json:"type"`
	Position             int                    `json:"position"`
	IsPrivate            bool                   `json:"is_private"`
	Recipient            *User                  `json:"recipient"`
	PermissionOverwrites []*PermissionOverwrite `json:"permission_overwrites"`
	Topic                string                 `json:"topic"`
	LastMessageID        string                 `json:"last_message_id"`
	Bitrate              int                    `json:"bitrate"`
	UserLimit            int                    `json:"user_limit"`
}

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
	Reactions       []*Reaction   `json:"reactions"`
	Nonce           string        `json:"nonce"`
	Pinned          bool          `json:"pinned"`
	WebhookID       string        `json:"webhook_id"`
}

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

type ChannelEndpoint struct {
	*endpoint
}

func (c *Client) Channel(cID string) *ChannelEndpoint {
	e2 := c.e.appendMajor("channels").appendMajor(cID)
	return &ChannelEndpoint{e2}
}

func (e *ChannelEndpoint) Get() (ch *Channel, err error) {
	return ch, e.doMethod("GET", nil, &ch)
}

type ChannelModifyParams struct {
	Name      string `json:"name,omitempty"`
	Position  int    `json:"position,omitempty"`
	Topic     string `json:"topic,omitempty"`
	Bitrate   int    `json:"bitrate,omitempty"`
	UserLimit int    `json:"user_limit,omitempty"`
}

func (e *ChannelEndpoint) Modify(params *ChannelModifyParams) (ch *Channel, err error) {
	return ch, e.doMethod("PATCH", params, &ch)
}

func (e *ChannelEndpoint) Delete() (ch *Channel, err error) {
	return ch, e.doMethod("DELETE", nil, &ch)
}

type MessagesEndpoint struct {
	*endpoint
}

func (e *ChannelEndpoint) Messages() *MessagesEndpoint {
	return &MessagesEndpoint{e.appendMajor("messages")}
}

func (e *MessagesEndpoint) BulkDelete(messageIDs []string) error {
	e2 := e.appendMajor("bulk-delete")
	return e2.doMethod("POST", messageIDs, nil)
}

type GetMessagesParams struct {
	AroundID string
	BeforeID string
	AfterID  string
	Limit    int
}

func (params *GetMessagesParams) rawQuery() string {
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

func (e *MessagesEndpoint) Get(params *GetMessagesParams) (messages []*Message, err error) {
	req := e.newRequest("GET", nil)
	if params != nil {
		req.URL.RawQuery = params.rawQuery()
	}
	return messages, e.do(req, &messages)
}

type CreateMessageParams struct {
	Content string `json:"content,omitempty"`
	Nonce   string `json:"nonce,omitempty"`
	TTS     bool   `json:"tts,omitempty"`
	File    *File  `json:"-"`
	Embed   *Embed `json:"embed,omitempty"`
}

type File struct {
	Name    string
	Content io.Reader
}

func (e *MessagesEndpoint) Create(params *CreateMessageParams) (m *Message, err error) {
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

type MessageEndpoint struct {
	*endpoint
}

func (e *ChannelEndpoint) Message(mID string) *MessageEndpoint {
	e2 := e.Messages().appendMinor(mID)
	return &MessageEndpoint{e2}
}

func (e *MessageEndpoint) Get() (m *Message, err error) {
	return m, e.doMethod("GET", nil, &m)
}

type MessageEditParams struct {
	Content string `json:"content,omitempty"`
	Embed   *Embed `json:"embed,omitempty"`
}

func (e *MessageEndpoint) Patch(params *MessageEditParams) (m *Message, err error) {
	return m, e.doMethod("PATCH", params, &m)
}

func (e *MessageEndpoint) Delete() error {
	return e.doMethod("DELETE", nil, nil)
}

type ReactionsEndpoint struct {
	*endpoint
}

func (e *MessageEndpoint) Reactions() *ReactionsEndpoint {
	return &ReactionsEndpoint{e.appendMajor("reactions")}
}

func (e *ReactionsEndpoint) Delete() error {
	return e.doMethod("DELETE", nil, nil)
}

func (e *ReactionsEndpoint) Get(emoji string) (users []*User, err error) {
	e2 := e.appendMinor(emoji)
	return users, e2.doMethod("GET", nil, &users)
}

func (e *ReactionsEndpoint) Create(emoji string) error {
	e2 := e.appendMinor(emoji).appendMinor("@me")
	return e2.doMethod("PUT", nil, nil)
}

type ReactionEndpoint struct {
	*endpoint
}

func (e *MessageEndpoint) Reaction(emoji, uID string) *ReactionEndpoint {
	e2 := e.Reactions().appendMinor(emoji).appendMinor(uID)
	return &ReactionEndpoint{e2}
}

func (e *ReactionEndpoint) Delete() error {
	return e.doMethod("DELETE", nil, nil)
}

type PermissionOverwriteEndpoint struct {
	*endpoint
}

func (e *ChannelEndpoint) PermissionOverwrite(overwriteID string) *PermissionOverwriteEndpoint {
	e2 := e.appendMajor("permissions").appendMinor(overwriteID)
	return &PermissionOverwriteEndpoint{e2}
}

type PermissionOverwriteEditParams struct {
	Allow int    `json:"allow"`
	Deny  int    `json:"deny"`
	Type  string `json:"type"`
}

func (e *PermissionOverwriteEndpoint) Edit(params *PermissionOverwriteEditParams) error {
	return e.doMethod("PUT", params, nil)
}

func (e *PermissionOverwriteEndpoint) Delete() error {
	return e.doMethod("DELETE", nil, nil)
}

type InvitesEndpoint struct {
	*endpoint
}

func (e *ChannelEndpoint) Invites() *InvitesEndpoint {
	e2 := e.appendMajor("invites")
	return &InvitesEndpoint{e2}
}

func (e *InvitesEndpoint) Get() (invites []*Invite, err error) {
	return invites, e.doMethod("GET", nil, &invites)
}

type InviteCreateParams struct {
	MaxAge    int  `json:"max_age,omitempty"`
	MaxUses   int  `json:"max_uses,omitempty"`
	Temporary bool `json:"temporary,omitempty"`
	Unique    bool `json:"unique,omitempty"`
}

func (e *InvitesEndpoint) Create(params *InviteCreateParams) (invite *Invite, err error) {
	return invite, e.doMethod("POST", params, &invite)
}

func (e *ChannelEndpoint) TriggerTypingIndicator() error {
	e2 := e.appendMajor("typing")
	return e2.doMethod("POST", nil, nil)
}

type PinsEndpoint struct {
	*endpoint
}

func (e *ChannelEndpoint) Pins() *PinsEndpoint {
	e2 := e.appendMajor("pins")
	return &PinsEndpoint{e2}
}

func (e *PinsEndpoint) Get() (messages []*Message, err error) {
	return messages, e.doMethod("GET", nil, &messages)
}

type PinEndpoint struct {
	*endpoint
}

func (e *ChannelEndpoint) Pin(mID string) *PinEndpoint {
	e2 := e.Pins().appendMinor(mID)
	return &PinEndpoint{e2}
}

func (e *PinEndpoint) Add() error {
	return e.doMethod("PUT", nil, nil)
}

func (e *PinEndpoint) Delete() error {
	return e.doMethod("DELETE", nil, nil)
}

type RecipientEndpoint struct {
	*endpoint
}

func (e *ChannelEndpoint) Recipient(uID string) *RecipientEndpoint {
	e2 := e.appendMajor("recipients").appendMinor(uID)
	return &RecipientEndpoint{e2}
}

func (e *RecipientEndpoint) Add() error {
	return e.doMethod("PUT", nil, nil)
}

func (e *RecipientEndpoint) Delete() error {
	return e.doMethod("DELETE", nil, nil)
}