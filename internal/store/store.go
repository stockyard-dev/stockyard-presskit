package store

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"time"
)

type DB struct{ db *sql.DB }
type Asset struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	FileType    string `json:"file_type,omitempty"`
	Public      bool   `json:"public"`
	CreatedAt   string `json:"created_at"`
}
type Fact struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Value     string `json:"value"`
	Position  int    `json:"position"`
	CreatedAt string `json:"created_at"`
}
type Contact struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Role      string `json:"role,omitempty"`
	Email     string `json:"email,omitempty"`
	Bio       string `json:"bio,omitempty"`
	CreatedAt string `json:"created_at"`
}

func Open(d string) (*DB, error) {
	if err := os.MkdirAll(d, 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(d, "presskit.db")+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	for _, q := range []string{
		`CREATE TABLE IF NOT EXISTS assets(id TEXT PRIMARY KEY,name TEXT NOT NULL,category TEXT DEFAULT '',description TEXT DEFAULT '',url TEXT DEFAULT '',file_type TEXT DEFAULT '',public INTEGER DEFAULT 1,created_at TEXT DEFAULT(datetime('now')))`,
		`CREATE TABLE IF NOT EXISTS facts(id TEXT PRIMARY KEY,label TEXT NOT NULL,value TEXT NOT NULL,position INTEGER DEFAULT 0,created_at TEXT DEFAULT(datetime('now')))`,
		`CREATE TABLE IF NOT EXISTS contacts(id TEXT PRIMARY KEY,name TEXT NOT NULL,role TEXT DEFAULT '',email TEXT DEFAULT '',bio TEXT DEFAULT '',created_at TEXT DEFAULT(datetime('now')))`,
	} {
		if _, err := db.Exec(q); err != nil {
			return nil, fmt.Errorf("migrate: %w", err)
		}
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS extras(resource TEXT NOT NULL,record_id TEXT NOT NULL,data TEXT NOT NULL DEFAULT '{}',PRIMARY KEY(resource, record_id))`)
	return &DB{db: db}, nil
}
func (d *DB) Close() error { return d.db.Close() }
func genID() string        { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string          { return time.Now().UTC().Format(time.RFC3339) }
func (d *DB) CreateAsset(a *Asset) error {
	a.ID = genID()
	a.CreatedAt = now()
	pub := 1
	if !a.Public {
		pub = 0
	}
	_, err := d.db.Exec(`INSERT INTO assets VALUES(?,?,?,?,?,?,?,?)`, a.ID, a.Name, a.Category, a.Description, a.URL, a.FileType, pub, a.CreatedAt)
	return err
}
func (d *DB) ListAssets() []Asset {
	rows, _ := d.db.Query(`SELECT * FROM assets ORDER BY category,name`)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var o []Asset
	for rows.Next() {
		var a Asset
		var pub int
		rows.Scan(&a.ID, &a.Name, &a.Category, &a.Description, &a.URL, &a.FileType, &pub, &a.CreatedAt)
		a.Public = pub == 1
		o = append(o, a)
	}
	return o
}
func (d *DB) DeleteAsset(id string) error {
	_, err := d.db.Exec(`DELETE FROM assets WHERE id=?`, id)
	return err
}
func (d *DB) CreateFact(f *Fact) error {
	f.ID = genID()
	f.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO facts VALUES(?,?,?,?,?)`, f.ID, f.Label, f.Value, f.Position, f.CreatedAt)
	return err
}
func (d *DB) ListFacts() []Fact {
	rows, _ := d.db.Query(`SELECT * FROM facts ORDER BY position,label`)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var o []Fact
	for rows.Next() {
		var f Fact
		rows.Scan(&f.ID, &f.Label, &f.Value, &f.Position, &f.CreatedAt)
		o = append(o, f)
	}
	return o
}
func (d *DB) DeleteFact(id string) error {
	_, err := d.db.Exec(`DELETE FROM facts WHERE id=?`, id)
	return err
}
func (d *DB) CreateContact(c *Contact) error {
	c.ID = genID()
	c.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO contacts VALUES(?,?,?,?,?,?)`, c.ID, c.Name, c.Role, c.Email, c.Bio, c.CreatedAt)
	return err
}
func (d *DB) ListContacts() []Contact {
	rows, _ := d.db.Query(`SELECT * FROM contacts ORDER BY name`)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var o []Contact
	for rows.Next() {
		var c Contact
		rows.Scan(&c.ID, &c.Name, &c.Role, &c.Email, &c.Bio, &c.CreatedAt)
		o = append(o, c)
	}
	return o
}
func (d *DB) DeleteContact(id string) error {
	_, err := d.db.Exec(`DELETE FROM contacts WHERE id=?`, id)
	return err
}

type Stats struct {
	Assets   int `json:"assets"`
	Facts    int `json:"facts"`
	Contacts int `json:"contacts"`
}

func (d *DB) Stats() Stats {
	var s Stats
	d.db.QueryRow(`SELECT COUNT(*) FROM assets`).Scan(&s.Assets)
	d.db.QueryRow(`SELECT COUNT(*) FROM facts`).Scan(&s.Facts)
	d.db.QueryRow(`SELECT COUNT(*) FROM contacts`).Scan(&s.Contacts)
	return s
}
func (d *DB) PublicKit() map[string]any {
	return map[string]any{"assets": d.ListAssets(), "facts": d.ListFacts(), "contacts": d.ListContacts()}
}

// ─── Extras: generic key-value storage for personalization custom fields ───

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
