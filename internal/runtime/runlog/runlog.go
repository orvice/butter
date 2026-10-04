// Package runlog keeps the Run Log of a Detached AG-UI run (ADR-0016
// decision 5): the run's events, in order, which every observer of the run
// replays from the start and follows to the run's end, on any Pod.
//
// One writer, the run, appends to a log. It appends in batches, each at the
// length it knows the log has, so an append retried after a lost reply never
// adds its entries twice. A log ends with the run's terminal event or with a
// truncation marker, when the run's events outgrew the log or its writer fell
// behind; nothing follows either. The writer keeps the log alive while the
// run lives and gives it a retention once the run ended, so a log lapses a
// while after its run ended or died. Once it lapsed or was dropped it never
// comes back: an append finds it gone.
//
// Observers read a log from any process. Read never blocks; Wait blocks until
// the log may hold more. The Redis store multiplexes the waits of all the
// observers of a process over one blocking XREAD, so an observer never holds
// a Redis connection of its own.
//
// A log is a replay buffer, not a record: the session stays the record of the
// conversation, so losing a log costs only the live view.
package runlog

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"sync"
	"time"
)

// Kind says what a log entry holds.
type Kind string

const (
	// KindEvent is one of the run's events.
	KindEvent Kind = "event"
	// KindEnd is the run's terminal event. Nothing follows it.
	KindEnd Kind = "end"
	// KindTruncated marks a log that stopped short of the run's end: the
	// run's later events were not kept. It carries no frame, and nothing
	// follows it.
	KindTruncated Kind = "truncated"
)

// Final reports whether nothing follows an entry of this kind.
func (k Kind) Final() bool { return k == KindEnd || k == KindTruncated }

// Entry is one entry of a log.
type Entry struct {
	Kind Kind
	// Frame is the event as observers send it on; empty for a truncation
	// marker.
	Frame []byte
}

// Record is an entry read back, with its place in the log.
type Record struct {
	// ID places the entry in its log. It is opaque: a reader hands the last
	// one it read back to Read and Wait.
	ID string
	Entry
}

var (
	// ErrGone reports a log that does not exist: it lapsed, was dropped, or
	// was never opened.
	ErrGone = errors.New("run log is gone")
	// ErrConflict reports an append that found the log holding neither what
	// its writer expected nor that plus the entries being appended.
	ErrConflict = errors.New("run log does not hold what its writer appended")
)

// Store keeps run logs, by key. A key names one run's log.
type Store interface {
	// Open starts an empty log at key, replacing any log there, and keeps it
	// for ttl.
	Open(ctx context.Context, key string, ttl time.Duration) error
	// Append adds entries, in order, to the log at key, which holds at
	// entries: as many as its writer appended so far. An append whose
	// entries are already there — its reply was lost and it is retried —
	// changes nothing. Either way the log is then kept for ttl from now.
	// ErrGone when the log does not exist: an append never brings a log
	// back. ErrConflict when it holds neither at nor at+len(entries)
	// entries.
	Append(ctx context.Context, key string, at int, entries []Entry, ttl time.Duration) error
	// Expire keeps the log at key for ttl from now. ErrGone when it does not
	// exist.
	Expire(ctx context.Context, key string, ttl time.Duration) error
	// Read returns up to count entries of the log at key, oldest first: from
	// its start when after is empty, else those after the entry whose ID it
	// is. ErrGone when the log does not exist.
	Read(ctx context.Context, key, after string, count int) ([]Record, error)
	// Wait returns once the log at key may hold entries after the one whose
	// ID is after (empty: any entry), or may be gone; else after timeout, or
	// when ctx ends. The caller reads to find out.
	Wait(ctx context.Context, key, after string, timeout time.Duration)
	// Drop removes the log at key.
	Drop(ctx context.Context, key string) error
}

// --- In-process store ------------------------------------------------------------

// Memory is the in-process Store. It serves a single process and the tests,
// and honors the same TTLs as Redis. A log that lapsed is let go when it is
// next touched, and every Open lets go of all that lapsed, so the logs of
// ended runs do not pile up in memory.
type Memory struct {
	mu   sync.Mutex
	seq  uint64
	logs map[string]*memoryLog
}

