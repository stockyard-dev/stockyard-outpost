package main
import ("fmt";"log";"net/http";"os";"github.com/stockyard-dev/stockyard-outpost/internal/server";"github.com/stockyard-dev/stockyard-outpost/internal/store")
func main(){port:=os.Getenv("PORT");if port==""{port="8740"};dataDir:=os.Getenv("DATA_DIR");if dataDir==""{dataDir="./outpost-data"}
db,err:=store.Open(dataDir);if err!=nil{log.Fatalf("outpost: %v",err)};defer db.Close();srv:=server.New(db,server.DefaultLimits())
fmt.Printf("\n  Outpost — Self-hosted remote server monitor\n  ─────────────────────────────────\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n  Report:     POST http://localhost:%s/api/report\n  Data:       %s\n  ─────────────────────────────────\n\n",port,port,port,dataDir)
log.Printf("outpost: listening on :%s",port);log.Fatal(http.ListenAndServe(":"+port,srv))}
