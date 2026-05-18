package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"go.orx.me/apps/butter/internal/repo/auth"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	sessionKeyPrefix   = "butter:auth:session:"
	sessionIDKeyPrefix = "butter:auth:session:id:"
)

const touchSessionScript = `
if redis.call("GET", KEYS[2]) ~= ARGV[1] then
	return 0
end
local ttl = redis.call("PTTL", KEYS[1])
if ttl <= 0 then
	return 0
end
redis.call("SET", KEYS[1], ARGV[2], "PX", ttl)
redis.call("PEXPIRE", KEYS[2], ttl)
return 1
`

// Store keeps dashboard auth sessions in Redis while delegating user storage to
// the existing user repository.
type Store struct {
	users auth.Repository
	rdb   *goredis.Client
}

var _ auth.Repository = (*Store)(nil)

func New(users auth.Repository, rdb *goredis.Client) *Store {
	return &Store{users: users, rdb: rdb}
}

func (s *Store) EnsureIndexes(ctx context.Context) error {
	if s.users == nil {
		return nil
	}
	return s.users.EnsureIndexes(ctx)
}

func (s *Store) CountUsers(ctx context.Context) (int64, error) {
	return s.users.CountUsers(ctx)
}

func (s *Store) ListUsers(ctx context.Context) ([]*agentsv1.User, error) {
	return s.users.ListUsers(ctx)
}

func (s *Store) CreateUser(ctx context.Context, user *agentsv1.User, passwordHash string) error {
	return s.users.CreateUser(ctx, user, passwordHash)
}

func (s *Store) UpdateUserPassword(ctx context.Context, id string, passwordHash string, updatedAt time.Time) (*agentsv1.User, error) {
	return s.users.UpdateUserPassword(ctx, id, passwordHash, updatedAt)
}

func (s *Store) UpdateUserProfile(ctx context.Context, id string, displayName string, updatedAt time.Time) (*agentsv1.User, error) {
	return s.users.UpdateUserProfile(ctx, id, displayName, updatedAt)
}

func (s *Store) SetUserDisabled(ctx context.Context, id string, disabled bool, updatedAt time.Time) (*agentsv1.User, error) {
	return s.users.SetUserDisabled(ctx, id, disabled, updatedAt)
}

func (s *Store) FindUserByUsername(ctx context.Context, username string) (*agentsv1.User, string, error) {
	return s.users.FindUserByUsername(ctx, username)
}

func (s *Store) FindUserByID(ctx context.Context, id string) (*agentsv1.User, string, error) {
	return s.users.FindUserByID(ctx, id)
}

func (s *Store) GetUser(ctx context.Context, id string) (*agentsv1.User, error) {
	return s.users.GetUser(ctx, id)
}

func (s *Store) CreateSession(ctx context.Context, session *auth.Session) error {
	if s.rdb == nil {
		return errors.New("redis auth session store not available")
	}
	if session == nil || session.ID == "" || session.TokenHash == "" {
		return errors.New("invalid auth session")
	}
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return auth.ErrSessionNotFound
	}
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal auth session: %w", err)
	}

	pipe := s.rdb.TxPipeline()
	pipe.Set(ctx, sessionKey(session.TokenHash), data, ttl)
	pipe.Set(ctx, sessionIDKey(session.ID), session.TokenHash, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("create redis auth session: %w", err)
	}
	return nil
}

func (s *Store) LookupSession(ctx context.Context, tokenHash string, now time.Time) (*auth.Session, *agentsv1.User, error) {
	if s.rdb == nil {
		return nil, nil, errors.New("redis auth session store not available")
	}
	session, err := s.getSessionByHash(ctx, tokenHash)
	if err != nil {
		return nil, nil, err
	}
	if session.Revoked || !session.ExpiresAt.After(now) {
		_ = s.deleteSession(ctx, session)
		return nil, nil, auth.ErrSessionNotFound
	}

	user, err := s.users.GetUser(ctx, session.UserID)
	if err != nil {
		return nil, nil, err
	}
	if user.GetDisabled() {
		return nil, nil, auth.ErrUserDisabled
	}
	return session, user, nil
}

func (s *Store) TouchSession(ctx context.Context, id string, at time.Time) error {
	if s.rdb == nil || id == "" {
		return nil
	}
	tokenHash, err := s.rdb.Get(ctx, sessionIDKey(id)).Result()
	if errors.Is(err, goredis.Nil) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("get redis auth session id: %w", err)
	}

	session, err := s.getSessionByHash(ctx, tokenHash)
	if errors.Is(err, auth.ErrSessionNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if !session.ExpiresAt.After(at) {
		_ = s.deleteSession(ctx, session)
		return nil
	}
	session.LastUsedAt = at
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal auth session: %w", err)
	}
	if err := s.rdb.Eval(ctx, touchSessionScript, []string{sessionKey(tokenHash), sessionIDKey(id)}, tokenHash, data).Err(); err != nil {
		return fmt.Errorf("touch redis auth session: %w", err)
	}
	return nil
}

func (s *Store) RevokeSession(ctx context.Context, id string) error {
	if s.rdb == nil || id == "" {
		return nil
	}
	tokenHash, err := s.rdb.Get(ctx, sessionIDKey(id)).Result()
	if errors.Is(err, goredis.Nil) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("get redis auth session id: %w", err)
	}
	return s.deleteSession(ctx, &auth.Session{ID: id, TokenHash: tokenHash})
}

func (s *Store) getSessionByHash(ctx context.Context, tokenHash string) (*auth.Session, error) {
	raw, err := s.rdb.Get(ctx, sessionKey(tokenHash)).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, auth.ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lookup redis auth session: %w", err)
	}
	session := &auth.Session{}
	if err := json.Unmarshal(raw, session); err != nil {
		return nil, fmt.Errorf("unmarshal auth session: %w", err)
	}
	return session, nil
}

func (s *Store) deleteSession(ctx context.Context, session *auth.Session) error {
	if session == nil {
		return nil
	}
	if err := s.rdb.Del(ctx, sessionKey(session.TokenHash), sessionIDKey(session.ID)).Err(); err != nil {
		return fmt.Errorf("delete redis auth session: %w", err)
	}
	return nil
}

func sessionKey(tokenHash string) string {
	return sessionKeyPrefix + tokenHash
}

func sessionIDKey(id string) string {
	return sessionIDKeyPrefix + id
}
