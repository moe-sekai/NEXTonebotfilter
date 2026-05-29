package filter

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"github.com/exmeaning/nextonebotfilter/internal/store"
)

// Manager owns the lifecycle of the OneBot filter gateway.
type Manager struct {
	db store.Store

	mu        sync.Mutex
	cancel    context.CancelFunc
	httpSrv   *http.Server
	server    *wsServer
	upgrader  websocket.Upgrader
	clients   map[string]*wsClient
	filters   map[string]*Filter
	bus       *eventBus
	gateway   store.FilterGateway
	debug     bool
	startedAt time.Time
	running   bool
}

func New(db store.Store) *Manager {
	return &Manager{
		db:      db,
		clients: map[string]*wsClient{},
		filters: map[string]*Filter{},
		bus:     newEventBus(512),
	}
}

func (m *Manager) RecentEvents(limit int) []Event { return m.bus.Snapshot(limit) }

func (m *Manager) Subscribe() (<-chan Event, func()) { return m.bus.Subscribe() }

func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.running {
		return errors.New("filter manager already running")
	}

	gw, err := m.db.GetOrCreateFilterGateway()
	if err != nil {
		return fmt.Errorf("load filter gateway: %w", err)
	}
	m.gateway = *gw
	if !gw.Enabled {
		log.Info().Msg("Filter gateway disabled in config; not starting")
		return nil
	}

	m.upgrader = websocket.Upgrader{
		ReadBufferSize:  gw.BufferSize,
		WriteBufferSize: gw.BufferSize,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	m.server = newWsServer()
	if gw.DedupEnabled {
		m.server.dedup = newDedupCache(gw.DedupTTL)
		log.Info().Int("ttl", gw.DedupTTL).Msg("Filter: message dedup enabled")
	}

	mux := http.NewServeMux()
	suffix := gw.Suffix
	if suffix == "" {
		suffix = "/ws"
	}
	mux.HandleFunc(suffix, m.handleUpstream)
	addr := fmt.Sprintf("%s:%d", gw.Host, gw.Port)
	m.httpSrv = &http.Server{Addr: addr, Handler: mux}

	managerCtx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	m.debug = gw.Debug
	m.startedAt = time.Now()

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		cancel()
		m.cancel = nil
		return fmt.Errorf("listen %s: %w", addr, err)
	}
	go func() {
		log.Info().Str("addr", addr).Str("path", suffix).Msg("Filter gateway listening")
		if err := m.httpSrv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("Filter gateway server stopped")
		}
	}()

	if err := m.startClientsLocked(managerCtx); err != nil {
		log.Warn().Err(err).Msg("Filter clients failed to (re)load")
	}
	m.running = true
	return nil
}

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.running {
		return
	}
	m.running = false
	m.stopClientsLocked()
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	if m.httpSrv != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = m.httpSrv.Shutdown(shutdownCtx)
		m.httpSrv = nil
	}
	if m.server != nil && m.server.dedup != nil {
		m.server.dedup.Stop()
	}
	m.server = nil
}

func (m *Manager) Reload(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.running {
		return nil
	}
	gw, err := m.db.GetOrCreateFilterGateway()
	if err != nil {
		return err
	}
	m.gateway = *gw
	m.debug = gw.Debug
	if m.server != nil {
		if m.server.dedup != nil {
			m.server.dedup.Stop()
			m.server.dedup = nil
		}
		if gw.DedupEnabled {
			m.server.dedup = newDedupCache(gw.DedupTTL)
		}
	}
	m.stopClientsLocked()
	return m.startClientsLocked(ctx)
}

func (m *Manager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

type Status struct {
	Running    bool             `json:"running"`
	Listen     string           `json:"listen"`
	Suffix     string           `json:"suffix"`
	UpstreamUp bool             `json:"upstream_up"`
	StartedAt  *time.Time       `json:"started_at,omitempty"`
	Upstreams  []UpstreamStatus `json:"upstreams"`
	Clients    []ClientStatus   `json:"clients"`
}

type UpstreamStatus struct {
	SelfID    string     `json:"self_id"`
	Remote    string     `json:"remote"`
	Connected bool       `json:"connected"`
	Since     *time.Time `json:"since,omitempty"`
}

type ClientStatus struct {
	Name      string `json:"name"`
	URI       string `json:"uri"`
	Connected bool   `json:"connected"`
	Builtin   bool   `json:"builtin"`
}

func (m *Manager) Status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	st := Status{
		Running: m.running,
		Suffix:  m.gateway.Suffix,
	}
	if m.running {
		st.Listen = fmt.Sprintf("%s:%d", m.gateway.Host, m.gateway.Port)
		t := m.startedAt
		st.StartedAt = &t
	}
	if m.server != nil {
		st.Upstreams = m.server.snapshotUpstreams()
		st.UpstreamUp = len(st.Upstreams) > 0
	}
	apps, _ := m.db.ListFilterApps()
	for _, app := range apps {
		c, ok := m.clients[app.Name]
		st.Clients = append(st.Clients, ClientStatus{
			Name:      app.Name,
			URI:       app.URI,
			Builtin:   app.Builtin,
			Connected: ok && c.isConnected(),
		})
	}
	return st
}

