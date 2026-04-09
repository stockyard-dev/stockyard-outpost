package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/stockyard-dev/stockyard-outpost/internal/store"
)

const resourceName = "hosts"

type Server struct {
	db         *store.DB
	mux        *http.ServeMux
	limMu      sync.RWMutex // guards limits, hot-reloadable by /api/license/activate
	limits     Limits
	dataDir    string
	pCfg       map[string]json.RawMessage
	staleAfter int
}

func New(db *store.DB, limits Limits, dataDir string) *Server {
	s := &Server{
		db:         db,
		mux:        http.NewServeMux(),
		limits:     limits,
		dataDir:    dataDir,
		staleAfter: 120,
	}
	s.loadPersonalConfig()

	// Hosts CRUD (admin)
	s.mux.HandleFunc("GET /api/hosts", s.list)
	s.mux.HandleFunc("POST /api/hosts", s.register)
	s.mux.HandleFunc("GET /api/hosts/{id}", s.get)
	s.mux.HandleFunc("PUT /api/hosts/{id}", s.update) // NEW
	s.mux.HandleFunc("DELETE /api/hosts/{id}", s.del)

	// Agent-facing report endpoint
	s.mux.HandleFunc("POST /api/report", s.report)

	// Stats / health
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)

	// Personalization
	s.mux.HandleFunc("GET /api/config", s.configHandler)

	// Extras
	s.mux.HandleFunc("GET /api/extras/{resource}", s.listExtras)
	s.mux.HandleFunc("GET /api/extras/{resource}/{id}", s.getExtras)
	s.mux.HandleFunc("PUT /api/extras/{resource}/{id}", s.putExtras)

	// License activation — accepts a key, validates, persists, hot-reloads tier
	s.mux.HandleFunc("POST /api/license/activate", s.activateLicense)

	// Tier — read-only license info for dashboard banner. Always reachable.
	s.mux.HandleFunc("GET /api/tier", s.tierInfo)

	// Dashboard
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)

	return s
}

// ServeHTTP wraps the underlying mux with a license-gate middleware.
// In trial-required mode, all writes (POST/PUT/DELETE/PATCH) return 402
// EXCEPT POST /api/license/activate (the only way out of trial state).
// Reads are always allowed — the brand promise is that data on disk
// stays accessible even without an active license.
//
// Outpost also runs opportunistic stale detection on every list/stats
// GET request — that runs after the trial check so reads still get
// fresh stale-marking even when the tool is locked.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.shouldBlockWrite(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		w.Write([]byte(`{"error":"Trial required. Start a 14-day free trial at https://stockyard.dev/ — or paste an existing license key in the dashboard under \"Activate License\".","tier":"trial-required"}`))
		return
	}
	// Opportunistic stale detection on every list/stats request.
	// The index on last_report makes the UPDATE cheap.
	if (r.URL.Path == "/api/hosts" || r.URL.Path == "/api/stats") && r.Method == http.MethodGet {
		s.db.MarkStale(s.staleAfter)
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) shouldBlockWrite(r *http.Request) bool {
	s.limMu.RLock()
	tier := s.limits.Tier
	s.limMu.RUnlock()
	if tier != "trial-required" {
		return false
	}
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	}
	switch r.URL.Path {
	case "/api/license/activate":
		return false
	}
	return true
}

