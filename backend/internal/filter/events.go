package filter

import (
	"sync"
	"time"
)

type EventKind string

const (
	EventAllow        EventKind = "allow"
	EventBlock        EventKind = "block"
	EventPrefixPass   EventKind = "prefix_pass"
	EventClientUp     EventKind = "client_up"
	EventClientDown   EventKind = "client_down"
	EventUpstreamUp   EventKind = "upstream_up"
	EventUpstreamDown EventKind = "upstream_down"
)

type Event struct {
	Seq     uint64    `json:"seq"`
	Time    time.Time `json:"time"`
	Kind    EventKind `json:"kind"`
	Filter  string    `json:"filter,omitempty"`
	Reason  string    `json:"reason,omitempty"`
	UserID  int64     `json:"user_id,omitempty"`
	GroupID int64     `json:"group_id,omitempty"`
	MsgType string    `json:"msg_type,omitempty"`
	Raw     string    `json:"raw,omitempty"`
}

type eventBus struct {
	mu        sync.RWMutex
	buf       []Event
	cap       int
	head      int
	count     int
	seq       uint64
	subs      map[int]chan Event
	subNextID int
}

func newEventBus(capacity int) *eventBus {
	if capacity <= 0 {
		capacity = 256
	}
	return &eventBus{
		buf:  make([]Event, capacity),
		cap:  capacity,
		subs: map[int]chan Event{},
	}
}

func (b *eventBus) Publish(ev Event) {
	if len(ev.Raw) > 256 {
		ev.Raw = ev.Raw[:256] + "..."
	}
	b.mu.Lock()
	b.seq++
	ev.Seq = b.seq
	if ev.Time.IsZero() {
		ev.Time = time.Now()
	}
	b.buf[b.head] = ev
	b.head = (b.head + 1) % b.cap
	if b.count < b.cap {
		b.count++
	}
	subs := make([]chan Event, 0, len(b.subs))
	for _, ch := range b.subs {
		subs = append(subs, ch)
	}
	b.mu.Unlock()
	for _, ch := range subs {
		select {
		case ch <- ev:
		default:
		}
	}
}

func (b *eventBus) Snapshot(limit int) []Event {
	b.mu.RLock()
	defer b.mu.RUnlock()
	n := b.count
	if limit > 0 && limit < n {
		n = limit
	}
	out := make([]Event, 0, n)
	start := (b.head - b.count + b.cap) % b.cap
	skip := b.count - n
	for i := 0; i < n; i++ {
		idx := (start + skip + i) % b.cap
		out = append(out, b.buf[idx])
	}
	return out
}

func (b *eventBus) Subscribe() (<-chan Event, func()) {
	ch := make(chan Event, 64)
	b.mu.Lock()
	id := b.subNextID
	b.subNextID++
	b.subs[id] = ch
	b.mu.Unlock()
	return ch, func() {
		b.mu.Lock()
		if c, ok := b.subs[id]; ok {
			delete(b.subs, id)
			close(c)
		}
		b.mu.Unlock()
	}
}
