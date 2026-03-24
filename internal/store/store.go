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
		WHERE name ILIKE $1 AND namespace NOT LIKE '@%'
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
