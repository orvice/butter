// Package memory implements telegram.Repository in process memory for tests
// and single-process development. It enforces the same uniqueness, revision,
// and immutability rules as the Mongo store — those rules are the contract
// the service layer relies on, so a memory store that skipped them would let
// tests pass on writes production rejects.
package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type channelRecord struct {
	spec          *agentsv1.TelegramChannel
	cred          telegramrepo.Credential
	credUpdatedAt time.Time
	webhookSecret telegramrepo.Credential
}

type destinationRecord struct {
	spec *agentsv1.TelegramDestination
}

// Store implements telegram.Repository in memory.
type Store struct {
	mu           sync.RWMutex
	channels     map[string]*channelRecord
	destinations map[string]*destinationRecord
}

var _ telegramrepo.Repository = (*Store)(nil)

func New() *Store {
	return &Store{
		channels:     make(map[string]*channelRecord),
		destinations: make(map[string]*destinationRecord),
	}
}

func (s *Store) EnsureIndexes(context.Context) error { return nil }

// --- Channels --------------------------------------------------------------

func (s *Store) decodeChannel(rec *channelRecord) *agentsv1.TelegramChannel {
	out := proto.Clone(rec.spec).(*agentsv1.TelegramChannel)
	if !rec.cred.Set() {
		out.CredentialState = agentsv1.TelegramCredentialState_TELEGRAM_CREDENTIAL_STATE_MISSING
		out.CredentialUpdatedAt = nil
	} else if !rec.credUpdatedAt.IsZero() {
		out.CredentialUpdatedAt = timestamppb.New(rec.credUpdatedAt)
	}
	out.WebhookSecretSet = rec.webhookSecret.Set()
	return out
}

func (s *Store) ListChannels(_ context.Context, workspaceID string) ([]*agentsv1.TelegramChannel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*agentsv1.TelegramChannel
	for _, rec := range s.channels {
		if rec.spec.GetWorkspaceId() != workspaceID {
			continue
		}
		out = append(out, s.decodeChannel(rec))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].GetKey() < out[j].GetKey() })
	return out, nil
}

func (s *Store) GetChannel(_ context.Context, workspaceID, id string) (*agentsv1.TelegramChannel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.channels[id]
	if !ok || rec.spec.GetWorkspaceId() != workspaceID {
		return nil, channelNotFound(id)
	}
	return s.decodeChannel(rec), nil
}

func (s *Store) FindChannel(_ context.Context, id string) (*agentsv1.TelegramChannel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.channels[id]
	if !ok {
		return nil, channelNotFound(id)
	}
	return s.decodeChannel(rec), nil
}

func (s *Store) CreateChannel(_ context.Context, workspaceID string, channel *agentsv1.TelegramChannel, credentials telegramrepo.ChannelCredentials) (*agentsv1.TelegramChannel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, rec := range s.channels {
		if rec.spec.GetWorkspaceId() == workspaceID && rec.spec.GetKey() == channel.GetKey() {
			return nil, fmt.Errorf("telegram channel %q: %w", channel.GetKey(), telegramrepo.ErrKeyExists)
		}
		if channel.GetBotId() != "" && rec.spec.GetBotId() == channel.GetBotId() {
			return nil, fmt.Errorf("telegram bot %q: %w", channel.GetBotId(), telegramrepo.ErrBotExists)
		}
	}

	stored := proto.Clone(channel).(*agentsv1.TelegramChannel)
	stored.WorkspaceId = workspaceID
	now := timestamppb.New(time.Now().UTC())
	stored.CreatedAt = now
	stored.UpdatedAt = now
	stored.Revision = 1

	rec := &channelRecord{
		spec:          stored,
		cred:          credentials.BotToken,
		webhookSecret: credentials.WebhookSecret,
	}
	if credentials.BotToken.Set() {
		rec.credUpdatedAt = time.Now().UTC()
	}
	s.channels[stored.GetId()] = rec
	return s.decodeChannel(rec), nil
}

func (s *Store) UpdateChannel(_ context.Context, workspaceID string, channel *agentsv1.TelegramChannel, expectedRevision int64) (*agentsv1.TelegramChannel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.updateChannelLocked(workspaceID, channel, expectedRevision, nil)
}

func (s *Store) RotateChannelCredential(_ context.Context, workspaceID string, channel *agentsv1.TelegramChannel, cred telegramrepo.Credential, expectedRevision int64) (*agentsv1.TelegramChannel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.updateChannelLocked(workspaceID, channel, expectedRevision, &cred)
}

