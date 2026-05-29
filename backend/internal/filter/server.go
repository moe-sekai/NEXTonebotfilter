package filter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type dedupCache struct {
	ttl   time.Duration
	store sync.Map
	stop  chan struct{}
}

func newDedupCache(ttlSeconds int) *dedupCache {
	if ttlSeconds <= 0 {
		ttlSeconds = 60
	}
	d := &dedupCache{
		ttl:  time.Duration(ttlSeconds) * time.Second,
		stop: make(chan struct{}),
	}
	go d.cleanup()
	return d
}

func (d *dedupCache) IsDup(key uint64) bool {
	now := time.Now().UnixNano()
	expireAt := now + int64(d.ttl)
	for {
		v, loaded := d.store.LoadOrStore(key, expireAt)
		if !loaded {
			return false
		}
		oldExpireAt := v.(int64)
		if oldExpireAt > now {
			return true
		}
		if d.store.CompareAndSwap(key, oldExpireAt, expireAt) {
			return false
		}
	}
}

func (d *dedupCache) cleanup() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-d.stop:
			return
		case <-ticker.C:
			now := time.Now().UnixNano()
			d.store.Range(func(k, v any) bool {
				if v.(int64) <= now {
					d.store.Delete(k)
				}
				return true
			})
		}
	}
}

func (d *dedupCache) Stop() { close(d.stop) }

type dedupProbe struct {
	PostType      string          `json:"post_type"`
	SelfID        int64           `json:"self_id"`
	MessageType   string          `json:"message_type"`
	GroupID       int64           `json:"group_id"`
	UserID        int64           `json:"user_id"`
	Time          int64           `json:"time"`
	RawMessage    string          `json:"raw_message"`
	Message       json.RawMessage `json:"message"`
	MessageID     json.RawMessage `json:"message_id"`
	MessageSeq    json.RawMessage `json:"message_seq"`
	RealID        json.RawMessage `json:"real_id"`
	MessageIDV12  json.RawMessage `json:"id"`
	AltMessageV12 string          `json:"alt_message"`
}

func (p *dedupProbe) contentFingerprint() string {
	if p.RawMessage != "" {
		return p.RawMessage
	}
	if len(p.Message) > 0 && string(p.Message) != "null" {
		return string(p.Message)
	}
	return p.AltMessageV12
}

func (p *dedupProbe) transportMessageID() string {
	for _, raw := range []json.RawMessage{p.MessageID, p.MessageSeq, p.RealID, p.MessageIDV12} {
		if len(raw) == 0 || string(raw) == "null" {
			continue
		}
		return string(raw)
	}
	return ""
}

func dedupKey(p *dedupProbe) uint64 {
	h := fnv.New64a()
	content := p.contentFingerprint()
	if content == "" {
		content = p.transportMessageID()
	}
	fmt.Fprintf(h, "%s:%d:%d:%d:%s", p.MessageType, p.GroupID, p.UserID, p.Time, content)
	return h.Sum64()
}

func dedupCandidate(p *dedupProbe) bool {
	if p.PostType != "message" {
		return false
	}
	if p.MessageType != MessageTypeGroup || p.GroupID == 0 {
		return false
	}
	return p.UserID != 0 && p.Time != 0 && (p.contentFingerprint() != "" || p.transportMessageID() != "")
}

type wsServer struct {
	mu        sync.RWMutex
	upstreams map[string]*upstream
	dedup     *dedupCache

	clientsMu sync.RWMutex
	clients   []*wsClient
}

type upstream struct {
	selfID    string
	remote    string
	conn      *websocket.Conn
	writeChan chan wsMsg
	connected time.Time
}

func newWsServer() *wsServer {
	return &wsServer{upstreams: map[string]*upstream{}}
}

func (s *wsServer) serve(ctx context.Context, conn *websocket.Conn, selfID, remote string) error {
	if selfID == "" {
		selfID = "anon-" + remote
	}
	u := &upstream{
		selfID:    selfID,
		remote:    remote,
		conn:      conn,
		writeChan: make(chan wsMsg, 64),
		connected: time.Now(),
	}
	s.mu.Lock()
	if existing, ok := s.upstreams[selfID]; ok {
		delete(s.upstreams, selfID)
		_ = existing.conn.Close()
	}
	s.upstreams[selfID] = u
	s.mu.Unlock()

	innerCtx, cancel := context.WithCancel(ctx)
	defer s.removeUpstream(selfID, u, cancel)

	go s.upstreamWriteLoop(innerCtx, u)

	for {
		mt, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		if mt == websocket.TextMessage {
			data = ensurePayloadSelfID(data, selfID)
		}
		if s.dedup != nil && mt == websocket.TextMessage {
			var p dedupProbe
			if json.Unmarshal(data, &p) == nil && dedupCandidate(&p) {
				key := dedupKey(&p)
				if s.dedup.IsDup(key) {
					log.Debug().
						Str("self_id", selfID).
						Uint64("key", key).
						Int64("group_id", p.GroupID).
						Int64("user_id", p.UserID).
						Int64("time", p.Time).
						Msg("filter: dedup skipped duplicate group message")
					continue
				}
			}
		}
		s.broadcastToClients(wsMsg{mt: mt, data: data, selfID: selfID})
	}
}

