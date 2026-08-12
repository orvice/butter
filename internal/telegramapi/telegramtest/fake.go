// Package telegramtest provides an in-process stand-in for the Telegram Bot
// API (issue #264). Tests wire it in through telegramapi.Factory so the whole
// service layer — credential validation, rotation, enablement preflight — runs
// its real code paths without a live Bot.
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

// Fake resolves tokens to Bots. Unknown tokens are rejected exactly as
// Telegram rejects them: HTTP 401, which telegramapi maps to ErrUnauthorized.
type Fake struct {
	mu   sync.Mutex
	bots map[string]Bot
	// GetMeCalls counts validation attempts so tests can assert that a
	// credential really was checked before anything was committed.
	getMeCalls int
}

func NewFake() *Fake {
	return &Fake{bots: make(map[string]Bot)}
}

// Register makes token resolve to the given Bot.
func (f *Fake) Register(token string, identity telegramapi.BotIdentity) *Fake {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.bots[token] = Bot{Identity: identity}
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

type fakeClient struct {
	fake  *Fake
	token string
}

func (c *fakeClient) GetMe(context.Context) (telegramapi.BotIdentity, error) {
	c.fake.mu.Lock()
	defer c.fake.mu.Unlock()
	c.fake.getMeCalls++
	bot, ok := c.fake.bots[c.token]
	if !ok {
		return telegramapi.BotIdentity{}, &telegramapi.APIError{
			Code:        http.StatusUnauthorized,
			Description: "Unauthorized",
		}
	}
	return bot.Identity, nil
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