func (s *Store) updateChannelLocked(workspaceID string, channel *agentsv1.TelegramChannel, expectedRevision int64, cred *telegramrepo.Credential) (*agentsv1.TelegramChannel, error) {
	rec, ok := s.channels[channel.GetId()]
	if !ok || rec.spec.GetWorkspaceId() != workspaceID {
		return nil, channelNotFound(channel.GetId())
	}
	if rec.spec.GetRevision() != expectedRevision {
		return nil, fmt.Errorf("telegram channel %q (stored revision %d, expected %d): %w",
			channel.GetId(), rec.spec.GetRevision(), expectedRevision, telegramrepo.ErrRevisionConflict)
	}

	stored := proto.Clone(channel).(*agentsv1.TelegramChannel)
	// Immutable and repo-owned fields always come from the stored record.
	stored.WorkspaceId = workspaceID
	stored.Key = rec.spec.GetKey()
	stored.BotId = rec.spec.GetBotId()
	stored.CreatedAt = rec.spec.GetCreatedAt()
	stored.UpdatedAt = timestamppb.New(time.Now().UTC())
	stored.Revision = expectedRevision + 1
	rec.spec = stored
	if cred != nil {
		rec.cred = *cred
		if cred.Set() {
			rec.credUpdatedAt = time.Now().UTC()
		} else {
			rec.credUpdatedAt = time.Time{}
		}
	}
	return s.decodeChannel(rec), nil
}

func (s *Store) DeleteChannel(_ context.Context, workspaceID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.channels[id]
	if !ok || rec.spec.GetWorkspaceId() != workspaceID {
		return channelNotFound(id)
	}
	for _, dest := range s.destinations {
		if dest.spec.GetChannelId() == id {
			return fmt.Errorf("telegram channel %q is referenced by destination %q: %w",
				id, dest.spec.GetId(), telegramrepo.ErrChannelInUse)
		}
	}
	delete(s.channels, id)
	return nil
}

func (s *Store) SetChannelCredential(_ context.Context, workspaceID, id string, cred telegramrepo.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.channels[id]
	if !ok || rec.spec.GetWorkspaceId() != workspaceID {
		return channelNotFound(id)
	}
	rec.cred = cred
	if cred.Set() {
		rec.credUpdatedAt = time.Now().UTC()
	} else {
		rec.credUpdatedAt = time.Time{}
	}
	return nil
}

func (s *Store) GetChannelCredential(_ context.Context, workspaceID, id string) (telegramrepo.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.channels[id]
	if !ok || rec.spec.GetWorkspaceId() != workspaceID {
		return telegramrepo.Credential{}, channelNotFound(id)
	}
	if !rec.cred.Set() {
		return telegramrepo.Credential{}, fmt.Errorf("telegram channel %q: %w", id, telegramrepo.ErrNoCredential)
	}
	return rec.cred, nil
}

func (s *Store) SetWebhookSecret(_ context.Context, workspaceID, id string, cred telegramrepo.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.channels[id]
	if !ok || rec.spec.GetWorkspaceId() != workspaceID {
		return channelNotFound(id)
	}
	rec.webhookSecret = cred
	return nil
}

func (s *Store) SetWebhookSecretIfAbsent(_ context.Context, workspaceID, id string, cred telegramrepo.Credential) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.channels[id]
	if !ok || rec.spec.GetWorkspaceId() != workspaceID {
		return false, channelNotFound(id)
	}
	if rec.webhookSecret.Set() {
		return false, nil
	}
	rec.webhookSecret = cred
	return true, nil
}

func (s *Store) GetWebhookSecret(_ context.Context, workspaceID, id string) (telegramrepo.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.channels[id]
	if !ok || rec.spec.GetWorkspaceId() != workspaceID {
		return telegramrepo.Credential{}, channelNotFound(id)
	}
	if !rec.webhookSecret.Set() {
		return telegramrepo.Credential{}, fmt.Errorf("telegram channel %q: %w", id, telegramrepo.ErrNoCredential)
	}
	return rec.webhookSecret, nil
}

func (s *Store) ListChannelsAcrossWorkspaces(_ context.Context) ([]*agentsv1.TelegramChannel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*agentsv1.TelegramChannel, 0, len(s.channels))
	for _, rec := range s.channels {
		out = append(out, s.decodeChannel(rec))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].GetId() < out[j].GetId() })
	return out, nil
}

// --- Destinations ----------------------------------------------------------

func (s *Store) ListDestinations(_ context.Context, workspaceID, channelID string) ([]*agentsv1.TelegramDestination, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*agentsv1.TelegramDestination
	for _, rec := range s.destinations {
		if rec.spec.GetWorkspaceId() != workspaceID {
			continue
		}
		if channelID != "" && rec.spec.GetChannelId() != channelID {
			continue
		}
		out = append(out, proto.Clone(rec.spec).(*agentsv1.TelegramDestination))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].GetKey() < out[j].GetKey() })
	return out, nil
}

