package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/exmeaning/nextonebotfilter/internal/filter"
	"github.com/exmeaning/nextonebotfilter/internal/store"
)

// API exposes the console-facing REST + SSE endpoints.
type API struct {
	db      store.Store
	manager *filter.Manager
}

func NewAPI(db store.Store, m *filter.Manager) *API {
	return &API{db: db, manager: m}
}

// Routes returns a mux populated with all /api/* handlers.
func (a *API) Routes(mux *http.ServeMux) {
	mux.HandleFunc("/api/health", a.health)

	mux.HandleFunc("/api/gateway", a.gateway)
	mux.HandleFunc("/api/gateway/restart", a.restart)
	mux.HandleFunc("/api/status", a.status)

	mux.HandleFunc("/api/apps", a.apps)
	mux.HandleFunc("/api/apps/", a.appByID)

	mux.HandleFunc("/api/templates", a.templates)
	mux.HandleFunc("/api/templates/", a.templateByID)

	mux.HandleFunc("/api/events", a.recentEvents)
	mux.HandleFunc("/api/events/stream", a.eventStream)

	mux.HandleFunc("/api/yaml/export", a.yamlExport)
	mux.HandleFunc("/api/yaml/import", a.yamlImport)

	mux.HandleFunc("/api/regex/test", a.regexTest)
}

func (a *API) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "time": time.Now()})
}

func (a *API) status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.manager.Status())
}

func (a *API) restart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := a.manager.Reload(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *API) gateway(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		gw, err := a.db.GetOrCreateFilterGateway()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, gw)
	case http.MethodPut, http.MethodPost:
		var gw store.FilterGateway
		if err := json.NewDecoder(r.Body).Decode(&gw); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if err := a.db.UpdateFilterGateway(&gw); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		_ = a.manager.Reload(r.Context())
		writeJSON(w, http.StatusOK, gw)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *API) apps(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		apps, err := a.db.ListFilterApps()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, apps)
	case http.MethodPost:
		var app store.FilterApp
		if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		app.ID = 0
		if err := a.db.CreateFilterApp(&app); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		_ = a.manager.Reload(r.Context())
		writeJSON(w, http.StatusCreated, app)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *API) appByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseTrailingID(r.URL.Path, "/api/apps/")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	switch r.Method {
	case http.MethodPut, http.MethodPatch:
		var app store.FilterApp
		if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		app.ID = id
		if err := a.db.UpdateFilterApp(&app); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		_ = a.manager.Reload(r.Context())
		writeJSON(w, http.StatusOK, app)
	case http.MethodDelete:
		if err := a.db.DeleteFilterApp(id); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		_ = a.manager.Reload(r.Context())
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *API) templates(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ts, err := a.db.ListFilterTemplates()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, ts)
	case http.MethodPost:
		var t store.FilterTemplate
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		t.ID = 0
		if err := a.db.CreateFilterTemplate(&t); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusCreated, t)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *API) templateByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseTrailingID(r.URL.Path, "/api/templates/")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	switch r.Method {
	case http.MethodGet:
		t, err := a.db.GetFilterTemplate(id)
		if err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeJSON(w, http.StatusOK, t)
	case http.MethodPut, http.MethodPatch:
		var t store.FilterTemplate
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		t.ID = id
		if err := a.db.UpdateFilterTemplate(&t); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		_ = a.manager.Reload(r.Context())
		writeJSON(w, http.StatusOK, t)
	case http.MethodDelete:
		if err := a.db.DeleteFilterTemplate(id); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *API) recentEvents(w http.ResponseWriter, r *http.Request) {
	limit := 200
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	writeJSON(w, http.StatusOK, a.manager.RecentEvents(limit))
}

// eventStream is a server-sent events feed of filter events.
// Designed for Next.js EventSource consumption.
func (a *API) eventStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch, unsub := a.manager.Subscribe()
	defer unsub()

	enc := json.NewEncoder(w)
	flusher.Flush()
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				return
			}
			if _, err := w.Write([]byte("data: ")); err != nil {
				return
			}
			if err := enc.Encode(ev); err != nil {
				return
			}
			if _, err := w.Write([]byte("\n")); err != nil {
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (a *API) yamlExport(w http.ResponseWriter, r *http.Request) {
	gw, err := a.db.GetOrCreateFilterGateway()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	def, err := a.db.GetDefaultFilterTemplate()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	tps, err := a.db.ListFilterTemplates()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	apps, err := a.db.ListFilterApps()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out, err := filter.ExportYAML(gw, def, tps, apps)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Content-Disposition", `attachment; filename="filter.yaml"`)
	_, _ = w.Write(out)
}

func (a *API) yamlImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	defer body.Close()
	buf := make([]byte, 1<<20)
	n, err := readAll(body, buf)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	cfg, err := filter.ParseYAML(buf[:n])
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	gw, err := a.db.GetOrCreateFilterGateway()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	def, err := a.db.GetDefaultFilterTemplate()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	apps, _ := filter.ApplyYAMLToModels(cfg, gw, def)
	if err := a.db.UpdateFilterGateway(gw); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := a.db.UpdateFilterTemplate(def); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	for i := range apps {
		existing, err := a.db.GetFilterAppByName(apps[i].Name)
		if err == nil && existing != nil {
			apps[i].ID = existing.ID
			if err := a.db.UpdateFilterApp(&apps[i]); err != nil {
				log.Warn().Err(err).Str("name", apps[i].Name).Msg("yaml import: update app failed")
			}
			continue
		}
		if err := a.db.CreateFilterApp(&apps[i]); err != nil {
			log.Warn().Err(err).Str("name", apps[i].Name).Msg("yaml import: create app failed")
		}
	}
	if err := a.manager.Reload(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "apps": len(apps)})
}

func (a *API) regexTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Pattern string `json:"pattern"`
		Text    string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	compiled, matched, msg := filter.TestRegex(req.Pattern, req.Text)
	writeJSON(w, http.StatusOK, map[string]any{
		"compiled": compiled,
		"matched":  matched,
		"error":    msg,
	})
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"error": err.Error()})
}

func parseTrailingID(path, prefix string) (uint, error) {
	rest := strings.TrimPrefix(path, prefix)
	if rest == "" || rest == path {
		return 0, errors.New("missing id")
	}
	if i := strings.IndexByte(rest, '/'); i >= 0 {
		rest = rest[:i]
	}
	n, err := strconv.ParseUint(rest, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

func readAll(r interface{ Read([]byte) (int, error) }, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := r.Read(buf[total:])
		total += n
		if err != nil {
			if err.Error() == "EOF" {
				return total, nil
			}
			return total, err
		}
		if n == 0 {
			break
		}
	}
	return total, nil
}
