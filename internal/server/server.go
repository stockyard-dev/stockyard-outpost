package server
import("encoding/json";"net/http";"github.com/stockyard-dev/stockyard-outpost/internal/store")
type Server struct{db *store.DB;limits Limits;mux *http.ServeMux}
func New(db *store.DB,tier string)*Server{s:=&Server{db:db,limits:LimitsFor(tier),mux:http.NewServeMux()};s.routes();return s}
func(s *Server)ListenAndServe(addr string)error{return(&http.Server{Addr:addr,Handler:s.mux}).ListenAndServe()}
func(s *Server)routes(){
    s.mux.HandleFunc("GET /health",s.handleHealth)
    s.mux.HandleFunc("GET /api/stats",s.handleStats)
    s.mux.HandleFunc("GET /api/bookmarks",s.handleListBookmarks)
    s.mux.HandleFunc("POST /api/bookmarks",s.handleCreateBookmark)
    s.mux.HandleFunc("POST /api/bookmarks/{id}/pin",s.handleTogglePin)
    s.mux.HandleFunc("DELETE /api/bookmarks/{id}",s.handleDeleteBookmark)
    s.mux.HandleFunc("GET /api/categories",s.handleCategories)
    s.mux.HandleFunc("GET /api/widgets",s.handleListWidgets)
    s.mux.HandleFunc("POST /api/widgets",s.handleCreateWidget)
    s.mux.HandleFunc("DELETE /api/widgets/{id}",s.handleDeleteWidget)
    s.mux.HandleFunc("GET /",s.handleUI)
}
func(s *Server)handleHealth(w http.ResponseWriter,r *http.Request){writeJSON(w,200,map[string]string{"status":"ok","service":"stockyard-outpost"})}
func writeJSON(w http.ResponseWriter,status int,v interface{}){w.Header().Set("Content-Type","application/json");w.WriteHeader(status);json.NewEncoder(w).Encode(v)}
func writeError(w http.ResponseWriter,status int,msg string){writeJSON(w,status,map[string]string{"error":msg})}
func(s *Server)handleUI(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/"{http.NotFound(w,r);return};w.Header().Set("Content-Type","text/html");w.Write(dashboardHTML)}
