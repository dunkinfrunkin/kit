package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Item struct {
	ID        int       `json:"id"`
	Namespace string    `json:"namespace"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	Content   []byte    `json:"-"`
	Author    string    `json:"author"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Profile struct {
	ID        int          `json:"id"`
	Namespace string       `json:"namespace"`
	Name      string       `json:"name"`
	Items     []ProfileRef `json:"items"`
	Author    string       `json:"author"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type ProfileRef struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type APITokenRow struct {
	Name      string    `json:"name"`
	Prefix    string    `json:"prefix"`
	CreatedAt time.Time `json:"created_at"`
}

type Event struct {
	ID        int       `json:"id"`
	Action    string    `json:"action"`
	Namespace string    `json:"namespace"`
	ItemType  string    `json:"item_type,omitempty"`
	ItemName  string    `json:"item_name,omitempty"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type ItemStats struct {
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Count     int    `json:"count"`
}

type MetricsSummary struct {
	TotalItems    int         `json:"total_items"`
	TotalPushes   int         `json:"total_pushes"`
	TotalInstalls int         `json:"total_installs"`
	TopInstalled  []ItemStats `json:"top_installed"`
	TopPushed     []ItemStats `json:"top_pushed"`
	RecentEvents  []Event     `json:"recent_events"`
	ActiveUsers   int         `json:"active_users"`
}

type ItemVersion struct {
	Version   int       `json:"version"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	Size      int       `json:"size"`
}

type Store struct {
	db *sql.DB
}

var (
	ErrNotFound      = errors.New("not found")
	ErrNotAuthorized = errors.New("not authorized")
)

func New(databaseURL string) (*Store, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS items (
			id         SERIAL PRIMARY KEY,
			namespace  TEXT NOT NULL,
			type       TEXT NOT NULL,
			name       TEXT NOT NULL,
			content    BYTEA NOT NULL,
			author     TEXT NOT NULL,
			version    INTEGER DEFAULT 1,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(namespace, type, name)
		);

		CREATE TABLE IF NOT EXISTS profiles (
			id         SERIAL PRIMARY KEY,
			namespace  TEXT NOT NULL,
			name       TEXT NOT NULL,
			items      JSONB NOT NULL,
			author     TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(namespace, name)
		);

		CREATE TABLE IF NOT EXISTS api_tokens (
			id         SERIAL PRIMARY KEY,
			email      TEXT NOT NULL,
			name       TEXT NOT NULL,
			token_hash TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(email, name)
		);

		CREATE TABLE IF NOT EXISTS events (
			id         SERIAL PRIMARY KEY,
			action     TEXT NOT NULL,
			namespace  TEXT NOT NULL,
			item_type  TEXT,
			item_name  TEXT,
			email      TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_events_created_at ON events(created_at);
		CREATE INDEX IF NOT EXISTS idx_events_action ON events(action);

		CREATE TABLE IF NOT EXISTS item_versions (
			id         SERIAL PRIMARY KEY,
			namespace  TEXT NOT NULL,
			type       TEXT NOT NULL,
			name       TEXT NOT NULL,
			version    INTEGER NOT NULL,
			content    BYTEA NOT NULL,
			author     TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_item_versions_lookup ON item_versions(namespace, type, name, version DESC);
	`)
	return err
}

func (s *Store) PushItem(namespace, itemType, name string, content []byte, author string) (*Item, error) {
	var item Item
	err := s.db.QueryRow(`
		INSERT INTO items (namespace, type, name, content, author)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (namespace, type, name) DO UPDATE SET
			content    = EXCLUDED.content,
			author     = EXCLUDED.author,
			version    = items.version + 1,
			updated_at = NOW()
		RETURNING id, namespace, type, name, content, author, version, created_at, updated_at
	`, namespace, itemType, name, content, author).Scan(
		&item.ID, &item.Namespace, &item.Type, &item.Name,
		&item.Content, &item.Author, &item.Version,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pushing item: %w", err)
	}

	// Best-effort: record version history
	_, _ = s.db.Exec(`
		INSERT INTO item_versions (namespace, type, name, version, content, author)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, namespace, itemType, name, item.Version, content, author)

	return &item, nil
}

func (s *Store) GetItem(namespace, itemType, name string) (*Item, error) {
	var item Item
	err := s.db.QueryRow(`
		SELECT id, namespace, type, name, content, author, version, created_at, updated_at
		FROM items
		WHERE namespace = $1 AND type = $2 AND name = $3
	`, namespace, itemType, name).Scan(
		&item.ID, &item.Namespace, &item.Type, &item.Name,
		&item.Content, &item.Author, &item.Version,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting item: %w", err)
	}
	return &item, nil
}

func (s *Store) ListItems(namespace, itemType string) ([]Item, error) {
	query := `SELECT id, namespace, type, name, author, version, created_at, updated_at FROM items WHERE namespace = $1`
	args := []any{namespace}

	if itemType != "" {
		query += ` AND type = $2`
		args = append(args, itemType)
	}
	query += ` ORDER BY name`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing items: %w", err)
	}
	defer rows.Close()

	return scanItems(rows)
}

func (s *Store) ListAllItems(itemType string) ([]Item, error) {
	query := `SELECT id, namespace, type, name, author, version, created_at, updated_at FROM items`
	var args []any

	if itemType != "" {
		query += ` WHERE type = $1`
		args = append(args, itemType)
	}
	query += ` ORDER BY name`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing all items: %w", err)
	}
	defer rows.Close()

	return scanItems(rows)
}

func (s *Store) DeleteItem(namespace, itemType, name, author string) error {
	var rowAuthor string
	err := s.db.QueryRow(`
		SELECT author FROM items WHERE namespace = $1 AND type = $2 AND name = $3
	`, namespace, itemType, name).Scan(&rowAuthor)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("checking item: %w", err)
	}
	if rowAuthor != author {
		return ErrNotAuthorized
	}

	_, err = s.db.Exec(`
		DELETE FROM items WHERE namespace = $1 AND type = $2 AND name = $3
	`, namespace, itemType, name)
	if err != nil {
		return fmt.Errorf("deleting item: %w", err)
	}
	return nil
}

func (s *Store) SearchItems(query string) ([]Item, error) {
	pattern := "%" + query + "%"
	rows, err := s.db.Query(`
		SELECT id, namespace, type, name, author, version, created_at, updated_at
		FROM items
		WHERE (name ILIKE $1 OR namespace ILIKE $1 OR author ILIKE $1 OR type ILIKE $1)
		  AND namespace NOT LIKE '@%'
		ORDER BY name
	`, pattern)
	if err != nil {
		return nil, fmt.Errorf("searching items: %w", err)
	}
	defer rows.Close()

	return scanItems(rows)
}

func (s *Store) CreateProfile(namespace, name, author string) (*Profile, error) {
	emptyItems, _ := json.Marshal([]ProfileRef{})
	var p Profile
	var itemsJSON []byte
	err := s.db.QueryRow(`
		INSERT INTO profiles (namespace, name, items, author)
		VALUES ($1, $2, $3, $4)
		RETURNING id, namespace, name, items, author, created_at, updated_at
	`, namespace, name, emptyItems, author).Scan(
		&p.ID, &p.Namespace, &p.Name, &itemsJSON,
		&p.Author, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("creating profile: %w", err)
	}
	if err := json.Unmarshal(itemsJSON, &p.Items); err != nil {
		return nil, fmt.Errorf("unmarshaling profile items: %w", err)
	}
	return &p, nil
}

func (s *Store) GetProfile(namespace, name string) (*Profile, error) {
	var p Profile
	var itemsJSON []byte
	err := s.db.QueryRow(`
		SELECT id, namespace, name, items, author, created_at, updated_at
		FROM profiles
		WHERE namespace = $1 AND name = $2
	`, namespace, name).Scan(
		&p.ID, &p.Namespace, &p.Name, &itemsJSON,
		&p.Author, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting profile: %w", err)
	}
	if err := json.Unmarshal(itemsJSON, &p.Items); err != nil {
		return nil, fmt.Errorf("unmarshaling profile items: %w", err)
	}
	return &p, nil
}

func (s *Store) ListProfiles(namespace string) ([]Profile, error) {
	rows, err := s.db.Query(`
		SELECT id, namespace, name, items, author, created_at, updated_at
		FROM profiles
		WHERE namespace = $1
		ORDER BY name
	`, namespace)
	if err != nil {
		return nil, fmt.Errorf("listing profiles: %w", err)
	}
	defer rows.Close()

	var profiles []Profile
	for rows.Next() {
		var p Profile
		var itemsJSON []byte
		if err := rows.Scan(&p.ID, &p.Namespace, &p.Name, &itemsJSON, &p.Author, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning profile: %w", err)
		}
		if err := json.Unmarshal(itemsJSON, &p.Items); err != nil {
			return nil, fmt.Errorf("unmarshaling profile items: %w", err)
		}
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}

func (s *Store) DeleteProfile(namespace, name, author string) error {
	var rowAuthor string
	err := s.db.QueryRow(`
		SELECT author FROM profiles WHERE namespace = $1 AND name = $2
	`, namespace, name).Scan(&rowAuthor)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("checking profile: %w", err)
	}
	if rowAuthor != author {
		return ErrNotAuthorized
	}

	_, err = s.db.Exec(`
		DELETE FROM profiles WHERE namespace = $1 AND name = $2
	`, namespace, name)
	if err != nil {
		return fmt.Errorf("deleting profile: %w", err)
	}
	return nil
}

func (s *Store) AddProfileItem(namespace, profileName string, ref ProfileRef) error {
	p, err := s.GetProfile(namespace, profileName)
	if err != nil {
		return err
	}

	p.Items = append(p.Items, ref)
	itemsJSON, err := json.Marshal(p.Items)
	if err != nil {
		return fmt.Errorf("marshaling profile items: %w", err)
	}

	_, err = s.db.Exec(`
		UPDATE profiles SET items = $1, updated_at = NOW()
		WHERE namespace = $2 AND name = $3
	`, itemsJSON, namespace, profileName)
	if err != nil {
		return fmt.Errorf("updating profile items: %w", err)
	}
	return nil
}

func (s *Store) CreateAPIToken(email, name, tokenHash string) error {
	_, err := s.db.Exec(`
		INSERT INTO api_tokens (email, name, token_hash)
		VALUES ($1, $2, $3)
	`, email, name, tokenHash)
	if err != nil {
		return fmt.Errorf("creating api token: %w", err)
	}
	return nil
}

func (s *Store) ListAPITokens(email string) ([]APITokenRow, error) {
	rows, err := s.db.Query(`
		SELECT name, token_hash, created_at
		FROM api_tokens
		WHERE email = $1
		ORDER BY created_at
	`, email)
	if err != nil {
		return nil, fmt.Errorf("listing api tokens: %w", err)
	}
	defer rows.Close()

	var tokens []APITokenRow
	for rows.Next() {
		var row APITokenRow
		var hash string
		if err := rows.Scan(&row.Name, &hash, &row.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning api token: %w", err)
		}
		if len(hash) >= 4 {
			row.Prefix = hash[len(hash)-4:]
		} else {
			row.Prefix = hash
		}
		tokens = append(tokens, row)
	}
	return tokens, rows.Err()
}

func (s *Store) DeleteAPIToken(email, name string) error {
	result, err := s.db.Exec(`
		DELETE FROM api_tokens WHERE email = $1 AND name = $2
	`, email, name)
	if err != nil {
		return fmt.Errorf("deleting api token: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) GetAPITokenEmail(tokenHash string) (string, error) {
	var email string
	err := s.db.QueryRow(`
		SELECT email FROM api_tokens WHERE token_hash = $1
	`, tokenHash).Scan(&email)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("looking up api token: %w", err)
	}
	return email, nil
}

func (s *Store) RecordEvent(action, namespace, itemType, itemName, email string) error {
	_, err := s.db.Exec(
		`INSERT INTO events (action, namespace, item_type, item_name, email) VALUES ($1, $2, $3, $4, $5)`,
		action, namespace, itemType, itemName, email,
	)
	return err
}

func (s *Store) GetMetrics() (*MetricsSummary, error) {
	var m MetricsSummary

	// Total items
	s.db.QueryRow(`SELECT COUNT(*) FROM items`).Scan(&m.TotalItems)

	// Total pushes
	s.db.QueryRow(`SELECT COUNT(*) FROM events WHERE action = 'push'`).Scan(&m.TotalPushes)

	// Total installs
	s.db.QueryRow(`SELECT COUNT(*) FROM events WHERE action = 'install'`).Scan(&m.TotalInstalls)

	// Active users (last 30 days)
	s.db.QueryRow(`SELECT COUNT(DISTINCT email) FROM events WHERE created_at > NOW() - INTERVAL '30 days'`).Scan(&m.ActiveUsers)

	// Top 10 installed
	rows, err := s.db.Query(`
		SELECT namespace, item_type, item_name, COUNT(*) as cnt
		FROM events WHERE action = 'install'
		GROUP BY namespace, item_type, item_name
		ORDER BY cnt DESC LIMIT 10
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var is ItemStats
			rows.Scan(&is.Namespace, &is.Type, &is.Name, &is.Count)
			m.TopInstalled = append(m.TopInstalled, is)
		}
	}
	if m.TopInstalled == nil {
		m.TopInstalled = []ItemStats{}
	}

	// Top 10 pushed
	rows2, err := s.db.Query(`
		SELECT namespace, item_type, item_name, COUNT(*) as cnt
		FROM events WHERE action = 'push'
		GROUP BY namespace, item_type, item_name
		ORDER BY cnt DESC LIMIT 10
	`)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var is ItemStats
			rows2.Scan(&is.Namespace, &is.Type, &is.Name, &is.Count)
			m.TopPushed = append(m.TopPushed, is)
		}
	}
	if m.TopPushed == nil {
		m.TopPushed = []ItemStats{}
	}

	// Recent 50 events
	rows3, err := s.db.Query(`
		SELECT id, action, namespace, COALESCE(item_type, ''), COALESCE(item_name, ''), email, created_at
		FROM events ORDER BY created_at DESC LIMIT 50
	`)
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var e Event
			rows3.Scan(&e.ID, &e.Action, &e.Namespace, &e.ItemType, &e.ItemName, &e.Email, &e.CreatedAt)
			m.RecentEvents = append(m.RecentEvents, e)
		}
	}
	if m.RecentEvents == nil {
		m.RecentEvents = []Event{}
	}

	return &m, nil
}