func (s *Store) GetDestination(_ context.Context, workspaceID, id string) (*agentsv1.TelegramDestination, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.destinations[id]
	if !ok || rec.spec.GetWorkspaceId() != workspaceID {
		return nil, destinationNotFound(id)
	}
	return proto.Clone(rec.spec).(*agentsv1.TelegramDestination), nil
}

func (s *Store) CreateDestination(_ context.Context, workspaceID string, dest *agentsv1.TelegramDestination) (*agentsv1.TelegramDestination, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, rec := range s.destinations {
		if rec.spec.GetWorkspaceId() == workspaceID && rec.spec.GetKey() == dest.GetKey() {
			return nil, fmt.Errorf("telegram destination %q: %w", dest.GetKey(), telegramrepo.ErrKeyExists)
		}
		if rec.spec.GetChannelId() == dest.GetChannelId() &&
			rec.spec.GetChatId() == dest.GetChatId() &&
			rec.spec.GetMessageThreadId() == dest.GetMessageThreadId() {
			return nil, fmt.Errorf("telegram destination for chat %q thread %q: %w",
				dest.GetChatId(), dest.GetMessageThreadId(), telegramrepo.ErrAddressExists)
		}
	}

	stored := proto.Clone(dest).(*agentsv1.TelegramDestination)
	stored.WorkspaceId = workspaceID
	now := timestamppb.New(time.Now().UTC())
	stored.CreatedAt = now
	stored.UpdatedAt = now
	stored.Revision = 1
	s.destinations[stored.GetId()] = &destinationRecord{spec: stored}
	return proto.Clone(stored).(*agentsv1.TelegramDestination), nil
}

func (s *Store) UpdateDestination(_ context.Context, workspaceID string, dest *agentsv1.TelegramDestination, expectedRevision int64) (*agentsv1.TelegramDestination, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.destinations[dest.GetId()]
	if !ok || rec.spec.GetWorkspaceId() != workspaceID {
		return nil, destinationNotFound(dest.GetId())
	}
	if rec.spec.GetRevision() != expectedRevision {
		return nil, fmt.Errorf("telegram destination %q (stored revision %d, expected %d): %w",
			dest.GetId(), rec.spec.GetRevision(), expectedRevision, telegramrepo.ErrRevisionConflict)
	}

	stored := proto.Clone(dest).(*agentsv1.TelegramDestination)
	stored.WorkspaceId = workspaceID
	// The address and key are immutable: redirecting traffic requires a new
	// Destination so persisted Cron / Notify Group references cannot be
	// silently repointed.
	stored.Key = rec.spec.GetKey()
	stored.ChannelId = rec.spec.GetChannelId()
	stored.ChatId = rec.spec.GetChatId()
	stored.MessageThreadId = rec.spec.GetMessageThreadId()
	stored.CreatedAt = rec.spec.GetCreatedAt()
	stored.UpdatedAt = timestamppb.New(time.Now().UTC())
	stored.Revision = expectedRevision + 1
	rec.spec = stored
	return proto.Clone(stored).(*agentsv1.TelegramDestination), nil
}

func (s *Store) DeleteDestination(_ context.Context, workspaceID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.destinations[id]
	if !ok || rec.spec.GetWorkspaceId() != workspaceID {
		return destinationNotFound(id)
	}
	delete(s.destinations, id)
	return nil
}

func (s *Store) FindDestinationByAddress(_ context.Context, channelID, chatID, threadID string) (*agentsv1.TelegramDestination, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, rec := range s.destinations {
		if rec.spec.GetChannelId() == channelID &&
			rec.spec.GetChatId() == chatID &&
			rec.spec.GetMessageThreadId() == threadID {
			return proto.Clone(rec.spec).(*agentsv1.TelegramDestination), nil
		}
	}
	return nil, fmt.Errorf("telegram destination for channel %q chat %q thread %q: %w",
		channelID, chatID, threadID, telegramrepo.ErrNotFound)
}

func (s *Store) ListDestinationsAcrossWorkspaces(_ context.Context) ([]*agentsv1.TelegramDestination, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*agentsv1.TelegramDestination, 0, len(s.destinations))
	for _, rec := range s.destinations {
		out = append(out, proto.Clone(rec.spec).(*agentsv1.TelegramDestination))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].GetId() < out[j].GetId() })
	return out, nil
}

func channelNotFound(id string) error {
	return fmt.Errorf("telegram channel %q: %w", id, telegramrepo.ErrNotFound)
}

func destinationNotFound(id string) error {
	return fmt.Errorf("telegram destination %q: %w", id, telegramrepo.ErrNotFound)
}