func (s *wsServer) removeUpstream(selfID string, expected *upstream, cancel context.CancelFunc) {
	cancel()
	s.mu.Lock()
	u, ok := s.upstreams[selfID]
	if ok && (expected == nil || u == expected) {
		delete(s.upstreams, selfID)
	} else {
		ok = false
	}
	s.mu.Unlock()
	if ok && u.conn != nil {
		_ = u.conn.Close()
	}
}

func (s *wsServer) broadcastToClients(msg wsMsg) {
	for _, c := range s.snapshotClients() {
		go func(c *wsClient, m wsMsg) {
			if err := c.write(m); err != nil {
				log.Debug().Str("client", c.name).Err(err).Msg("filter: forward to client failed")
			}
		}(c, msg)
	}
}

func (s *wsServer) writeMessage(mt int, data []byte, boundSelfID string) error {
	s.mu.RLock()
	count := len(s.upstreams)
	s.mu.RUnlock()
	if count == 0 {
		return errors.New("filter: no upstream OneBot client")
	}

	target := boundSelfID
	if target == "" && mt == websocket.TextMessage {
		var probe struct {
			SelfID json.Number `json:"self_id"`
		}
		if err := json.Unmarshal(data, &probe); err == nil {
			target = probe.SelfID.String()
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	if target != "" {
		if u, ok := s.upstreams[target]; ok {
			return enqueue(u, wsMsg{mt: mt, data: data, selfID: target})
		}
		return fmt.Errorf("filter: upstream self_id %s not connected", target)
	}
	var firstErr error
	for _, u := range s.upstreams {
		if err := enqueue(u, wsMsg{mt: mt, data: data, selfID: u.selfID}); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func enqueue(u *upstream, msg wsMsg) error {
	select {
	case u.writeChan <- msg:
		return nil
	default:
		return fmt.Errorf("filter: upstream %s write channel full", u.selfID)
	}
}

func (s *wsServer) addClient(c *wsClient) error {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	for _, existing := range s.clients {
		if existing.name == c.name {
			return fmt.Errorf("filter: client %s already connected", c.name)
		}
	}
	s.clients = append(s.clients, c)
	return nil
}

func (s *wsServer) removeClient(name string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	for i, c := range s.clients {
		if c.name == name {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			return
		}
	}
}

func (s *wsServer) snapshotClients() []*wsClient {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	out := make([]*wsClient, len(s.clients))
	copy(out, s.clients)
	return out
}

func (s *wsServer) snapshotUpstreams() []UpstreamStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]UpstreamStatus, 0, len(s.upstreams))
	for _, u := range s.upstreams {
		t := u.connected
		out = append(out, UpstreamStatus{
			SelfID:    u.selfID,
			Remote:    u.remote,
			Connected: true,
			Since:     &t,
		})
	}
	return out
}

func (s *wsServer) upstreamWriteLoop(ctx context.Context, u *upstream) {
	for {
		select {
		case msg, ok := <-u.writeChan:
			if !ok {
				return
			}
			if u.conn == nil {
				continue
			}
			if err := u.conn.WriteMessage(msg.mt, msg.data); err != nil {
				log.Warn().Str("self_id", u.selfID).Err(err).Msg("filter: write to upstream OneBot client failed")
			}
		case <-ctx.Done():
			return
		}
	}
}

func ensurePayloadSelfID(data []byte, selfID string) []byte {
	if selfID == "" {
		return data
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return data
	}
	if _, ok := obj["self_id"]; ok {
		return data
	}
	if numericSelfID, err := strconv.ParseInt(selfID, 10, 64); err == nil {
		encoded, _ := json.Marshal(numericSelfID)
		obj["self_id"] = encoded
	} else {
		encoded, _ := json.Marshal(selfID)
		obj["self_id"] = encoded
	}
	return encodeRawMap(obj)
}

func encodeRawMap(obj map[string]json.RawMessage) []byte {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	return data
}
