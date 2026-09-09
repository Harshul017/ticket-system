package models

// Status values allowed for a ticket. Kept as constants so the rest of the
// codebase never deals with raw string typos.
const (
	StatusOpen       = "open"
	StatusInProgress = "in_progress"
	StatusClosed     = "closed"
)

// IsValidStatus checks whether a string is one of the three allowed
// ticket statuses.
func IsValidStatus(s string) bool {
	switch s {
	case StatusOpen, StatusInProgress, StatusClosed:
		return true
	default:
		return false
	}
}

// User represents a registered account.
// PasswordHash is never sent back in JSON responses (json:"-").
type User struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
}

// Ticket represents a single support ticket owned by a user.
type Ticket struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