func (s *Store) ListItemVersions(namespace, itemType, name string) ([]ItemVersion, error) {
	rows, err := s.db.Query(`
		SELECT version, author, created_at, octet_length(content)
		FROM item_versions
		WHERE namespace = $1 AND type = $2 AND name = $3
		ORDER BY version DESC
	`, namespace, itemType, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var versions []ItemVersion
	for rows.Next() {
		var v ItemVersion
		rows.Scan(&v.Version, &v.Author, &v.CreatedAt, &v.Size)
		versions = append(versions, v)
	}
	if versions == nil {
		versions = []ItemVersion{}
	}
	return versions, rows.Err()
}

func (s *Store) GetItemVersion(namespace, itemType, name string, version int) (*Item, error) {
	var item Item
	err := s.db.QueryRow(`
		SELECT namespace, type, name, version, content, author, created_at
		FROM item_versions
		WHERE namespace = $1 AND type = $2 AND name = $3 AND version = $4
	`, namespace, itemType, name, version).Scan(
		&item.Namespace, &item.Type, &item.Name,
		&item.Version, &item.Content, &item.Author, &item.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	item.UpdatedAt = item.CreatedAt
	return &item, nil
}

func scanItems(rows *sql.Rows) ([]Item, error) {
	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(
			&item.ID, &item.Namespace, &item.Type, &item.Name,
			&item.Author, &item.Version, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
