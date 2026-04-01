package server
import("encoding/json";"net/http";"strconv";"github.com/stockyard-dev/stockyard-outpost/internal/store")
func(s *Server)handleListBookmarks(w http.ResponseWriter,r *http.Request){cat:=r.URL.Query().Get("category");list,_:=s.db.ListBookmarks(cat);if list==nil{list=[]store.Bookmark{}};writeJSON(w,200,list)}
func(s *Server)handleCreateBookmark(w http.ResponseWriter,r *http.Request){var b store.Bookmark;json.NewDecoder(r.Body).Decode(&b);if b.Title==""||b.URL==""{writeError(w,400,"title and url required");return};if b.Category==""{b.Category="general"};if err:=s.db.CreateBookmark(&b);err!=nil{writeError(w,500,err.Error());return};writeJSON(w,201,b)}
func(s *Server)handleTogglePin(w http.ResponseWriter,r *http.Request){id,_:=strconv.ParseInt(r.PathValue("id"),10,64);s.db.TogglePin(id);writeJSON(w,200,map[string]string{"status":"toggled"})}
func(s *Server)handleDeleteBookmark(w http.ResponseWriter,r *http.Request){id,_:=strconv.ParseInt(r.PathValue("id"),10,64);s.db.DeleteBookmark(id);writeJSON(w,200,map[string]string{"status":"deleted"})}
func(s *Server)handleListWidgets(w http.ResponseWriter,r *http.Request){list,_:=s.db.ListWidgets();if list==nil{list=[]store.Widget{}};writeJSON(w,200,list)}
func(s *Server)handleCreateWidget(w http.ResponseWriter,r *http.Request){var wg store.Widget;json.NewDecoder(r.Body).Decode(&wg);if wg.WidgetType==""{writeError(w,400,"widget_type required");return};if wg.Config==""{wg.Config="{}"};if err:=s.db.CreateWidget(&wg);err!=nil{writeError(w,500,err.Error());return};writeJSON(w,201,wg)}
func(s *Server)handleDeleteWidget(w http.ResponseWriter,r *http.Request){id,_:=strconv.ParseInt(r.PathValue("id"),10,64);s.db.DeleteWidget(id);writeJSON(w,200,map[string]string{"status":"deleted"})}
func(s *Server)handleCategories(w http.ResponseWriter,r *http.Request){cats,_:=s.db.Categories();if cats==nil{cats=[]string{}};writeJSON(w,200,cats)}
func(s *Server)handleStats(w http.ResponseWriter,r *http.Request){n,_:=s.db.CountBookmarks();writeJSON(w,200,map[string]interface{}{"bookmarks":n})}
