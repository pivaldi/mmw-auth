// modules/auth/internal/adapters/inbound/connect/errors_test.go
package connect

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/piprim/mmw/pkg/platform"
	defauth "github.com/pivaldi/mmw-contracts/definitions/auth"
)

func TestConnectErrorFrom_DomainError_MapsToCorrectCode(t *testing.T) {
	cases := []struct {
		name     string
		code     platform.ErrorCode
		wantCode connect.Code
	}{
		{"InvalidLogin", platform.ErrorCode(defauth.ErrorCodeInvalidLogin), connect.CodeInvalidArgument},
		{"InvalidPassword", platform.ErrorCode(defauth.ErrorCodeInvalidPassword), connect.CodeInvalidArgument},
		{"InvalidCredentials", platform.ErrorCode(defauth.ErrorCodeInvalidCredentials), connect.CodeUnauthenticated},
		{"InvalidToken", platform.ErrorCode(defauth.ErrorCodeInvalidToken), connect.CodeUnauthenticated},
		{"UserNotFound", platform.ErrorCode(defauth.ErrorCodeUserNotFound), connect.CodeNotFound},
		{"UserAlreadyExists", platform.ErrorCode(defauth.ErrorCodeUserAlreadyExists), connect.CodeAlreadyExists},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			domainErr := &platform.DomainError{Code: tc.code, Message: "test error"}

			result := connectErrorFrom(domainErr)

			var connectErr *connect.Error
			if !errors.As(result, &connectErr) {
				t.Fatalf("expected *connect.Error, got %T", result)
			}

			if connectErr.Code() != tc.wantCode {
				t.Errorf("Code() = %v, want %v", connectErr.Code(), tc.wantCode)
			}
		})
	}
}

func TestConnectErrorFrom_DomainError_HasDetail(t *testing.T) {
	domainErr := &platform.DomainError{
		Code:    platform.ErrorCode(defauth.ErrorCodeInvalidCredentials),
		Message: "invalid credentials",
	}

	result := connectErrorFrom(domainErr)

	var connectErr *connect.Error
	if !errors.As(result, &connectErr) {
		t.Fatalf("expected *connect.Error, got %T", result)
	}

	if len(connectErr.Details()) == 0 {
		t.Error("expected at least one error detail")
	}
}

func TestConnectErrorFrom_UnknownDomainCode_IsInternal(t *testing.T) {
	domainErr := &platform.DomainError{Code: 9999, Message: "unknown"}

	result := connectErrorFrom(domainErr)

	var connectErr *connect.Error
	if !errors.As(result, &connectErr) {
		t.Fatalf("expected *connect.Error, got %T", result)
	}

	if connectErr.Code() != connect.CodeInternal {
		t.Errorf("Code() = %v, want %v", connectErr.Code(), connect.CodeInternal)
	}
}

func TestConnectErrorFrom_NonDomainError_IsInternal(t *testing.T) {
	infraErr := errors.New("db connection refused")

	result := connectErrorFrom(infraErr)

	var connectErr *connect.Error
	if !errors.As(result, &connectErr) {
		t.Fatalf("expected *connect.Error, got %T", result)
	}

	if connectErr.Code() != connect.CodeInternal {
		t.Errorf("Code() = %v, want %v", connectErr.Code(), connect.CodeInternal)
	}
}
