package user

import (
	"encoding/json"
	"time"

	"github.com/rotisserie/eris"
)

const (
	EventUserRegistered = "auth.user.registered"
	EventUserDeleted    = "auth.user.deleted"
	//nolint:gosec // Event type constant, not a credential
	EventPasswordChanged = "auth.user.password_changed"
	EventUserLoggedIn    = "auth.user.logged_in"
)

var AllEvents = []string{
	EventUserRegistered,
	EventUserDeleted,
	EventPasswordChanged,
	EventUserLoggedIn,
}

// DomainEvent is the interface all auth domain events implement.
type DomainEvent interface {
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
}

// UserRegistered is emitted when a new user is created.
type UserRegistered struct {
	aggregateID string
	occurredAt  time.Time
	Login       string
}

func (e UserRegistered) EventType() string     { return EventUserRegistered }
func (e UserRegistered) AggregateID() string   { return e.aggregateID }
func (e UserRegistered) OccurredAt() time.Time { return e.occurredAt }

func (e UserRegistered) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		AggregateID string    `json:"aggregateId"`
		OccurredAt  time.Time `json:"occurredAt"`
		Login       string    `json:"login"`
	}{e.aggregateID, e.occurredAt, e.Login})

	return data, eris.Wrap(err, "marshaling UserRegistered event")
}

// UserDeleted is emitted when a user is deleted.
type UserDeleted struct {
	aggregateID string
	occurredAt  time.Time
}

func (e UserDeleted) EventType() string     { return EventUserDeleted }
func (e UserDeleted) AggregateID() string   { return e.aggregateID }
func (e UserDeleted) OccurredAt() time.Time { return e.occurredAt }

func (e UserDeleted) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		AggregateID string    `json:"aggregateId"`
		OccurredAt  time.Time `json:"occurredAt"`
	}{e.aggregateID, e.occurredAt})

	return data, eris.Wrap(err, "marshaling UserDeleted event")
}

// PasswordChanged is emitted when a user changes their password.
type PasswordChanged struct {
	aggregateID string
	occurredAt  time.Time
}

func (e PasswordChanged) EventType() string     { return EventPasswordChanged }
func (e PasswordChanged) AggregateID() string   { return e.aggregateID }
func (e PasswordChanged) OccurredAt() time.Time { return e.occurredAt }

func (e PasswordChanged) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		AggregateID string    `json:"aggregateId"`
		OccurredAt  time.Time `json:"occurredAt"`
	}{e.aggregateID, e.occurredAt})

	return data, eris.Wrap(err, "marshaling PasswordChanged event")
}

// UserLoggedIn is emitted when a user successfully logs in.
type UserLoggedIn struct {
	aggregateID string
	occurredAt  time.Time
}

func (e UserLoggedIn) EventType() string     { return EventUserLoggedIn }
func (e UserLoggedIn) AggregateID() string   { return e.aggregateID }
func (e UserLoggedIn) OccurredAt() time.Time { return e.occurredAt }

func (e UserLoggedIn) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		AggregateID string    `json:"aggregateId"`
		OccurredAt  time.Time `json:"occurredAt"`
	}{e.aggregateID, e.occurredAt})

	return data, eris.Wrap(err, "marshaling UserLoggedIn event")
}
