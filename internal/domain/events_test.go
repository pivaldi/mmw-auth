package domain_test

import (
	"testing"

	"github.com/pivaldi/mmw-auth/internal/domain"
)

func TestEventTypes_ReturnOwnConstants(t *testing.T) {
	cases := []struct {
		event domain.DomainEvent
		want  string
	}{
		{domain.UserRegistered{}, domain.EventTypeUserRegistered},
		{domain.UserDeleted{}, domain.EventTypeUserDeleted},
		{domain.PasswordChanged{}, domain.EventTypePasswordChanged},
		{domain.UserLoggedIn{}, domain.EventTypeUserLoggedIn},
	}

	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			if got := tc.event.EventType(); got != tc.want {
				t.Errorf("%T.EventType() = %q, want %q", tc.event, got, tc.want)
			}
		})
	}
}
