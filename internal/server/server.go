package server
import ("encoding/json";"log";"net/http";"github.com/stockyard-dev/stockyard-outpost/internal/store")
type Server struct{db *store.DB;mux *http.ServeMux;limits Limits}
func New(db *store.DB,limits Limits)*Server{s:=&Server{db:db,mux:http.NewServeMux(),limits:limits}
s.mux.HandleFunc("GET /api/hosts",s.list);s.mux.HandleFunc("POST /api/hosts",s.register);s.mux.HandleFunc("GET /api/hosts/{id}",s.get);s.mux.HandleFunc("DELETE /api/hosts/{id}",s.del)
s.mux.HandleFunc("POST /api/report",s.report)
s.mux.HandleFunc("GET /api/stats",s.stats);s.mux.HandleFunc("GET /api/health",s.health)
s.mux.HandleFunc("GET /ui",s.dashboard);s.mux.HandleFunc("GET /ui/",s.dashboard);s.mux.HandleFunc("GET /",s.root);return s}
func(s *Server)ServeHTTP(w http.ResponseWriter,r *http.Request){s.mux.ServeHTTP(w,r)}
func wj(w http.ResponseWriter,c int,v any){w.Header().Set("Content-Type","application/json");w.WriteHeader(c);json.NewEncoder(w).Encode(v)}
func we(w http.ResponseWriter,c int,m string){wj(w,c,map[string]string{"error":m})}
func(s *Server)root(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/"{http.NotFound(w,r);return};http.Redirect(w,r,"/ui",302)}
func(s *Server)list(w http.ResponseWriter,r *http.Request){s.db.MarkStale(120);wj(w,200,map[string]any{"hosts":oe(s.db.List())})}
func(s *Server)register(w http.ResponseWriter,r *http.Request){var h store.Host;json.NewDecoder(r.Body).Decode(&h);if h.Name==""{we(w,400,"name required");return};s.db.Register(&h);wj(w,201,h)}
func(s *Server)get(w http.ResponseWriter,r *http.Request){h:=s.db.Get(r.PathValue("id"));if h==nil{we(w,404,"not found");return};wj(w,200,h)}
func(s *Server)del(w http.ResponseWriter,r *http.Request){s.db.Delete(r.PathValue("id"));wj(w,200,map[string]string{"deleted":"ok"})}
func(s *Server)report(w http.ResponseWriter,r *http.Request){var req struct{Hostname string `json:"hostname"`;CPU float64 `json:"cpu_pct"`;Mem float64 `json:"mem_pct"`;Disk float64 `json:"disk_pct"`;Uptime string `json:"uptime"`;IP string `json:"ip"`;OS string `json:"os"`}
json.NewDecoder(r.Body).Decode(&req);if req.Hostname==""{we(w,400,"hostname required");return}
h:=s.db.Report(req.Hostname,req.CPU,req.Mem,req.Disk,req.Uptime,req.IP,req.OS);wj(w,200,h)}
func(s *Server)stats(w http.ResponseWriter,r *http.Request){wj(w,200,s.db.Stats())}
func(s *Server)health(w http.ResponseWriter,r *http.Request){st:=s.db.Stats();wj(w,200,map[string]any{"status":"ok","service":"outpost","hosts":st.Total,"online":st.Online})}
func oe[T any](s []T)[]T{if s==nil{return[]T{}};return s}
func init(){log.SetFlags(log.LstdFlags|log.Lshortfile)}
