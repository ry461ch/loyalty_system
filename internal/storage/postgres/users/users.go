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
	DB  *sql.DB
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
		DB:  nil,
	}
}

func (ups *UserPGStorage) Initialize(ctx context.Context, DB *sql.DB) error {
	if DB == nil {
		newDB, err := sql.Open("pgx", ups.dsn)
		if err != nil {
			return err
		}
		DB = newDB
	}
	ups.DB = DB

	requests := strings.Split(getDDL(), ";")
	for _, request := range requests {
		if request != "" {
			_, err := ups.DB.ExecContext(ctx, request)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (ups *UserPGStorage) GetUser(ctx context.Context, login string) (*user.User, error) {
	getUserFromDB := `
		SELECT id, login, password_hash FROM content.users WHERE login = $1;
	`
	row := ups.DB.QueryRowContext(ctx, getUserFromDB, login)

	var userInDB user.User
	err := row.Scan(&userInDB.ID, &userInDB.Login, &userInDB.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.NewUserNotFoundError()
		}
		return nil, err
	}
	return &userInDB, nil
}

func (ups *UserPGStorage) InsertUser(ctx context.Context, newUser *user.User, tx *transaction.Trx) error {
	insertUserQuery := `
		INSERT INTO content.users (id, login, password_hash) VALUES ($1, $2, $3);
	`

	_, err := tx.ExecContext(ctx, insertUserQuery, newUser.ID, newUser.Login, newUser.PasswordHash)
	return err
}

func (ups *UserPGStorage) BeginTx(ctx context.Context) (*transaction.Trx, error) {
	return transaction.BeginTx(ctx, ups.DB)
}
