package domain

import (
	"encoding/json"
	"time"

	authdef "github.com/pivaldi/mmw-contracts/definitions/auth"

	"github.com/rotisserie/eris"
)

// DomainEvent is the interface all auth domain events implement.
type DomainEvent interface {
	// EventType returns the string identifier for this event kind.
	EventType() string
	// AggregateID returns the ID of the aggregate that emitted this event.
	AggregateID() string
	// OccurredAt returns when the event was emitted.
	OccurredAt() time.Time
}

// UserRegistered is emitted when a new user is created.
type UserRegistered struct {
	aggregateID string
	occurredAt  time.Time
	Login       string
}

func (UserRegistered) EventType() string       { return authdef.TopicUserRegistered }
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

func (UserDeleted) EventType() string       { return authdef.TopicUserDeleted }
func (e UserDeleted) AggregateID() string   { return e.aggregateID }
func (e UserDeleted) OccurredAt() time.Time { return e.occurredAt }

// MarshalJSON serialises UserDeleted for internal logging and non-proto consumers.
// WARNING: Do NOT use this for message-bus payloads — the field names (aggregateId/
// occurredAt) do not match the proto contract (userId/deletedAt). Use
// outbox_dispatcher.marshalEvent which routes UserDeleted through protojson.
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

func (PasswordChanged) EventType() string       { return authdef.TopicPasswordChanged }
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

func (UserLoggedIn) EventType() string       { return authdef.TopicUserLoggedIn }
func (e UserLoggedIn) AggregateID() string   { return e.aggregateID }
func (e UserLoggedIn) OccurredAt() time.Time { return e.occurredAt }

func (e UserLoggedIn) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		AggregateID string    `json:"aggregateId"`
		OccurredAt  time.Time `json:"occurredAt"`
	}{e.aggregateID, e.occurredAt})

	return data, eris.Wrap(err, "marshaling UserLoggedIn event")
}
