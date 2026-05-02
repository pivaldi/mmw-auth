// services/auth/module.go
package auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	pfdbmigrator "github.com/piprim/mmw/pkg/platform/db/migrator"
	pfoutbox "github.com/piprim/mmw/pkg/platform/db/outbox"
	pfevents "github.com/piprim/mmw/pkg/platform/events"
	pfuow "github.com/piprim/mmw/pkg/platform/pg/uow"
	pfserver "github.com/piprim/mmw/pkg/platform/server"
	"github.com/pivaldi/mmw-auth/internal/adapters/inbound/connect"
	"github.com/pivaldi/mmw-auth/internal/adapters/inbound/inproc"
	outboxevents "github.com/pivaldi/mmw-auth/internal/adapters/outbound/events"
	"github.com/pivaldi/mmw-auth/internal/adapters/outbound/persistence/postgres"
	"github.com/pivaldi/mmw-auth/internal/application"
	"github.com/pivaldi/mmw-auth/internal/infra/config"
	"github.com/pivaldi/mmw-auth/internal/infra/persistence/migrations"
	authdef "github.com/pivaldi/mmw-contracts/go/application/auth"
	"github.com/pivaldi/mmw-contracts/go/network/auth/v1/authv1connect"
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"
)

const (
	relayTableName = "auth.event"
	ModuleName     = "Auth"
	PGSchema       = "auth"
)

// Module implements pfcore.Module for the auth service.
type Module struct {
	relay   *pfoutbox.EventsRelay
	server  *pfserver.HTTPServer
	logger  *slog.Logger
	service *application.AuthApplicationService
}

// PublicService returns the auth module's public-facing operations as AuthPublicService.
// Callers that only need registration, login, and password management use this accessor.
func (m *Module) PublicService() authdef.AuthPublicService {
	return inproc.NewContractAdapter(m.service)
}

// PrivateService returns the auth module's internal operations as AuthPrivateService.
// Use this when the caller only needs to validate tokens (e.g. the todo module).
func (m *Module) PrivateService() authdef.AuthPrivateService {
	return inproc.NewContractAdapter(m.service)
}

// Handler returns the module's HTTP handler so tests can wrap it in
// httptest.NewServer without starting a real server on a port.
func (m *Module) Handler() http.Handler {
	return m.server.Handler()
}

// Migrate runs all pending database migrations for the auth module.
// Intended for use in tests and migration tooling.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	m, err := pfdbmigrator.New(db, migrations.FS, "scripts", PGSchema)
	if err != nil {
		return eris.Wrap(err, "failed to create migrator")
	}

	_, err = m.Up(ctx)

	return eris.Wrap(err, "failed to migrate up")
}

// The infrastructure need to start the Auth App.
type Infrastructure struct {
	DBPool   *pgxpool.Pool
	EventBus pfevents.SystemEventBus
	Logger   *slog.Logger
}

// New wires the auth module with all its dependencies.
func New(ctx context.Context, infra Infrastructure) (*Module, error) {
	cfg, err := config.Load(ctx, "")
	if err != nil {
		return nil, eris.Wrap(err, "app failed to load config")
	}

	uow := pfuow.New(infra.DBPool)
	userRepo := postgres.NewUserRepository(uow)
	sessionRepo := postgres.NewSessionRepository(uow)
	dispatcher := outboxevents.NewOutboxDispatcher(uow)

	authService := application.NewAuthService(userRepo, sessionRepo, uow, dispatcher, cfg.JWT.Secret)
	authHandler := connect.NewAuthHandler(authService)

	mux := http.NewServeMux()
	pubPath, pubHandler := authv1connect.NewAuthPublicServiceHandler(authHandler)
	privPath, privHandler := authv1connect.NewAuthPrivateServiceHandler(authHandler)
	mux.Handle(pubPath, pubHandler)
	mux.Handle(privPath, privHandler)

	httpInfra := pfserver.HTTPServerInfra{
		Config:      cfg.Server,
		Handler:     mux,
		Logger:      infra.Logger,
		HealthFns:   pfserver.HealthFns{"database": userRepo.Health},
		LogPayloads: true,
		ServiceNames: []string{
			authv1connect.AuthPublicServiceName,
			authv1connect.AuthPrivateServiceName,
		},
	}

	server := pfserver.NewHTTPServer(httpInfra)

	eventRelay, err := pfoutbox.NewEventsRelay(infra.DBPool, infra.EventBus, infra.Logger, relayTableName)
	if err != nil {
		return nil, eris.Wrap(err, "failed to create events relay")
	}

	return &Module{
		relay:   eventRelay,
		server:  server,
		logger:  infra.Logger,
		service: authService,
	}, nil
}

// Start runs the HTTP server and the outbox relay concurrently.
func (m *Module) Start(ctx context.Context) error {
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