func (m *Manager) handleUpstream(w http.ResponseWriter, r *http.Request) {
	if m.server == nil {
		http.Error(w, "filter gateway not ready", http.StatusServiceUnavailable)
		return
	}
	m.mu.Lock()
	expected := m.gateway.AccessToken
	m.mu.Unlock()
	if expected != "" && !checkAccessToken(r, expected) {
		log.Warn().Str("remote", r.RemoteAddr).Msg("Filter: upstream rejected, token mismatch")
		w.Header().Set("WWW-Authenticate", `Bearer realm="nextonebotfilter"`)
		http.Error(w, "invalid access token", http.StatusUnauthorized)
		return
	}
	selfID := r.Header.Get("x-self-id")
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warn().Err(err).Msg("Filter: upstream upgrade failed")
		return
	}
	log.Info().Str("remote", r.RemoteAddr).Str("self_id", selfID).Msg("Filter: upstream OneBot client connected")
	m.bus.Publish(Event{Kind: EventUpstreamUp, Reason: r.RemoteAddr, Filter: selfID})
	defer m.bus.Publish(Event{Kind: EventUpstreamDown, Reason: r.RemoteAddr, Filter: selfID})
	if err := m.server.serve(r.Context(), conn, selfID, r.RemoteAddr); err != nil {
		log.Info().Err(err).Str("self_id", selfID).Msg("Filter: upstream OneBot client disconnected")
	}
}

func checkAccessToken(r *http.Request, expected string) bool {
	if expected == "" {
		return true
	}
	if got := r.URL.Query().Get("access_token"); got != "" && got == expected {
		return true
	}
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return false
	}
	for _, prefix := range []string{"Bearer ", "Token "} {
		if strings.HasPrefix(auth, prefix) {
			return strings.TrimSpace(auth[len(prefix):]) == expected
		}
	}
	return strings.TrimSpace(auth) == expected
}

func (m *Manager) startClientsLocked(ctx context.Context) error {
	apps, err := m.db.ListFilterApps()
	if err != nil {
		return err
	}
	templates, err := m.db.ListFilterTemplates()
	if err != nil {
		return err
	}
	tplByID := map[uint]*store.FilterTemplate{}
	for i := range templates {
		t := &templates[i]
		tplByID[t.ID] = t
	}
	defaultTpl, err := m.db.GetDefaultFilterTemplate()
	if err != nil {
		return err
	}
	defaultUserID := DecodeIDRule(defaultTpl.UserIDRules)
	defaultGroupID := DecodeIDRule(defaultTpl.GroupIDRules)
	snap := gatewaySnapshot{
		BotID:      m.gateway.BotID,
		UserAgent:  m.gateway.UserAgent,
		BufferSize: m.gateway.BufferSize,
		SleepTime:  m.gateway.SleepTime,
		Debug:      m.gateway.Debug,
	}
	if snap.BufferSize <= 0 {
		snap.BufferSize = 4096
	}
	if snap.SleepTime <= 0 {
		snap.SleepTime = 5
	}
	for _, app := range apps {
		if !app.Enabled {
			continue
		}
		userID, groupID, msg, priv, grp := appOrTemplateRules(&app, tplByID)
		f := &Filter{}
		f.Compile(CompiledRules{
			Name:           app.Name,
			UserID:         userID,
			GroupID:        groupID,
			Message:        msg,
			PrivateMessage: priv,
			GroupMessage:   grp,
			DefaultUserID:  defaultUserID,
			DefaultGroupID: defaultGroupID,
		})
		f.SetPublisher(m.bus.Publish)
		m.filters[app.Name] = f
		if app.Internal {
			continue
		}
		c := newWsClient(app.Name, app.URI, app.AccessToken, f, m.debug, m.bus.Publish, false)
		m.clients[app.Name] = c
		go c.run(ctx, m.server, snap)
	}
	return nil
}

// AllowMessage lets external callers run a message through a named app's
// compiled rules. Returns true when the app is missing/disabled (caller decides).
func (m *Manager) AllowMessage(appName string, groupID, userID int64, isPrivate bool, raw string) bool {
	m.mu.Lock()
	f := m.filters[appName]
	m.mu.Unlock()
	if f == nil {
		return true
	}
	mt := MessageTypeGroup
	if isPrivate {
		mt = MessageTypePrivate
	}
	probe := &OneBotMessage{
		Partial: OneBotMessagePartial{
			MessageType:   mt,
			MessageFormat: MessageFormatString,
			MessageString: raw,
			RawMessage:    raw,
			UserID:        userID,
			GroupID:       groupID,
		},
	}
	return f.Allow(probe, m.debug)
}

func (m *Manager) IsAppEnabled(appName string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.filters[appName]
	return ok
}

func appOrTemplateRules(app *store.FilterApp, tplByID map[uint]*store.FilterTemplate) (IDRule, IDRule, MessageRule, MessageRule, MessageRule) {
	if app.TemplateID != nil {
		if t, ok := tplByID[*app.TemplateID]; ok {
			return DecodeIDRule(t.UserIDRules),
				DecodeIDRule(t.GroupIDRules),
				DecodeMessageRule(t.MessageRules),
				DecodeMessageRule(t.PrivateMessageRules),
				DecodeMessageRule(t.GroupMessageRules)
		}
	}
	return DecodeIDRule(app.UserIDRules),
		DecodeIDRule(app.GroupIDRules),
		DecodeMessageRule(app.MessageRules),
		DecodeMessageRule(app.PrivateMessageRules),
		DecodeMessageRule(app.GroupMessageRules)
}

func (m *Manager) stopClientsLocked() {
	for name, c := range m.clients {
		close(c.stop)
		<-c.stopped
		_ = name
	}
	m.clients = map[string]*wsClient{}
	m.filters = map[string]*Filter{}
}
