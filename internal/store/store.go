package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct { db *sql.DB }

type Asset struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	URL          string   `json:"url"`
	Description  string   `json:"description"`
	CreatedAt    string   `json:"created_at"`
}

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	dsn := filepath.Join(dataDir, "presskit.db") + "?_journal_mode=WAL&_busy_timeout=5000"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS assets (
			id TEXT PRIMARY KEY,\n\t\t\tname TEXT DEFAULT '',\n\t\t\ttype TEXT DEFAULT '',\n\t\t\turl TEXT DEFAULT '',\n\t\t\tdescription TEXT DEFAULT '',
			created_at TEXT DEFAULT (datetime('now'))
		)`)
	if err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }

func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }

func (d *DB) Create(e *Asset) error {
	e.ID = genID()
	e.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	_, err := d.db.Exec(`INSERT INTO assets (id, name, type, url, description, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		e.ID, e.Name, e.Type, e.URL, e.Description, e.CreatedAt)
	return err
}

func (d *DB) Get(id string) *Asset {
	row := d.db.QueryRow(`SELECT id, name, type, url, description, created_at FROM assets WHERE id=?`, id)
	var e Asset
	if err := row.Scan(&e.ID, &e.Name, &e.Type, &e.URL, &e.Description, &e.CreatedAt); err != nil {
		return nil
	}
	return &e
}

func (d *DB) List() []Asset {
	rows, err := d.db.Query(`SELECT id, name, type, url, description, created_at FROM assets ORDER BY created_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var result []Asset
	for rows.Next() {
		var e Asset
		if err := rows.Scan(&e.ID, &e.Name, &e.Type, &e.URL, &e.Description, &e.CreatedAt); err != nil {
			continue
		}
		result = append(result, e)
	}
	return result
}

func (d *DB) Delete(id string) error {
	_, err := d.db.Exec(`DELETE FROM assets WHERE id=?`, id)
	return err
}

func (d *DB) Count() int {
	var n int
	d.db.QueryRow(`SELECT COUNT(*) FROM assets`).Scan(&n)
	return n
}
