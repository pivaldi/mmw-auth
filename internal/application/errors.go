// modules/auth/internal/application/errors.go
package application

import (
	"errors"

	"github.com/piprim/mmw/pkg/platform"
	"github.com/pivaldi/mmw-auth/internal/domain"
	defauth "github.com/pivaldi/mmw-contracts/definitions/auth"
)

// DomainErrorFor translates a domain sentinel error into a *platform.DomainError
// so the inbound adapter can map it to a typed Connect error detail.
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
