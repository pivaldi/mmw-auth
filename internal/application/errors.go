// modules/auth/internal/application/errors.go
package application

import (
	"errors"

	"github.com/piprim/mmw/platform"
	"github.com/pivaldi/mmw-auth/internal/domain/user"
	defauth "github.com/pivaldi/mmw-contracts/definitions/auth"
)

// DomainErrorFor translates a domain sentinel error into a *platform.DomainError
// so the inbound adapter can map it to a typed Connect error detail.
// Non-domain errors (infra, unexpected) are returned unchanged.
func DomainErrorFor(err error) error {
	switch {
	case errors.Is(err, user.ErrInvalidLogin):
		return &platform.DomainError{
			Code:    platform.ErrorCode(defauth.ErrorCodeInvalidLogin),
			Message: err.Error(),
		}

	case errors.Is(err, user.ErrInvalidPassword):
		return &platform.DomainError{
			Code:    platform.ErrorCode(defauth.ErrorCodeInvalidPassword),
			Message: err.Error(),
		}

	case errors.Is(err, user.ErrInvalidCredentials):
		return &platform.DomainError{
			Code:    platform.ErrorCode(defauth.ErrorCodeInvalidCredentials),
			Message: err.Error(),
		}

	case errors.Is(err, user.ErrInvalidToken):
		return &platform.DomainError{
			Code:    platform.ErrorCode(defauth.ErrorCodeInvalidToken),
			Message: err.Error(),
		}

	case errors.Is(err, user.ErrUserNotFound):
		return &platform.DomainError{
			Code:    platform.ErrorCode(defauth.ErrorCodeUserNotFound),
			Message: err.Error(),
		}

	case errors.Is(err, user.ErrUserAlreadyExists):
		return &platform.DomainError{
			Code:    platform.ErrorCode(defauth.ErrorCodeUserAlreadyExists),
			Message: err.Error(),
		}
	}

	return err
}
