package inproc

import (
	"context"

	"github.com/google/uuid"
	"github.com/pivaldi/mmw-auth/internal/adapters/inbound/mapper"
	"github.com/pivaldi/mmw-auth/internal/application"
	authdef "github.com/pivaldi/mmw-contracts/go/application/auth"
	authv1 "github.com/pivaldi/mmw-contracts/go/network/auth/v1"
	"github.com/rotisserie/eris"
)

// ContractAdapter wraps AuthApplicationService and implements both AuthPublicService
// and AuthPrivateService, translating between proto-typed requests/responses and
// domain-idiomatic signatures.
type ContractAdapter struct {
	svc *application.AuthApplicationService
}

var _ authdef.AuthPublicService = (*ContractAdapter)(nil)
var _ authdef.AuthPrivateService = (*ContractAdapter)(nil)

// NewContractAdapter creates a ContractAdapter around svc.
func NewContractAdapter(svc *application.AuthApplicationService) *ContractAdapter {
	return &ContractAdapter{svc: svc}
}

func (a *ContractAdapter) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	userID, err := a.svc.Register(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, mapper.DomainErrorFor(err)
	}

	return &authv1.RegisterResponse{UserId: userID.String()}, nil
}

func (a *ContractAdapter) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	token, userID, err := a.svc.Login(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, mapper.DomainErrorFor(err)
	}

	return &authv1.LoginResponse{Token: token, UserId: userID.String()}, nil
}

func (a *ContractAdapter) ValidateToken(
	ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	userID, err := a.svc.ValidateToken(ctx, req.GetToken())
	if err != nil {
		return nil, mapper.DomainErrorFor(err)
	}

	return &authv1.ValidateTokenResponse{IsValid: true, UserId: userID.String()}, nil
}

func (a *ContractAdapter) ChangePassword(
	ctx context.Context, req *authv1.ChangePasswordRequest) (*authv1.ChangePasswordResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, eris.Wrap(err, "failed to parse UUID")
	}
	if err := a.svc.ChangePassword(ctx, userID, req.GetOldPassword(), req.GetNewPassword()); err != nil {
		return nil, mapper.DomainErrorFor(err)
	}

	return &authv1.ChangePasswordResponse{}, nil
}

func (a *ContractAdapter) DeleteUser(
	ctx context.Context, req *authv1.DeleteUserRequest) (*authv1.DeleteUserResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, eris.Wrap(err, "failed to parse UUID")
	}
	if err := a.svc.DeleteUser(ctx, userID); err != nil {
		return nil, mapper.DomainErrorFor(err)
	}

	return &authv1.DeleteUserResponse{}, nil
}