type memoryLog struct {
	records []memoryRecord
	expires time.Time
	// changed is closed, and replaced, whenever the log changes, which wakes
	// its waiters.
	changed chan struct{}
}

type memoryRecord struct {
	seq uint64
	Entry
}

// NewMemory builds an in-process store.
func NewMemory() *Memory {
	return &Memory{logs: map[string]*memoryLog{}}
}

var _ Store = (*Memory)(nil)

// current returns the unexpired log at key. The caller holds m.mu.
func (m *Memory) current(key string, now time.Time) *memoryLog {
	l, ok := m.logs[key]
	if !ok {
		return nil
	}
	if !now.Before(l.expires) {
		m.remove(key, l)
		return nil
	}
	return l
}

// remove deletes the log at key and wakes its waiters. The caller holds m.mu.
func (m *Memory) remove(key string, l *memoryLog) {
	delete(m.logs, key)
	close(l.changed)
}

// touched wakes the log's waiters. The caller holds m.mu.
func (l *memoryLog) touched() {
	close(l.changed)
	l.changed = make(chan struct{})
}

func (m *Memory) Open(_ context.Context, key string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for k, l := range m.logs {
		if !now.Before(l.expires) {
			m.remove(k, l)
		}
	}
	if l, ok := m.logs[key]; ok {
		m.remove(key, l)
	}
	m.logs[key] = &memoryLog{expires: now.Add(ttl), changed: make(chan struct{})}
	return nil
}

func (m *Memory) Append(_ context.Context, key string, at int, entries []Entry, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	l := m.current(key, now)
	if l == nil {
		return ErrGone
	}
	switch len(l.records) {
	case at + len(entries):
		// Already appended.
	case at:
		for _, e := range entries {
			m.seq++
			l.records = append(l.records, memoryRecord{seq: m.seq, Entry: Entry{Kind: e.Kind, Frame: slices.Clone(e.Frame)}})
		}
		l.touched()
	default:
		return ErrConflict
	}
	l.expires = now.Add(ttl)
	return nil
}

func (m *Memory) Expire(_ context.Context, key string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	l := m.current(key, now)
	if l == nil {
		return ErrGone
	}
	l.expires = now.Add(ttl)
	return nil
}

// memoryAfter is the sequence number an after ID names; 0 for the start.
func memoryAfter(after string) uint64 {
	if after == "" {
		return 0
	}
	seq, err := strconv.ParseUint(after, 10, 64)
	if err != nil {
		return 0
	}
	return seq
}

func (m *Memory) Read(_ context.Context, key, after string, count int) ([]Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	l := m.current(key, time.Now())
	if l == nil {
		return nil, ErrGone
	}
	from := memoryAfter(after)
	start, _ := slices.BinarySearchFunc(l.records, from+1, func(r memoryRecord, seq uint64) int {
		switch {
		case r.seq < seq:
			return -1
		case r.seq > seq:
			return 1
		default:
			return 0
		}
	})
	end := min(start+max(count, 1), len(l.records))
	out := make([]Record, 0, end-start)
	for _, r := range l.records[start:end] {
		out = append(out, Record{
			ID:    strconv.FormatUint(r.seq, 10),
			Entry: Entry{Kind: r.Kind, Frame: slices.Clone(r.Frame)},
		})
	}
	return out, nil
}

func (m *Memory) Wait(ctx context.Context, key, after string, timeout time.Duration) {
	m.mu.Lock()
	l := m.current(key, time.Now())
	if l == nil || (len(l.records) > 0 && l.records[len(l.records)-1].seq > memoryAfter(after)) {
		m.mu.Unlock()
		return
	}
	changed := l.changed
	m.mu.Unlock()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-changed:
	case <-timer.C:
	case <-ctx.Done():
	}
}

func (m *Memory) Drop(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if l, ok := m.logs[key]; ok {
		m.remove(key, l)
	}
	return nil
}
