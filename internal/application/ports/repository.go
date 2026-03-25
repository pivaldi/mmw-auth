package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/pivaldi/mmw-auth/internal/domain"
	"github.com/pivaldi/mmw-auth/internal/domain/user"
)

// UserRepository defines persistence operations for the User aggregate.
type UserRepository interface {
	// Save persists a new user.
	Save(ctx context.Context, u *user.User) error
	// FindByLogin retrieves a user by login.
	FindByLogin(ctx context.Context, login user.Login) (*user.User, error)
	// FindByID retrieves a user by its UUID.
	FindByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	// Update persists changes to an existing user.
	Update(ctx context.Context, u *user.User) error
	// Delete removes a user by ID.
	Delete(ctx context.Context, id uuid.UUID) error
	// Health returns a liveness indicator for the underlying store.
	Health(ctx context.Context) (any, error)
}

// SessionRepository defines persistence operations for sessions.
type SessionRepository interface {
	// Save persists a new session.
	Save(ctx context.Context, s *domain.Session) error
	// FindByToken retrieves a session by its token.
	FindByToken(ctx context.Context, token string) (*domain.Session, error)
}

// UnitOfWork manages transaction boundaries.
type UnitOfWork interface {
	// WithTransaction executes fn inside a database transaction.
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// EventDispatcher writes domain events to the outbox table.
// It is called inside a transaction so events are written atomically
// with the domain record that emitted them.
type EventDispatcher interface {
	// Dispatch persists domain events to the outbox.
	Dispatch(ctx context.Context, events []user.DomainEvent) error
}
