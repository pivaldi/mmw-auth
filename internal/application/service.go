package application

import (
	"context"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/pivaldi/mmw-auth/internal/application/ports"
	"github.com/pivaldi/mmw-auth/internal/domain"
	defauth "github.com/pivaldi/mmw-contracts/definitions/auth"
	"github.com/rotisserie/eris"
)

const tokenDuration = 24 * time.Hour

// AuthApplicationService orchestrates all auth use cases.
type AuthApplicationService struct {
	userRepo    ports.UserRepository
	sessionRepo ports.SessionRepository
	uow         ports.UnitOfWork
	dispatcher  ports.EventDispatcher
	jwtSecret   []byte
}

// NewAuthService creates an AuthApplicationService with all required dependencies.
func NewAuthService(
	userRepo ports.UserRepository,
	sessionRepo ports.SessionRepository,
	uow ports.UnitOfWork,
	dispatcher ports.EventDispatcher,
	jwtSecret string,
) *AuthApplicationService {
	return &AuthApplicationService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		uow:         uow,
		dispatcher:  dispatcher,
		jwtSecret:   []byte(jwtSecret),
	}
}

// Register creates a new user account.
func (s *AuthApplicationService) Register(ctx context.Context, login, password string) (uuid.UUID, error) {
	l, err := domain.NewLogin(login)
	if err != nil {
		return uuid.Nil, DomainErrorFor(err)
	}

	id := uuid.New()
	var userID uuid.UUID

	err = s.uow.WithTransaction(ctx, func(ctx context.Context) error {
		u, err := domain.Create(id, l, password)
		if err != nil {
			return DomainErrorFor(err)
		}
		if err := s.userRepo.Save(ctx, u); err != nil {
			return eris.Wrap(err, "saving user")
		}
		if err := s.dispatcher.Dispatch(ctx, u.ClearEvents()); err != nil {
			return eris.Wrap(err, "dispatching events")
		}
		userID = u.ID()

		return nil
	})
	if err != nil {
		return uuid.Nil, DomainErrorFor(err)
	}

	return userID, nil
}

// Login authenticates a user and returns a JWT token and the user ID.
func (s *AuthApplicationService) Login(ctx context.Context, login, password string) (string, uuid.UUID, error) {
	l, err := domain.NewLogin(login)
	if err != nil {
		return "", uuid.Nil, DomainErrorFor(domain.ErrInvalidCredentials)
	}

	var token string
	var userID uuid.UUID

	err = s.uow.WithTransaction(ctx, func(ctx context.Context) error {
		u, err := s.userRepo.FindByLogin(ctx, l)
		if err != nil {
			return domain.ErrInvalidCredentials
		}
		if !u.CheckPassword(password) {
			return domain.ErrInvalidCredentials
		}

		t, err := s.createJWT(u.ID())
		if err != nil {
			return err
		}

		u.MarkLoggedIn()
		if err := s.userRepo.Update(ctx, u); err != nil {
			return eris.Wrap(err, "updating user")
		}
		if err := s.dispatcher.Dispatch(ctx, u.ClearEvents()); err != nil {
			return eris.Wrap(err, "dispatching events")
		}

		sess := domain.NewSession(u.ID(), t, tokenDuration)
		if err := s.sessionRepo.Save(ctx, sess); err != nil {
			return eris.Wrap(err, "saving session")
		}

		token = t
		userID = u.ID()

		return nil
	})
	if err != nil {
		return "", uuid.Nil, DomainErrorFor(err)
	}

	return token, userID, nil
}

// ValidateToken verifies JWT signature and confirms the session exists in the DB.
func (s *AuthApplicationService) ValidateToken(ctx context.Context, tokenString string) (uuid.UUID, error) {
	parsed, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, DomainErrorFor(domain.ErrInvalidToken)
		}

		return s.jwtSecret, nil
	})
	if err != nil || !parsed.Valid {
		return uuid.Nil, DomainErrorFor(domain.ErrInvalidToken)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, DomainErrorFor(domain.ErrInvalidToken)
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, DomainErrorFor(domain.ErrInvalidToken)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, DomainErrorFor(domain.ErrInvalidToken)
	}

	sess, err := s.sessionRepo.FindByToken(ctx, tokenString)
	if err != nil || sess == nil {
		return uuid.Nil, DomainErrorFor(domain.ErrInvalidToken)
	}

	if sess.UserID() != userID {
		return uuid.Nil, DomainErrorFor(domain.ErrInvalidToken)
	}

	return userID, nil
}

// GetUser retrieves a user by UUID for cross-service in-process access.
func (s *AuthApplicationService) GetUser(ctx context.Context, id string) (*defauth.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, DomainErrorFor(domain.ErrUserNotFound)
	}

	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, DomainErrorFor(domain.ErrUserNotFound)
	}

	return &defauth.User{Id: u.ID().String(), Login: u.Login().String()}, nil
}

// ChangePassword replaces the user's password after verifying the old one.
func (s *AuthApplicationService) ChangePassword(
	ctx context.Context,
	userID uuid.UUID,
	oldPassword, newPassword string,
) error {
	err := s.uow.WithTransaction(ctx, func(ctx context.Context) error {
		u, err := s.userRepo.FindByID(ctx, userID)
		if err != nil {
			return DomainErrorFor(domain.ErrUserNotFound)
		}
		if err := u.ChangePassword(oldPassword, newPassword); err != nil {
			return DomainErrorFor(err)
		}
		if err := s.userRepo.Update(ctx, u); err != nil {
			return eris.Wrap(err, "updating user password")
		}

		if err := s.dispatcher.Dispatch(ctx, u.ClearEvents()); err != nil {
			return eris.Wrap(err, "dispatching events")
		}

		return nil
	})
	if err != nil {
		return DomainErrorFor(err)
	}

	return nil
}

// DeleteUser removes a user from the system.
func (s *AuthApplicationService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	err := s.uow.WithTransaction(ctx, func(ctx context.Context) error {
		u, err := s.userRepo.FindByID(ctx, userID)
		if err != nil {
			return DomainErrorFor(domain.ErrUserNotFound)
		}
		u.Delete()
		if err := s.dispatcher.Dispatch(ctx, u.ClearEvents()); err != nil {
			return eris.Wrap(err, "dispatching events")
		}

		if err := s.userRepo.Delete(ctx, userID); err != nil {
			return eris.Wrap(err, "deleting user")
		}

		return nil
	})
	if err != nil {
		return DomainErrorFor(err)
	}

	return nil
}

func (s *AuthApplicationService) createJWT(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"authorized": true,
		"user_id":    userID.String(),
		"exp":        time.Now().Add(tokenDuration).Unix(),
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
	if err != nil {
		return "", eris.Wrap(err, "signing JWT")
	}

	return token, nil
}

// Health return a simple database health check
func (s *AuthApplicationService) Health(ctx context.Context) (any, error) {
	count, err := s.userRepo.Health(ctx)
	if err != nil {
		return 0, DomainErrorFor(err)
	}

	return count, nil
}
