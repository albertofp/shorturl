package sqlite

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDB struct {
	Db *sql.DB
}

func New(path string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		shorturl TEXT UNIQUE NOT NULL,
		longurl TEXT NOT NULL,
		createdAt TEXT NOT NULL,
		ttl TEXT NOT NULL
	)`
	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}
	return &SQLiteDB{Db: db}, nil
}

// SaveShortURL stores a new short URL in the database.
func (s *SQLiteDB) SaveShortURL(ctx context.Context, shortURL, longURL string, createdAt, ttl time.Time) (int64, error) {
	r, err := s.Db.ExecContext(ctx, `INSERT INTO urls (shorturl, longurl, createdAt, ttl) VALUES (?, ?, ?, ?)`,
		shortURL, longURL, createdAt.Format(time.RFC3339), ttl.Format(time.RFC3339))
	if err != nil {
		return 0, err
	}

	id, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetLongURL retrieves the original long URL by its short URL ID.
func (s *SQLiteDB) GetLongURL(ctx context.Context, shortURL string) (string, time.Time, time.Time, error) {
	var longURL string
	var createdAtStr string
	var ttlStr string

	err := s.Db.QueryRowContext(ctx, `SELECT longurl, createdAt, ttl FROM urls WHERE shorturl = ?`, shortURL).
		Scan(&longURL, &createdAtStr, &ttlStr)
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}

	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}

	ttl, err := time.Parse(time.RFC3339, ttlStr)
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}

	return longURL, createdAt, ttl, nil
}

// Close closes the SQLite database connection.
func (s *SQLiteDB) Close() error {
	return s.Db.Close()
}
