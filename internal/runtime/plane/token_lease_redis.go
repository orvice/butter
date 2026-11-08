package plane

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

var ErrTokenFenceConflict = errors.New("plane token lease fence conflict")

// RedisFencedTokenLease is a token-specific CAS backend. Its keyspace is
// separate from redislease: generic liveness leases must never fence tokens.
type RedisFencedTokenLease struct {
	Client redis.UniversalClient
	Prefix string
}

func (r *RedisFencedTokenLease) key(k string) string { return r.Prefix + "plane:token-fence:" + k }
func (r *RedisFencedTokenLease) AcquireToken(ctx context.Context, key, owner string, revision int64, expiry time.Time) (uint64, error) {
	if r == nil || r.Client == nil || owner == "" || key == "" || expiry.IsZero() {
		return 0, ErrRuntimeDependency
	}
	v := r.Client.Incr(ctx, r.key(key)+":seq")
	if v.Err() != nil {
		return 0, v.Err()
	}
	f := v.Val()
	ms := expiry.Sub(time.Now()).Milliseconds()
	if ms <= 0 {
		return 0, ErrTokenFenceConflict
	}
	script := redis.NewScript(`local cur=redis.call('GET',KEYS[1]); if cur then local o,rev,ex=string.match(cur,'([^|]+)|([^|]+)|([^|]+)'); if tonumber(ex)>tonumber(ARGV[5]) and (o~=ARGV[1] or tonumber(rev)>tonumber(ARGV[2])) then return 0 end end; redis.call('SET',KEYS[1],ARGV[1]..'|'..ARGV[2]..'|'..ARGV[4], 'PX', ARGV[3]); return ARGV[4]`)
	out, e := script.Run(ctx, r.Client, []string{r.key(key)}, owner, strconv.FormatInt(revision, 10), strconv.FormatInt(ms, 10), strconv.FormatInt(f, 10), strconv.FormatInt(time.Now().UnixMilli(), 10)).Int64()
	if e != nil {
		return 0, e
	}
	if out == 0 {
		return 0, ErrTokenFenceConflict
	}
	return uint64(out), nil
}
func (r *RedisFencedTokenLease) ReleaseToken(ctx context.Context, key, owner string, fence uint64) error {
	if r == nil || r.Client == nil {
		return ErrRuntimeDependency
	}
	s := redis.NewScript(`local cur=redis.call('GET',KEYS[1]); if not cur then return 0 end; local o,rev,f=string.match(cur,'([^|]+)|([^|]+)|([^|]+)'); if o==ARGV[1] and f==ARGV[2] then return redis.call('DEL',KEYS[1]) end; return 0`)
	n, e := s.Run(ctx, r.Client, []string{r.key(key)}, owner, strconv.FormatUint(fence, 10)).Int64()
	if e != nil {
		return e
	}
	if n == 0 {
		return ErrTokenFenceConflict
	}
	return nil
}

var _ FencedTokenLease = (*RedisFencedTokenLease)(nil)
