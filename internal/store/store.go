package store
import ("database/sql";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type Host struct{ID string `json:"id"`;Name string `json:"name"`;Hostname string `json:"hostname"`;IP string `json:"ip,omitempty"`;OS string `json:"os,omitempty"`;Status string `json:"status"`;CPUPct float64 `json:"cpu_pct"`;MemPct float64 `json:"mem_pct"`;DiskPct float64 `json:"disk_pct"`;Uptime string `json:"uptime,omitempty"`;LastReport string `json:"last_report,omitempty"`;CreatedAt string `json:"created_at"`}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"outpost.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
db.Exec(`CREATE TABLE IF NOT EXISTS hosts(id TEXT PRIMARY KEY,name TEXT NOT NULL,hostname TEXT DEFAULT '',ip TEXT DEFAULT '',os TEXT DEFAULT '',status TEXT DEFAULT 'unknown',cpu_pct REAL DEFAULT 0,mem_pct REAL DEFAULT 0,disk_pct REAL DEFAULT 0,uptime TEXT DEFAULT '',last_report TEXT DEFAULT '',created_at TEXT DEFAULT(datetime('now')))`)
return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)Register(h *Host)error{h.ID=genID();h.CreatedAt=now();h.Status="online";h.LastReport=now()
_,err:=d.db.Exec(`INSERT INTO hosts VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,h.ID,h.Name,h.Hostname,h.IP,h.OS,h.Status,h.CPUPct,h.MemPct,h.DiskPct,h.Uptime,h.LastReport,h.CreatedAt);return err}
func(d *DB)Get(id string)*Host{var h Host;if d.db.QueryRow(`SELECT * FROM hosts WHERE id=?`,id).Scan(&h.ID,&h.Name,&h.Hostname,&h.IP,&h.OS,&h.Status,&h.CPUPct,&h.MemPct,&h.DiskPct,&h.Uptime,&h.LastReport,&h.CreatedAt)!=nil{return nil};return &h}
func(d *DB)GetByHostname(hostname string)*Host{var h Host;if d.db.QueryRow(`SELECT * FROM hosts WHERE hostname=?`,hostname).Scan(&h.ID,&h.Name,&h.Hostname,&h.IP,&h.OS,&h.Status,&h.CPUPct,&h.MemPct,&h.DiskPct,&h.Uptime,&h.LastReport,&h.CreatedAt)!=nil{return nil};return &h}
func(d *DB)List()[]Host{rows,_:=d.db.Query(`SELECT * FROM hosts ORDER BY name`);if rows==nil{return nil};defer rows.Close()
var o []Host;for rows.Next(){var h Host;rows.Scan(&h.ID,&h.Name,&h.Hostname,&h.IP,&h.OS,&h.Status,&h.CPUPct,&h.MemPct,&h.DiskPct,&h.Uptime,&h.LastReport,&h.CreatedAt);o=append(o,h)};return o}
func(d *DB)Delete(id string)error{_,err:=d.db.Exec(`DELETE FROM hosts WHERE id=?`,id);return err}
func(d *DB)Report(hostname string,cpu,mem,disk float64,uptime,ip,osName string)*Host{
h:=d.GetByHostname(hostname);t:=now()
if h==nil{h=&Host{ID:genID(),Name:hostname,Hostname:hostname,IP:ip,OS:osName,Status:"online",CPUPct:cpu,MemPct:mem,DiskPct:disk,Uptime:uptime,LastReport:t,CreatedAt:t}
d.db.Exec(`INSERT INTO hosts VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,h.ID,h.Name,h.Hostname,h.IP,h.OS,h.Status,h.CPUPct,h.MemPct,h.DiskPct,h.Uptime,h.LastReport,h.CreatedAt);return h}
d.db.Exec(`UPDATE hosts SET status='online',cpu_pct=?,mem_pct=?,disk_pct=?,uptime=?,ip=?,os=?,last_report=? WHERE id=?`,cpu,mem,disk,uptime,ip,osName,t,h.ID)
h.CPUPct=cpu;h.MemPct=mem;h.DiskPct=disk;h.Status="online";h.LastReport=t;return h}
func(d *DB)MarkStale(timeoutSec int){if timeoutSec<=0{timeoutSec=120};cutoff:=time.Now().Add(-time.Duration(timeoutSec)*time.Second).UTC().Format(time.RFC3339)
d.db.Exec(`UPDATE hosts SET status='offline' WHERE last_report<? AND status='online'`,cutoff)}
type Stats struct{Total int `json:"total"`;Online int `json:"online"`;Offline int `json:"offline"`}
func(d *DB)Stats()Stats{d.MarkStale(120);var s Stats;d.db.QueryRow(`SELECT COUNT(*) FROM hosts`).Scan(&s.Total);d.db.QueryRow(`SELECT COUNT(*) FROM hosts WHERE status='online'`).Scan(&s.Online);d.db.QueryRow(`SELECT COUNT(*) FROM hosts WHERE status='offline'`).Scan(&s.Offline);return s}