func (s *Server) activateLicense(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 10*1024))
	if err != nil {
		we(w, 400, "could not read request body")
		return
	}
	var req struct {
		LicenseKey string `json:"license_key"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		we(w, 400, "invalid json: "+err.Error())
		return
	}
	key := strings.TrimSpace(req.LicenseKey)
	if key == "" {
		we(w, 400, "license_key is required")
		return
	}
	if !ValidateLicenseKey(key) {
		we(w, 400, "license key is not valid for this product — make sure you copied the entire key from the welcome email, including the SY- prefix")
		return
	}
	if err := PersistLicense(s.dataDir, key); err != nil {
		log.Printf("outpost: license persist failed: %v", err)
		we(w, 500, "could not save the license key to disk: "+err.Error())
		return
	}
	s.limMu.Lock()
	s.limits = ProLimits()
	s.limMu.Unlock()
	log.Printf("outpost: license activated via dashboard, persisted to %s/%s", s.dataDir, licenseFilename)
	wj(w, 200, map[string]any{
		"ok":   true,
		"tier": "pro",
	})
}

func (s *Server) tierInfo(w http.ResponseWriter, r *http.Request) {
	s.limMu.RLock()
	tier := s.limits.Tier
	s.limMu.RUnlock()
	resp := map[string]any{
		"tier": tier,
	}
	if tier == "trial-required" {
		resp["trial_required"] = true
		resp["start_trial_url"] = "https://stockyard.dev/"
		resp["message"] = "Your trial is not active. Reads work, but you cannot register or change hosts until you start a 14-day trial or activate an existing license key."
	} else {
		resp["trial_required"] = false
	}
	wj(w, 200, resp)
}

// ─── helpers ──────────────────────────────────────────────────────

func wj(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func we(w http.ResponseWriter, code int, msg string) {
	wj(w, code, map[string]string{"error": msg})
}

func oe[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

func (s *Server) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/ui", 302)
}

// ─── personalization ──────────────────────────────────────────────

func (s *Server) loadPersonalConfig() {
	path := filepath.Join(s.dataDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("outpost: warning: could not parse config.json: %v", err)
		return
	}
	s.pCfg = cfg

	// Optional: stale_after_seconds in the config can override the default
	if v, ok := cfg["stale_after_seconds"]; ok {
		var n int
		if err := json.Unmarshal(v, &n); err == nil && n > 0 {
			s.staleAfter = n
		}
	}

	log.Printf("outpost: loaded personalization from %s", path)
}

func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	if s.pCfg == nil {
		wj(w, 200, map[string]any{})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.pCfg)
}

// ─── extras ───────────────────────────────────────────────────────

func (s *Server) listExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	all := s.db.AllExtras(resource)
	out := make(map[string]json.RawMessage, len(all))
	for id, data := range all {
		out[id] = json.RawMessage(data)
	}
	wj(w, 200, out)
}

func (s *Server) getExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	id := r.PathValue("id")
	data := s.db.GetExtras(resource, id)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(data))
}

func (s *Server) putExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	id := r.PathValue("id")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		we(w, 400, "read body")
		return
	}
	var probe map[string]any
	if err := json.Unmarshal(body, &probe); err != nil {
		we(w, 400, "invalid json")
		return
	}
	if err := s.db.SetExtras(resource, id, string(body)); err != nil {
		we(w, 500, "save failed")
		return
	}
	wj(w, 200, map[string]string{"ok": "saved"})
}

// ─── hosts ────────────────────────────────────────────────────────

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, map[string]any{"hosts": oe(s.db.List())})
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var h store.Host
	if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
		we(w, 400, "invalid json")
		return
	}
	if h.Name == "" {
		we(w, 400, "name required")
		return
	}
	if h.Hostname == "" {
		h.Hostname = h.Name
	}
	if err := s.db.Register(&h); err != nil {
		we(w, 500, "register failed")
		return
	}
	wj(w, 201, s.db.Get(h.ID))
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	h := s.db.Get(r.PathValue("id"))
	if h == nil {
		we(w, 404, "not found")
		return
	}
	wj(w, 200, h)
}

// update accepts a partial host metadata patch (name, hostname, ip, os).
// Status and metric fields are managed by Report and MarkStale, not by
// the dashboard. The original implementation had no PUT endpoint at all.
func (s *Server) update(w http.ResponseWriter, r *http.Request) {
	existing := s.db.Get(r.PathValue("id"))
	if existing == nil {
		we(w, 404, "not found")
		return
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		we(w, 400, "invalid json")
		return
	}

	patch := *existing
	if v, ok := raw["name"]; ok {
		var s string
		json.Unmarshal(v, &s)
		if s != "" {
			patch.Name = s
		}
	}
	if v, ok := raw["hostname"]; ok {
		var s string
		json.Unmarshal(v, &s)
		if s != "" {
			patch.Hostname = s
		}
	}
	if v, ok := raw["ip"]; ok {
		json.Unmarshal(v, &patch.IP)
	}
	if v, ok := raw["os"]; ok {
		json.Unmarshal(v, &patch.OS)
	}

	if err := s.db.Update(&patch); err != nil {
		we(w, 500, "update failed")
		return
	}
	wj(w, 200, s.db.Get(patch.ID))
}

func (s *Server) del(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.db.Delete(id)
	s.db.DeleteExtras(resourceName, id)
	wj(w, 200, map[string]string{"deleted": "ok"})
}

// report is the agent endpoint. Agents POST their current metrics here
// and outpost upserts the host. Returns the resulting record.
func (s *Server) report(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Hostname string  `json:"hostname"`
		CPUPct   float64 `json:"cpu_pct"`
		MemPct   float64 `json:"mem_pct"`
		DiskPct  float64 `json:"disk_pct"`
		Uptime   string  `json:"uptime"`
		IP       string  `json:"ip"`
		OS       string  `json:"os"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		we(w, 400, "invalid json")
		return
	}
	if req.Hostname == "" {
		we(w, 400, "hostname required")
		return
	}
	h := s.db.Report(req.Hostname, req.CPUPct, req.MemPct, req.DiskPct, req.Uptime, req.IP, req.OS)
	wj(w, 200, h)
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, s.db.Stats())
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, map[string]any{
		"status":  "ok",
		"service": "outpost",
		"hosts":   s.db.Count(),
	})
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
