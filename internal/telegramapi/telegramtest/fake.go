// Package telegramtest provides an in-process stand-in for the Telegram Bot
// API (issue #264). Tests wire it in through telegramapi.Factory so the whole
// service layer — credential validation, rotation, enablement preflight,
// delivery — runs its real code paths without a live Bot.
package telegramtest

import (
	"context"
	"net/http"
	"sync"

	"go.orx.me/apps/butter/internal/telegramapi"
)

// Bot is one fake Telegram Bot keyed by its token.
type Bot struct {
	Identity telegramapi.BotIdentity
}

// Sent records one delivered message together with the token that sent it,
// so tests can assert both the payload and which credential was used.
type Sent struct {
	Token  string
	Params telegramapi.SendMessageParams
	Edit   *telegramapi.EditMessageParams
}

// SendHook lets a test fail or delay a specific send. Returning a non-nil
// error makes the call fail; the fake still records the attempt.
type SendHook func(attempt int, params telegramapi.SendMessageParams) error

// Fake resolves tokens to Bots. Unknown tokens are rejected exactly as
// Telegram rejects them: HTTP 401, which telegramapi maps to ErrUnauthorized.
type Fake struct {
	mu   sync.Mutex
	bots map[string]Bot
	// getMeCalls counts validation attempts so tests can assert that a
	// credential really was checked before anything was committed.
	getMeCalls int
	sent       []Sent
	nextID     int64
	sendHook   SendHook
}

func NewFake() *Fake {
	return &Fake{bots: make(map[string]Bot), nextID: 1000}
}

// Register makes token resolve to the given Bot.
func (f *Fake) Register(token string, identity telegramapi.BotIdentity) *Fake {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.bots[token] = Bot{Identity: identity}
	return f
}

// OnSend installs a hook consulted before each SendMessage succeeds.
func (f *Fake) OnSend(hook SendHook) *Fake {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sendHook = hook
	return f
}

// Factory returns a telegramapi.Factory backed by this fake.
func (f *Fake) Factory() telegramapi.Factory {
	return func(token string) telegramapi.Client { return &fakeClient{fake: f, token: token} }
}

// GetMeCalls reports how many times a token was validated.
func (f *Fake) GetMeCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.getMeCalls
}

// Sent returns every delivery attempt in order.
func (f *Fake) Sent() []Sent {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]Sent(nil), f.sent...)
}

// LastSent returns the most recent delivery attempt, or false when none.
func (f *Fake) LastSent() (Sent, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.sent) == 0 {
		return Sent{}, false
	}
	return f.sent[len(f.sent)-1], true
}

// Reset clears recorded sends without forgetting registered bots.
func (f *Fake) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = nil
}

type fakeClient struct {
	fake  *Fake
	token string
}

func unauthorized() error {
	return &telegramapi.APIError{Code: http.StatusUnauthorized, Description: "Unauthorized"}
}

func (c *fakeClient) GetMe(context.Context) (telegramapi.BotIdentity, error) {
	c.fake.mu.Lock()
	defer c.fake.mu.Unlock()
	c.fake.getMeCalls++
	bot, ok := c.fake.bots[c.token]
	if !ok {
		return telegramapi.BotIdentity{}, unauthorized()
	}
	return bot.Identity, nil
}

func (c *fakeClient) SendMessage(_ context.Context, params telegramapi.SendMessageParams) (telegramapi.Message, error) {
	c.fake.mu.Lock()
	defer c.fake.mu.Unlock()
	if _, ok := c.fake.bots[c.token]; !ok {
		return telegramapi.Message{}, unauthorized()
	}
	c.fake.sent = append(c.fake.sent, Sent{Token: c.token, Params: params})
	if c.fake.sendHook != nil {
		if err := c.fake.sendHook(len(c.fake.sent), params); err != nil {
			return telegramapi.Message{}, err
		}
	}
	c.fake.nextID++
	return telegramapi.Message{ID: c.fake.nextID}, nil
}

func (c *fakeClient) EditMessageText(_ context.Context, params telegramapi.EditMessageParams) (telegramapi.Message, error) {
	c.fake.mu.Lock()
	defer c.fake.mu.Unlock()
	if _, ok := c.fake.bots[c.token]; !ok {
		return telegramapi.Message{}, unauthorized()
	}
	edit := params
	c.fake.sent = append(c.fake.sent, Sent{Token: c.token, Edit: &edit})
	c.fake.nextID++
	return telegramapi.Message{ID: c.fake.nextID}, nil
}

// Identity builds a BotIdentity with the group-friendly defaults most tests
// want: it can join groups and Group Privacy is off.
func Identity(id int64, username string) telegramapi.BotIdentity {
	return telegramapi.BotIdentity{
		ID:                      id,
		Username:                username,
		FirstName:               username,
		CanJoinGroups:           true,
		CanReadAllGroupMessages: true,
	}
}

// MarkdownRejection is the error Telegram returns when it refuses the markup
// rather than the request, which is what drives the plain-text fallback.
func MarkdownRejection() error {
	return &telegramapi.APIError{
		Code:        http.StatusBadRequest,
		Description: "Bad Request: can't parse entities: Character '-' is reserved",
	}
}

// RateLimited is Telegram's 429 with a retry_after hint.
func RateLimited(seconds int) error {
	return &telegramapi.APIError{
		Code:        http.StatusTooManyRequests,
		Description: "Too Many Requests: retry later",
		RetryAfter:  seconds,
	}
}
