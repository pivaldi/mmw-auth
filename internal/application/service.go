package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	authdomain "github.com/pivaldi/mmw/auth/internal/domain/auth"
	"github.com/pivaldi/mmw/auth/internal/domain/auth/user"
	"github.com/pivaldi/mmw/auth/internal/application/ports"
)

// Sentinel errors for use-case failures.
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrUserNotFound       = errors.New("user not found")
)

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
	l, err := user.NewLogin(login)
	if err != nil {
		return uuid.Nil, err
	}

	id := uuid.New()
	var userID uuid.UUID

	err = s.uow.WithTransaction(ctx, func(ctx context.Context) error {
		u, err := user.Create(id, l, password)
		if err != nil {
			return err
		}
		if err := s.userRepo.Save(ctx, u); err != nil {
			return err
		}
		if err := s.dispatcher.Dispatch(ctx, u.ClearEvents()); err != nil {
			return err
		}
		userID = u.ID()
		return nil
	})
	return userID, err
}

// Login authenticates a user and returns a JWT token and the user ID.
func (s *AuthApplicationService) Login(ctx context.Context, login, password string) (string, uuid.UUID, error) {
	l, err := user.NewLogin(login)
	if err != nil {
		return "", uuid.Nil, ErrInvalidCredentials
	}

	var token string
	var userID uuid.UUID

	err = s.uow.WithTransaction(ctx, func(ctx context.Context) error {
		u, err := s.userRepo.FindByLogin(ctx, l)
		if err != nil {
			return ErrInvalidCredentials
		}
		if !u.CheckPassword(password) {
			return ErrInvalidCredentials
		}

		t, err := s.createJWT(u.ID())
		if err != nil {
			return err
		}

		u.MarkLoggedIn()
		if err := s.userRepo.Update(ctx, u); err != nil {
			return err
		}
		if err := s.dispatcher.Dispatch(ctx, u.ClearEvents()); err != nil {
			return err
		}

		sess := authdomain.NewSession(u.ID(), t, 24*time.Hour)
		if err := s.sessionRepo.Save(ctx, sess); err != nil {
			return err
		}

		token = t
		userID = u.ID()
		return nil
	})

	return token, userID, err
}

// ValidateToken verifies JWT signature and confirms the session exists in the DB.
func (s *AuthApplicationService) ValidateToken(ctx context.Context, tokenString string) (uuid.UUID, error) {
	parsed, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})
	if err != nil || !parsed.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, ErrInvalidToken
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	sess, err := s.sessionRepo.FindByToken(ctx, tokenString)
	if err != nil || sess == nil {
		return uuid.Nil, ErrInvalidToken
	}

	if sess.UserID() != userID {
		return uuid.Nil, ErrInvalidToken
	}

	return userID, nil
}

// ChangePassword replaces the user's password after verifying the old one.
func (s *AuthApplicationService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	return s.uow.WithTransaction(ctx, func(ctx context.Context) error {
		u, err := s.userRepo.FindByID(ctx, userID)
		if err != nil {
			return ErrUserNotFound
		}
		if err := u.ChangePassword(oldPassword, newPassword); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidCredentials, err)
		}
		if err := s.userRepo.Update(ctx, u); err != nil {
			return err
		}
		return s.dispatcher.Dispatch(ctx, u.ClearEvents())
	})
}

// DeleteUser removes a user from the system.
func (s *AuthApplicationService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return s.uow.WithTransaction(ctx, func(ctx context.Context) error {
		u, err := s.userRepo.FindByID(ctx, userID)
		if err != nil {
			return ErrUserNotFound
		}
		u.Delete()
		if err := s.dispatcher.Dispatch(ctx, u.ClearEvents()); err != nil {
			return err
		}
		return s.userRepo.Delete(ctx, userID)
	})
}

func (s *AuthApplicationService) createJWT(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"authorized": true,
		"user_id":    userID.String(),
		"exp":        time.Now().Add(24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}
