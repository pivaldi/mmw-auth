package events

import (
	"testing"

	"github.com/pivaldi/mmw-auth/internal/domain"
	authdef "github.com/pivaldi/mmw-contracts/go/application/auth"
)

func TestDomainTopics_AllEventTypesCovered(t *testing.T) {
	expected := map[string]string{
		domain.EventTypeUserRegistered:  authdef.TopicUserRegistered,
		domain.EventTypeUserDeleted:     authdef.TopicUserDeleted,
		domain.EventTypePasswordChanged: authdef.TopicPasswordChanged,
		domain.EventTypeUserLoggedIn:    authdef.TopicUserLoggedIn,
	}

	for domainType, wantTopic := range expected {
		t.Run(domainType, func(t *testing.T) {
			got, ok := domainTopics[domainType]
			if !ok {
				t.Fatalf("domainTopics missing entry for %q", domainType)
			}
			if got != wantTopic {
				t.Errorf("domainTopics[%q] = %q, want %q", domainType, got, wantTopic)
			}
		})
	}
}

func TestDomainTopics_TopicIsNotSameAsEventType(t *testing.T) {
	for domainType, topic := range domainTopics {
		if topic == domainType {
			t.Errorf("domainTopics[%q]: topic should be a routing key, not the domain event type", domainType)
		}
	}
}
