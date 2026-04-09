// modules/auth/internal/adapters/inbound/mapper/errors.go
package mapper

import (
	"errors"

	"github.com/piprim/mmw/pkg/platform"
	"github.com/pivaldi/mmw-auth/internal/domain"
	defauth "github.com/pivaldi/mmw-contracts/go/application/auth"
)

// DomainErrorFor translates a domain sentinel error into a *platform.DomainError
// using the error codes from contracts (definitions/auth). This is the application
// layer's responsibility: binding domain errors to the shared wire protocol so that
// callers — including other modules communicating in-process — receive a typed error
// that carries no domain-specific knowledge.
// Non-domain errors (infra, unexpected) are returned unchanged.
func DomainErrorFor(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidLogin):
		return &platform.DomainError{
			Code:    platform.ErrorCode(defauth.ErrorCodeInvalidLogin),
			Message: err.Error(),
		}

	case errors.Is(err, domain.ErrInvalidPassword):
		return &platform.DomainError{
			Code:    platform.ErrorCode(defauth.ErrorCodeInvalidPassword),
			Message: err.Error(),
		}

	case errors.Is(err, domain.ErrInvalidCredentials):
		return &platform.DomainError{
			Code:    platform.ErrorCode(defauth.ErrorCodeInvalidCredentials),
			Message: err.Error(),
		}

	case errors.Is(err, domain.ErrInvalidToken):
		return &platform.DomainError{
			Code:    platform.ErrorCode(defauth.ErrorCodeInvalidToken),
			Message: err.Error(),
		}

	case errors.Is(err, domain.ErrUserNotFound):
		return &platform.DomainError{
			Code:    platform.ErrorCode(defauth.ErrorCodeUserNotFound),
			Message: err.Error(),
		}

	case errors.Is(err, domain.ErrUserAlreadyExists):
		return &platform.DomainError{
			Code:    platform.ErrorCode(defauth.ErrorCodeUserAlreadyExists),
			Message: err.Error(),
		}
	}

	return err
}
