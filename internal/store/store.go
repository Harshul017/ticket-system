package store

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"ticket-system/internal/models"

	_ "modernc.org/sqlite" // pure-Go SQLite driver, registers itself as "sqlite"
)

// ErrNotFound is returned when a lookup finds nothing.
// Handlers translate this into a 404.
var ErrNotFound = errors.New("not found")

// ErrDuplicateEmail is returned when registering with an email that
// already exists.
var ErrDuplicateEmail = errors.New("email already registered")

// Store wraps a *sql.DB and exposes the operations our handlers need.
// Keeping this as a small struct (rather than scattering raw SQL through
// handlers) keeps ownership checks and queries in one place.
type Store struct {
	db *sql.DB
}

// New opens (or creates) the SQLite file at path and ensures the schema exists.
func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// SQLite only supports one writer at a time; this keeps the driver
	// from opening multiple connections that could step on each other.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

// migrate creates the tables if they don't already exist.
func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS tickets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		status TEXT NOT NULL DEFAULT 'open',
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	`
	_, err := s.db.Exec(schema)
	return err
}

// ---------- Users ----------

// CreateUser inserts a new user with an already-hashed password.
func (s *Store) CreateUser(email, passwordHash string) (*models.User, error) {
	res, err := s.db.Exec(
		`INSERT INTO users (email, password_hash) VALUES (?, ?)`,
		email, passwordHash,
	)
	if err != nil {
		// SQLite raises a "UNIQUE constraint failed" error for duplicate emails.
		if strings.Contains(strings.ToLower(err.Error()), "unique constraint") {
			return nil, ErrDuplicateEmail
		}
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &models.User{ID: id, Email: email, PasswordHash: passwordHash}, nil
}

// GetUserByEmail looks up a user for login.
func (s *Store) GetUserByEmail(email string) (*models.User, error) {
	row := s.db.QueryRow(
		`SELECT id, email, password_hash FROM users WHERE email = ?`, email,
	)
	var u models.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

// ---------- Tickets ----------

// CreateTicket inserts a new ticket owned by userID, defaulting to "open".
func (s *Store) CreateTicket(userID int64, title, description string) (*models.Ticket, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec(
		`INSERT INTO tickets (user_id, title, description, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		userID, title, description, models.StatusOpen, now, now,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &models.Ticket{
		ID: id, UserID: userID, Title: title, Description: description,
		Status: models.StatusOpen, CreatedAt: now, UpdatedAt: now,
	}, nil
}

// ListTicketsByUser returns every ticket owned by userID.
// This is what powers GET /tickets — it never sees other users' rows.
func (s *Store) ListTicketsByUser(userID int64) ([]models.Ticket, error) {
	rows, err := s.db.Query(
		`SELECT id, user_id, title, description, status, created_at, updated_at
		 FROM tickets WHERE user_id = ? ORDER BY id`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tickets := []models.Ticket{}
	for rows.Next() {
		var t models.Ticket
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}
	return tickets, rows.Err()
}

// GetTicketByID fetches a single ticket regardless of owner.
// The handler is responsible for checking ownership after calling this —
// that's what makes the 404-for-non-owner behavior consistent in one place.
func (s *Store) GetTicketByID(id int64) (*models.Ticket, error) {
	row := s.db.QueryRow(
		`SELECT id, user_id, title, description, status, created_at, updated_at
		 FROM tickets WHERE id = ?`, id,
	)
	var t models.Ticket
	if err := row.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

// UpdateTicketStatus overwrites a ticket's status and updated_at timestamp.
// Ownership and status-transition validity must be checked by the caller
// BEFORE calling this — this method just performs the write.
func (s *Store) UpdateTicketStatus(id int64, newStatus string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`UPDATE tickets SET status = ?, updated_at = ? WHERE id = ?`,
		newStatus, now, id,
	)
	return err
}
