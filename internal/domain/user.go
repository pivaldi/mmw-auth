package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// User is the auth bounded context aggregate root.
type User struct {
	id           uuid.UUID
	login        Login
	passwordHash PasswordHash
	createdAt    time.Time
	updatedAt    time.Time
	events       []DomainEvent
}

// Create validates inputs, hashes the password, and emits UserRegistered.
func Create(id uuid.UUID, login Login, plainPassword string) (*User, error) {
	if plainPassword == "" {
		return nil, ErrInvalidPassword
	}
	ph, err := NewPasswordHash(plainPassword)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	u := &User{
		id:           id,
		login:        login,
		passwordHash: ph,
		createdAt:    now,
		updatedAt:    now,
	}
	u.addEvent(UserRegistered{
		aggregateID: id.String(),
		occurredAt:  now,
		Login:       login.String(),
	})

	return u, nil
}

// Snapshot returns the Memento for this User — a plain-data representation
// of the aggregate's current state, suitable for persistence.
// events is not included; it is runtime state only.
func (u *User) Snapshot() UserSnapshot {
	return UserSnapshot{
		ID:           u.id,
		Login:        u.login.String(),
		PasswordHash: u.passwordHash.String(),
		CreatedAt:    u.createdAt,
		UpdatedAt:    u.updatedAt,
	}
}

// ReconstructUser restores a User from persisted state without re-hashing.
// Only repositories should call this.
// Panics if the snapshot contains values that violate basic type invariants —
// this should never happen since the DB is the authoritative source of truth.
func ReconstructUser(snap *UserSnapshot) *User {
	login, err := NewLogin(snap.Login)
	if err != nil {
		panic(fmt.Sprintf("ReconstructUser: invalid login from DB: %v", err))
	}

	return &User{
		id:           snap.ID,
		login:        login,
		passwordHash: NewHashedPassword(snap.PasswordHash),
		createdAt:    snap.CreatedAt,
		updatedAt:    snap.UpdatedAt,
		events:       []DomainEvent{},
	}
}

func (u *User) ID() uuid.UUID              { return u.id }
func (u *User) Login() Login               { return u.login }
func (u *User) PasswordHash() PasswordHash { return u.passwordHash }
func (u *User) CreatedAt() time.Time       { return u.createdAt }
func (u *User) UpdatedAt() time.Time       { return u.updatedAt }

// CheckPassword returns true if plaintext matches the stored hash.
func (u *User) CheckPassword(plaintext string) bool {
	return u.passwordHash.Verify(plaintext)
}

// ChangePassword verifies the old password then replaces it with a new hash.
// Emits PasswordChanged on success.
func (u *User) ChangePassword(oldPassword, newPassword string) error {
	if !u.passwordHash.Verify(oldPassword) {
		return ErrInvalidCredentials
	}
	if newPassword == "" {
		return ErrInvalidPassword
	}
	ph, err := NewPasswordHash(newPassword)
	if err != nil {
		return err
	}
	u.passwordHash = ph
	u.updatedAt = time.Now()
	u.addEvent(PasswordChanged{
		aggregateID: u.id.String(),
		occurredAt:  u.updatedAt,
	})

	return nil
}

// MarkLoggedIn emits UserLoggedIn without changing aggregate state.
func (u *User) MarkLoggedIn() {
	u.addEvent(UserLoggedIn{
		aggregateID: u.id.String(),
		occurredAt:  time.Now(),
	})
}

// Delete emits UserDeleted without changing aggregate state.
func (u *User) Delete() {
	u.addEvent(UserDeleted{
		aggregateID: u.id.String(),
		occurredAt:  time.Now(),
	})
}

// ClearEvents returns all pending events and clears the internal slice.
// Called by the outbox dispatcher after writing events to the DB.
func (u *User) ClearEvents() []DomainEvent {
	evts := u.events
	u.events = nil

	return evts
}

// addEvent adds a domain event to the unpublished events list
func (u *User) addEvent(event DomainEvent) {
	u.events = append(u.events, event)
}
