package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	pfdb "github.com/piprim/mmw/platform/db"
	pfpguow "github.com/piprim/mmw/platform/pg/uow"
	"github.com/pivaldi/mmw-auth/internal/domain/user"
	"github.com/rotisserie/eris"
)

// UserRepository is the PostgreSQL implementation of ports.UserRepository.
type UserRepository struct {
	uow *pfpguow.UnitOfWork
}

func NewUserRepository(uow *pfpguow.UnitOfWork) *UserRepository {
	return &UserRepository{uow: uow}
}

func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	exec := r.uow.Executor(ctx)
	_, err := exec.Exec(ctx,
		`INSERT INTO auth.users (id, login, password_hash, created_at, updated_at)
		 VALUES (@id, @login, @password_hash, @created_at, @updated_at)`,
		pgx.NamedArgs(pfdb.StructArgs(u.Snapshot())),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return user.ErrUserAlreadyExists
		}

		return eris.Wrap(err, "save user")
	}

	return nil
}

func (r *UserRepository) FindByLogin(ctx context.Context, login user.Login) (*user.User, error) {
	exec := r.uow.Executor(ctx)
	rows, err := exec.Query(ctx,
		`SELECT id, login, password_hash, created_at, updated_at FROM auth.users WHERE login = $1`,
		login.String(),
	)
	if err != nil {
		return nil, eris.Wrap(err, "query user by login")
	}

	snap, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[user.UserSnapshot])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, eris.New("user not found")
		}

		return nil, eris.Wrap(err, "collect user row")
	}

	return user.ReconstructUser(&snap), nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	exec := r.uow.Executor(ctx)
	rows, err := exec.Query(ctx,
		`SELECT id, login, password_hash, created_at, updated_at FROM auth.users WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, eris.Wrap(err, "query user by id")
	}

	snap, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[user.UserSnapshot])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, eris.New("user not found")
		}

		return nil, eris.Wrap(err, "collect user row")
	}

	return user.ReconstructUser(&snap), nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	exec := r.uow.Executor(ctx)
	_, err := exec.Exec(ctx,
		`UPDATE auth.users SET login = @login, password_hash = @password_hash, updated_at = @updated_at WHERE id = @id`,
		pgx.NamedArgs(pfdb.StructArgs(u.Snapshot())),
	)

	return eris.Wrap(err, "update user")
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	exec := r.uow.Executor(ctx)
	_, err := exec.Exec(ctx, `DELETE FROM auth.users WHERE id = $1`, id)

	return eris.Wrap(err, "delete user")
}

func (r *UserRepository) Health(ctx context.Context) (any, error) {
	exec := r.uow.Executor(ctx)
	row := exec.QueryRow(ctx, "SELECT count(*) FROM auth.users")
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, eris.Wrap(err, "scan row")
	}

	return count, nil
}
