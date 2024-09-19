package userpgstorage

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/user"
	"github.com/ry461ch/loyalty_system/internal/storage/transaction"
)

type UserPGStorage struct {
	db  *sql.DB
	dsn string
}

func getDDL() string {
	return `
		CREATE TABLE IF NOT EXISTS content.users (
			id UUID PRIMARY KEY,
			login VARCHAR(255) NOT NULL,
			password_hash bytea NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE UNIQUE INDEX IF NOT EXISTS users_login_idx ON content.users(login);
		CREATE INDEX IF NOT EXISTS users_created_at_idx ON content.users(created_at);
		CREATE INDEX IF NOT EXISTS users_updated_at_idx ON content.users(updated_at);
	`
}

func NewUserPGStorage(DBDsn string) *UserPGStorage {
	return &UserPGStorage{
		dsn: DBDsn,
		db:  nil,
	}
}

func (ups *UserPGStorage) InitializeUserPGStorage(ctx context.Context, db *sql.DB) error {
	if db == nil {
		newDb, err := sql.Open("pgx", ups.dsn)
		if err != nil {
			return err
		}
		db = newDb
	}
	ups.db = db

	requests := strings.Split(getDDL(), ";")
	for _, request := range requests {
		if request != "" {
			_, err := ups.db.ExecContext(ctx, request)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (ups *UserPGStorage) GetUser(ctx context.Context, login string) (*user.User, error) {
	getUserFromDb := `
		SELECT uid, login, password_hash FROM content.users WHERE login = $1;
	`
	row := ups.db.QueryRowContext(ctx, getUserFromDb, login)

	var userInDB user.User
	err := row.Scan(&userInDB.Id, &userInDB.Login, &userInDB.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.NewUserNotFoundError()
		}
		return nil, err
	}
	return &userInDB, nil
}

func (ups *UserPGStorage) InsertUser(ctx context.Context, newUser *user.User, trx *transaction.Trx) error {
	insertUserQuery := `
		INSERT INTO content.users (id, login, pass_hash) VALUES ($1, $2, $3);
	`
	_, err := trx.ExecContext(ctx, insertUserQuery, newUser.Id, newUser.Login, newUser.PasswordHash)
	return err
}

func (ups *UserPGStorage) BeginTx(ctx context.Context) (*transaction.Trx, error) {
	return transaction.BeginTx(ctx, ups.db)
}
