package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ db *sql.DB }

// Host is a single monitored machine. Status is one of: online, offline,
// unknown. The metric percentages are floats in the range 0..100.
type Host struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Hostname   string  `json:"hostname"`
	IP         string  `json:"ip,omitempty"`
	OS         string  `json:"os,omitempty"`
	Status     string  `json:"status"`
	CPUPct     float64 `json:"cpu_pct"`
	MemPct     float64 `json:"mem_pct"`
	DiskPct    float64 `json:"disk_pct"`
	Uptime     string  `json:"uptime,omitempty"`
	LastReport string  `json:"last_report,omitempty"`
	CreatedAt  string  `json:"created_at"`
}

func Open(d string) (*DB, error) {
	if err := os.MkdirAll(d, 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(d, "outpost.db")+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS hosts(
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		hostname TEXT DEFAULT '',
		ip TEXT DEFAULT '',
		os TEXT DEFAULT '',
		status TEXT DEFAULT 'unknown',
		cpu_pct REAL DEFAULT 0,
		mem_pct REAL DEFAULT 0,
		disk_pct REAL DEFAULT 0,
		uptime TEXT DEFAULT '',
		last_report TEXT DEFAULT '',
		created_at TEXT DEFAULT(datetime('now'))
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_hosts_status ON hosts(status)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_hosts_hostname ON hosts(hostname)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_hosts_last_report ON hosts(last_report)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS extras(
		resource TEXT NOT NULL,
		record_id TEXT NOT NULL,
		data TEXT NOT NULL DEFAULT '{}',
		PRIMARY KEY(resource, record_id)
	)`)
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }

func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string   { return time.Now().UTC().Format(time.RFC3339) }

// Register adds a host (typically called from the dashboard, not from
// agents). Agents use Report instead.
func (d *DB) Register(h *Host) error {
	h.ID = genID()
	h.CreatedAt = now()
	if h.Status == "" {
		h.Status = "unknown"
	}
	if h.LastReport == "" {
		h.LastReport = h.CreatedAt
	}
	_, err := d.db.Exec(
		`INSERT INTO hosts(id, name, hostname, ip, os, status, cpu_pct, mem_pct, disk_pct, uptime, last_report, created_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		h.ID, h.Name, h.Hostname, h.IP, h.OS, h.Status, h.CPUPct, h.MemPct, h.DiskPct, h.Uptime, h.LastReport, h.CreatedAt,
	)
	return err
}

func (d *DB) Get(id string) *Host {
	var h Host
	err := d.db.QueryRow(
		`SELECT id, name, hostname, ip, os, status, cpu_pct, mem_pct, disk_pct, uptime, last_report, created_at
		 FROM hosts WHERE id=?`,
		id,
	).Scan(&h.ID, &h.Name, &h.Hostname, &h.IP, &h.OS, &h.Status, &h.CPUPct, &h.MemPct, &h.DiskPct, &h.Uptime, &h.LastReport, &h.CreatedAt)
	if err != nil {
		return nil
	}
	return &h
}

func (d *DB) GetByHostname(hostname string) *Host {
	var h Host
	err := d.db.QueryRow(
		`SELECT id, name, hostname, ip, os, status, cpu_pct, mem_pct, disk_pct, uptime, last_report, created_at
		 FROM hosts WHERE hostname=?`,
		hostname,
	).Scan(&h.ID, &h.Name, &h.Hostname, &h.IP, &h.OS, &h.Status, &h.CPUPct, &h.MemPct, &h.DiskPct, &h.Uptime, &h.LastReport, &h.CreatedAt)
	if err != nil {
		return nil
	}
	return &h
}

func (d *DB) List() []Host {
	rows, _ := d.db.Query(
		`SELECT id, name, hostname, ip, os, status, cpu_pct, mem_pct, disk_pct, uptime, last_report, created_at
		 FROM hosts ORDER BY name ASC`,
	)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var o []Host
	for rows.Next() {
		var h Host
		rows.Scan(&h.ID, &h.Name, &h.Hostname, &h.IP, &h.OS, &h.Status, &h.CPUPct, &h.MemPct, &h.DiskPct, &h.Uptime, &h.LastReport, &h.CreatedAt)
		o = append(o, h)
	}
	return o
}

// Update applies admin changes to host metadata (name, hostname, ip,
// os). Status and metrics are managed by Report and MarkStale. The
// original implementation had no Update method at all.
func (d *DB) Update(h *Host) error {
	_, err := d.db.Exec(
		`UPDATE hosts SET name=?, hostname=?, ip=?, os=? WHERE id=?`,
		h.Name, h.Hostname, h.IP, h.OS, h.ID,
	)
	return err
}

func (d *DB) Delete(id string) error {
	_, err := d.db.Exec(`DELETE FROM hosts WHERE id=?`, id)
	return err
}

func (d *DB) Count() int {
	var n int
	d.db.QueryRow(`SELECT COUNT(*) FROM hosts`).Scan(&n)
	return n
}

// Report is the agent-facing endpoint: it upserts a host by hostname,
// stamping it as online and updating the four metric fields. Returns
// the resulting host record.
func (d *DB) Report(hostname string, cpu, mem, disk float64, uptime, ip, osName string) *Host {
	h := d.GetByHostname(hostname)
	t := now()
	if h == nil {
		h = &Host{
			ID:         genID(),
			Name:       hostname,
			Hostname:   hostname,
			IP:         ip,
			OS:         osName,
			Status:     "online",
			CPUPct:     cpu,
			MemPct:     mem,
			DiskPct:    disk,
			Uptime:     uptime,
			LastReport: t,
			CreatedAt:  t,
		}
		d.db.Exec(
			`INSERT INTO hosts(id, name, hostname, ip, os, status, cpu_pct, mem_pct, disk_pct, uptime, last_report, created_at)
			 VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			h.ID, h.Name, h.Hostname, h.IP, h.OS, h.Status, h.CPUPct, h.MemPct, h.DiskPct, h.Uptime, h.LastReport, h.CreatedAt,
		)
		return h
	}
	d.db.Exec(
		`UPDATE hosts SET status='online', cpu_pct=?, mem_pct=?, disk_pct=?, uptime=?, ip=?, os=?, last_report=? WHERE id=?`,
		cpu, mem, disk, uptime, ip, osName, t, h.ID,
	)
	h.CPUPct = cpu
	h.MemPct = mem
	h.DiskPct = disk
	h.Uptime = uptime
	if ip != "" {
		h.IP = ip
	}
	if osName != "" {
		h.OS = osName
	}
	h.Status = "online"
	h.LastReport = t
	return h
}

// MarkStale flips any host whose last_report is older than timeoutSec
// to offline. Called opportunistically by the server before listing.
func (d *DB) MarkStale(timeoutSec int) int {
	if timeoutSec <= 0 {
		timeoutSec = 120
	}
	cutoff := time.Now().Add(-time.Duration(timeoutSec) * time.Second).UTC().Format(time.RFC3339)
	res, err := d.db.Exec(
		`UPDATE hosts SET status='offline' WHERE last_report<? AND status='online'`,
		cutoff,
	)
	if err != nil {
		return 0
	}
	n, _ := res.RowsAffected()
	return int(n)
}

// Stats returns total hosts, counts by status, and the worst observed
// CPU/mem/disk values. The original had no Stats method at all.
func (d *DB) Stats() map[string]any {
	m := map[string]any{
		"total":        d.Count(),
		"online":       0,
		"offline":      0,
		"by_status":    map[string]int{},
		"max_cpu_pct":  0.0,
		"max_mem_pct":  0.0,
		"max_disk_pct": 0.0,
	}

	var online, offline int
	d.db.QueryRow(`SELECT COUNT(*) FROM hosts WHERE status='online'`).Scan(&online)
	d.db.QueryRow(`SELECT COUNT(*) FROM hosts WHERE status='offline'`).Scan(&offline)
	m["online"] = online
	m["offline"] = offline

	var maxCPU, maxMem, maxDisk float64
	d.db.QueryRow(`SELECT COALESCE(MAX(cpu_pct), 0) FROM hosts WHERE status='online'`).Scan(&maxCPU)
	d.db.QueryRow(`SELECT COALESCE(MAX(mem_pct), 0) FROM hosts WHERE status='online'`).Scan(&maxMem)
	d.db.QueryRow(`SELECT COALESCE(MAX(disk_pct), 0) FROM hosts WHERE status='online'`).Scan(&maxDisk)
	m["max_cpu_pct"] = maxCPU
	m["max_mem_pct"] = maxMem
	m["max_disk_pct"] = maxDisk

	if rows, _ := d.db.Query(`SELECT status, COUNT(*) FROM hosts GROUP BY status`); rows != nil {
		defer rows.Close()
		by := map[string]int{}
		for rows.Next() {
			var s string
			var c int
			rows.Scan(&s, &c)
			by[s] = c
		}
		m["by_status"] = by
	}

	return m
}

// ─── Extras ───────────────────────────────────────────────────────

func (d *DB) GetExtras(resource, recordID string) string {
	var data string
	err := d.db.QueryRow(
		`SELECT data FROM extras WHERE resource=? AND record_id=?`,
		resource, recordID,
	).Scan(&data)
	if err != nil || data == "" {
		return "{}"
	}
	return data
}

func (d *DB) SetExtras(resource, recordID, data string) error {
	if data == "" {
		data = "{}"
	}
	_, err := d.db.Exec(
		`INSERT INTO extras(resource, record_id, data) VALUES(?, ?, ?)
		 ON CONFLICT(resource, record_id) DO UPDATE SET data=excluded.data`,
		resource, recordID, data,
	)
	return err
}

func (d *DB) DeleteExtras(resource, recordID string) error {
	_, err := d.db.Exec(
		`DELETE FROM extras WHERE resource=? AND record_id=?`,
		resource, recordID,
	)
	return err
}

func (d *DB) AllExtras(resource string) map[string]string {
	out := make(map[string]string)
	rows, _ := d.db.Query(
		`SELECT record_id, data FROM extras WHERE resource=?`,
		resource,
	)
	if rows == nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id, data string
		rows.Scan(&id, &data)
		out[id] = data
	}
	return out
}
