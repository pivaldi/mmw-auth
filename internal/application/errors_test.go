// modules/auth/internal/application/errors_test.go
package application_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ovya/ogl/platform"
	"github.com/pivaldi/mmw-auth/internal/application"
	"github.com/pivaldi/mmw-auth/internal/domain/user"
	defauth "github.com/pivaldi/mmw-contracts/definitions/auth"
)

func TestDomainErrorFor_KnownSentinels(t *testing.T) {
	cases := []struct {
		name     string
		input    error
		wantCode platform.ErrorCode
	}{
		{
			"ErrInvalidLogin", user.ErrInvalidLogin,
			platform.ErrorCode(defauth.ErrorCodeInvalidLogin),
		},
		{
			"ErrInvalidPassword", user.ErrInvalidPassword,
			platform.ErrorCode(defauth.ErrorCodeInvalidPassword),
		},
		{
			"ErrInvalidCredentials", user.ErrInvalidCredentials,
			platform.ErrorCode(defauth.ErrorCodeInvalidCredentials),
		},
		{
			"ErrInvalidToken", user.ErrInvalidToken,
			platform.ErrorCode(defauth.ErrorCodeInvalidToken),
		},
		{
			"ErrUserNotFound", user.ErrUserNotFound,
			platform.ErrorCode(defauth.ErrorCodeUserNotFound),
		},
		{
			"ErrUserAlreadyExists", user.ErrUserAlreadyExists,
			platform.ErrorCode(defauth.ErrorCodeUserAlreadyExists),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := application.DomainErrorFor(tc.input)

			domErr, ok := errors.AsType[*platform.DomainError](result)
			if !ok {
				t.Fatalf("expected *platform.DomainError, got %T", result)
			}

			if domErr.Code != tc.wantCode {
				t.Errorf("Code = %v, want %v", domErr.Code, tc.wantCode)
			}

			if domErr.Message == "" {
				t.Error("Message must not be empty")
			}
		})
	}
}

func TestDomainErrorFor_WrappedSentinel(t *testing.T) {
	wrapped := fmt.Errorf("context: %w", user.ErrInvalidCredentials)

	result := application.DomainErrorFor(wrapped)

	_, ok := errors.AsType[*platform.DomainError](result)
	if !ok {
		t.Fatal("expected *platform.DomainError for wrapped domain error")
	}
}

func TestDomainErrorFor_NonDomainError_PassesThrough(t *testing.T) {
	infra := errors.New("db connection refused")

	result := application.DomainErrorFor(infra)

	if result != infra {
		t.Errorf("expected original error to pass through, got %v", result)
	}
}
