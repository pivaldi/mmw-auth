package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserSnapshot is the Memento for the User aggregate.
// It captures the full persistent state as primitives so that
// pgx.RowToStructByName can scan DB columns directly via db tags.
// The events field is intentionally excluded — it is runtime state only.
type UserSnapshot struct {
	ID           uuid.UUID `db:"id"`
	Login        string    `db:"login"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
