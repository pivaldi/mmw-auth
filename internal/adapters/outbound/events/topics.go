package events

import (
	"github.com/pivaldi/mmw-auth/internal/domain"
	authdef "github.com/pivaldi/mmw-contracts/definitions/auth"
)

// domainTopics maps semantic domain event types to Watermill routing keys.
// This is the single place where domain semantics are translated to transport concerns.
//
//nolint:gochecknoglobals // package-level lookup table, not mutable state
var domainTopics = map[string]string{
	domain.EventTypeUserRegistered:  authdef.TopicUserRegistered,
	domain.EventTypeUserDeleted:     authdef.TopicUserDeleted,
	domain.EventTypePasswordChanged: authdef.TopicPasswordChanged,
	domain.EventTypeUserLoggedIn:    authdef.TopicUserLoggedIn,
}
