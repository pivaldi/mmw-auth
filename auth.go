// services/auth/module.go
package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	ogloutbox "github.com/ovya/ogl/db/outbox"
	ogluow "github.com/ovya/ogl/pg/uow"
	oglevents "github.com/ovya/ogl/platform/events"
	oglserver "github.com/ovya/ogl/platform/server"
	"github.com/pivaldi/mmw-auth/config"
	"github.com/pivaldi/mmw-auth/internal/adapters/inbound/connect"
	outboxevents "github.com/pivaldi/mmw-auth/internal/adapters/outbound/events"
	"github.com/pivaldi/mmw-auth/internal/adapters/outbound/persistence/postgres"
	"github.com/pivaldi/mmw-auth/internal/application"
	domainUser "github.com/pivaldi/mmw-auth/internal/domain/auth/user"
	defauth "github.com/pivaldi/mmw-contracts/definitions/auth"
	"github.com/pivaldi/mmw-contracts/gen/go/auth/v1/authv1connect"
	"github.com/rotisserie/eris"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/errgroup"
)

const (
	relayTableName = "auth.event"
	ModuleName     = "Auth"
)

var NotifyEvents = domainUser.AllEvents

// module implements oglcore.module for the auth service.
type module struct {
	relay       *ogloutbox.EventsRelay
	server      *oglserver.HTTPServer
	logger      *slog.Logger
	authService *application.AuthApplicationService
}

// The infrastructure need to start the Auth App.
type Infrastructure struct {
	DBPool   *pgxpool.Pool
	EventBus oglevents.SystemEventBus
	Logger   *slog.Logger
}

// New wires the auth module with all its dependencies.
func New(infra Infrastructure) (*module, error) {
	cfg, err := config.Load(context.Background(), "")
	if err != nil {
		return nil, eris.Wrap(err, "app failed to load config")
	}

	userRepo := postgres.NewUserRepository(infra.DBPool)
	sessionRepo := postgres.NewSessionRepository(infra.DBPool)
	uow := ogluow.NewUnitOfWork(infra.DBPool)
	dispatcher := outboxevents.NewOutboxDispatcher(infra.DBPool)

	authService := application.NewAuthService(userRepo, sessionRepo, uow, dispatcher, cfg.JWT.Secret)
	authHandler := connect.NewAuthHandler(authService)

	mux := http.NewServeMux()
	path, handler := authv1connect.NewAuthServiceHandler(authHandler)
	mux.Handle(path, handler)

	h2cHandler := h2c.NewHandler(mux, &http2.Server{})
	httpInfra := oglserver.HTTPServerInfra{
		Config:      cfg.Server,
		Handler:     h2cHandler,
		Logger:      infra.Logger,
		HealthFns:   oglserver.HealthFns{"database": userRepo.Health},
		LogPayloads: true,
	}

	server := oglserver.NewHTTPServer2(cfg.Environment.IsDev(), httpInfra)

	return &module{
		relay:       ogloutbox.NewEnventsRelay(infra.DBPool, infra.EventBus, infra.Logger, relayTableName),
		server:      server,
		logger:      infra.Logger,
		authService: authService,
	}, nil
}

// Start runs the HTTP server and the outbox relay concurrently.
func (m *module) Start(ctx context.Context) error {
	m.logger.Info("starting auth module")
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return m.server.Start(gCtx)
	})

	g.Go(func() error {
		m.relay.Start(gCtx)
		return nil
	})

	return eris.Wrapf(g.Wait(), "%s failure", ModuleName)
}

// Ensure Module implements defauth.AuthService so it can be passed directly to
// defauth.NewInprocClient without an intermediate wrapper.
var _ defauth.AuthService = (*module)(nil)

// GetUser delegates to the internal auth application service.
func (m *module) GetUser(ctx context.Context, id string) (*defauth.UserDTO, error) {
	u, err := m.authService.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}

	return u, nil
}

// ValidateToken delegates to the internal auth application service.
func (m *module) ValidateToken(ctx context.Context, token string) (uuid.UUID, error) {
	id, err := m.authService.ValidateToken(ctx, token)
	if err != nil {
		return uuid.Nil, fmt.Errorf("validating token: %w", err)
	}

	return id, nil
}
