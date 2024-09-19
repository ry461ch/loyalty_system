package pgstorage

import (
	"context"
	"database/sql"

	"github.com/ry461ch/loyalty_system/internal/storage/postgres/balances"
	"github.com/ry461ch/loyalty_system/internal/storage/postgres/orders"
	"github.com/ry461ch/loyalty_system/internal/storage/postgres/users"
	"github.com/ry461ch/loyalty_system/internal/storage/postgres/withdrawals"
	"github.com/ry461ch/loyalty_system/internal/storage/transaction"
)

type PGStorage struct {
	*withdrawalpgstorage.WithdrawalPGStorage
	*orderpgstorage.OrderPGStorage
	*userpgstorage.UserPGStorage
	*balancepgstorage.BalancePGStorage

	dsn string
	db  *sql.DB
}

func NewPGStorage(DBDsn string) *PGStorage {
	return &PGStorage{
		dsn: DBDsn,
		db:  nil,

		WithdrawalPGStorage: withdrawalpgstorage.NewWithdrawalPGStorage(DBDsn),
		OrderPGStorage:      orderpgstorage.NewOrderPGStorage(DBDsn),
		UserPGStorage:       userpgstorage.NewUserPGStorage(DBDsn),
		BalancePGStorage:    balancepgstorage.NewBalancePGStorage(DBDsn),
	}
}

func (pg *PGStorage) Initialize(ctx context.Context) error {
	db, err := sql.Open("pgx", pg.dsn)
	if err != nil {
		return err
	}
	pg.db = db

	err = pg.InitializeUserPGStorage(ctx, db)
	if err != nil {
		return err
	}

	err = pg.InitializeOrderPGStorage(ctx, db)
	if err != nil {
		return err
	}

	err = pg.InitializeWithdrawalPGStorage(ctx, db)
	if err != nil {
		return err
	}

	err = pg.InitializeBalancePGStorage(ctx, db)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PGStorage) BeginTx(ctx context.Context) (*transaction.Trx, error) {
	return transaction.BeginTx(ctx, pg.db)
}
