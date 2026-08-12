package telegramapi

import (
	"encoding/json"
	"strings"
)

// Update is the subset of a Telegram update Butter routes on.
//
// It is deliberately partial: the raw JSON travels through the queue intact,
// and this struct exists only so the receive path can extract the address,
// the update ID, and enough of the message to recognize `/where`. Fields the
// runtime does not act on are omitted rather than modeled speculatively.
type Update struct {
	UpdateID int64 `json:"update_id"`

	Message       *Message_ `json:"message,omitempty"`
	EditedMessage *Message_ `json:"edited_message,omitempty"`
	ChannelPost   *Message_ `json:"channel_post,omitempty"`
	CallbackQuery *Callback `json:"callback_query,omitempty"`
}

// Message_ is one Telegram message. The trailing underscore avoids colliding
// with the outbound Message type in this package.
type Message_ struct {
	MessageID       int64           `json:"message_id"`
	MessageThreadID int64           `json:"message_thread_id,omitempty"`
	From            *User           `json:"from,omitempty"`
	SenderChat      *Chat           `json:"sender_chat,omitempty"`
	Chat            *Chat           `json:"chat,omitempty"`
	Text            string          `json:"text,omitempty"`
	Caption         string          `json:"caption,omitempty"`
	Entities        []MessageEntity `json:"entities,omitempty"`
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	Photo           []PhotoSize     `json:"photo,omitempty"`

	// IsTopicMessage distinguishes a real Forum Topic message from the
	// group's general conversation.
	IsTopicMessage bool `json:"is_topic_message,omitempty"`
	// IsAutomaticForward marks a channel post mirrored into a discussion
	// group; Butter never treats those as user input.
	IsAutomaticForward bool `json:"is_automatic_forward,omitempty"`
	// NewChatMembers and friends mark service messages.
	NewChatMembers []User    `json:"new_chat_members,omitempty"`
	LeftChatMember *User     `json:"left_chat_member,omitempty"`
	PinnedMessage  *Message_ `json:"pinned_message,omitempty"`
	MediaGroupID   string    `json:"media_group_id,omitempty"`
}

// User is a Telegram user or bot.
type User struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
}

// Chat is the conversation a message belongs to.
type Chat struct {
	ID       int64  `json:"id"`
	Type     string `json:"type,omitempty"`
	Title    string `json:"title,omitempty"`
	Username string `json:"username,omitempty"`
	IsForum  bool   `json:"is_forum,omitempty"`
}

// MessageEntity marks a span of the text, e.g. a command or a mention.
type MessageEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	User   *User  `json:"user,omitempty"`
}

// PhotoSize is one rendition of an uploaded photo.
type PhotoSize struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FileSize     int64  `json:"file_size,omitempty"`
}

// Callback is an inline keyboard press.
type Callback struct {
	ID      string    `json:"id"`
	From    *User     `json:"from,omitempty"`
	Message *Message_ `json:"message,omitempty"`
	Data    string    `json:"data,omitempty"`
}

// ParseUpdate decodes the routable subset of a raw Telegram update.
func ParseUpdate(raw []byte) (*Update, error) {
	update := &Update{}
	if err := json.Unmarshal(raw, update); err != nil {
		return nil, err
	}
	return update, nil
}

// RoutableMessage returns the message an update should be addressed by, and
// whether the update carries one at all.
//
// Edited messages, channel posts, and automatic forwards deliberately return
// false: they are addressable but are not user input, and treating them as
// such would re-run agents on history edits.
func (u *Update) RoutableMessage() (*Message_, bool) {
	switch {
	case u == nil:
		return nil, false
	case u.Message != nil:
		return u.Message, true
	case u.CallbackQuery != nil && u.CallbackQuery.Message != nil:
		return u.CallbackQuery.Message, true
	default:
		return nil, false
	}
}

// AddressOf returns the canonical chat ID and optional thread ID of a
// message. A message outside a Forum Topic yields an empty thread ID, which
// is a distinct address from any real topic.
func AddressOf(msg *Message_) (chatID, threadID string) {
	if msg == nil || msg.Chat == nil {
		return "", ""
	}
	chatID = FormatID(msg.Chat.ID)
	// message_thread_id is also set on plain replies inside a non-forum
	// group, so it only names a Topic when Telegram says so.
	if msg.IsTopicMessage && msg.MessageThreadID > 0 {
		threadID = FormatID(msg.MessageThreadID)
	}
	return chatID, threadID
}

// Command returns the bot command a message starts with, without the leading
// slash and without any `@botname` suffix, plus the remaining argument text.
//
// Telegram marks commands with an entity rather than by text convention, so
// matching on the entity is what keeps a message that merely mentions "/where"
// mid-sentence from being treated as one.
func Command(msg *Message_) (command, args string, ok bool) {
	if msg == nil {
		return "", "", false
	}
	text := msg.Text
	entities := msg.Entities
	if text == "" {
		text, entities = msg.Caption, msg.CaptionEntities
	}
	if text == "" {
		return "", "", false
	}
	runes := []rune(text)
	for _, entity := range entities {
		if entity.Type != "bot_command" || entity.Offset != 0 {
			continue
		}
		end := min(entity.Offset+entity.Length, len(runes))
		raw := strings.TrimPrefix(string(runes[entity.Offset:end]), "/")
		// `/command@botname` addresses one bot in a multi-bot group.
		if name, _, hasSuffix := strings.Cut(raw, "@"); hasSuffix {
			raw = name
		}
		return strings.ToLower(raw), strings.TrimSpace(string(runes[end:])), true
	}
	return "", "", false
}

// CommandTargetsBot reports whether a `/command@botname` suffix addresses
// this bot. A bare command with no suffix targets every bot in the chat.
func CommandTargetsBot(msg *Message_, botUsername string) bool {
	if msg == nil {
		return false
	}
	text := msg.Text
	entities := msg.Entities
	if text == "" {
		text, entities = msg.Caption, msg.CaptionEntities
	}
	runes := []rune(text)
	for _, entity := range entities {
		if entity.Type != "bot_command" || entity.Offset != 0 {
			continue
		}
		end := min(entity.Offset+entity.Length, len(runes))
		_, target, hasSuffix := strings.Cut(string(runes[entity.Offset:end]), "@")
		if !hasSuffix {
			// A bare command addresses every bot in the chat.
			return true
		}
		return strings.EqualFold(target, botUsername)
	}
	return false
}
