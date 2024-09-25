package pgstorage

import (
	"context"
	"database/sql"
	"strings"

	"github.com/ry461ch/loyalty_system/internal/storage/postgres/balances"
	"github.com/ry461ch/loyalty_system/internal/storage/postgres/orders"
	"github.com/ry461ch/loyalty_system/internal/storage/postgres/users"
	"github.com/ry461ch/loyalty_system/internal/storage/postgres/withdrawals"
)

type PGStorage struct {
	connectionsLimit  int
	dsn               string
	DB                *sql.DB
	OrderStorage      *orderpgstorage.OrderPGStorage
	WithdrawalStorage *withdrawalpgstorage.WithdrawalPGStorage
	BalanceStorage    *balancepgstorage.BalancePGStorage
	UserStorage       *userpgstorage.UserPGStorage
}

func NewPGStorage(DBDsn string, connectionsLimit int) *PGStorage {
	return &PGStorage{
		connectionsLimit:  connectionsLimit,
		dsn:               DBDsn,
		OrderStorage:      orderpgstorage.NewOrderPGStorage(DBDsn),
		UserStorage:       userpgstorage.NewUserPGStorage(DBDsn),
		BalanceStorage:    balancepgstorage.NewBalancePGStorage(DBDsn),
		WithdrawalStorage: withdrawalpgstorage.NewWithdrawalPGStorage(DBDsn),
	}
}

func getDDL() string {
	return `
		CREATE SCHEMA IF NOT EXISTS content;
	`
}

func (ps *PGStorage) Init(ctx context.Context) error {
	DB, err := sql.Open("pgx", ps.dsn)
	if err != nil {
		return err
	}
	DB.SetMaxIdleConns(ps.connectionsLimit)
	DB.SetMaxOpenConns(ps.connectionsLimit)

	requests := strings.Split(getDDL(), ";")
	for _, request := range requests {
		if request != "" {
			_, err := DB.ExecContext(ctx, request)
			if err != nil {
				return err
			}
		}
	}

	err = ps.BalanceStorage.Initialize(ctx, DB)
	if err != nil {
		return err
	}

	err = ps.OrderStorage.Initialize(ctx, DB)
	if err != nil {
		return err
	}

	err = ps.UserStorage.Initialize(ctx, DB)
	if err != nil {
		return err
	}

	err = ps.WithdrawalStorage.Initialize(ctx, DB)
	if err != nil {
		return err
	}

	ps.DB = DB
	return nil
}

func (ps *PGStorage) Close() {
	ps.DB.Close()
}
