package balancepgstorage

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/models/balance"
	"github.com/ry461ch/loyalty_system/internal/storage/transaction"
)

type BalancePGStorage struct {
	db  *sql.DB
	dsn string
}

func getDDL() string {
	return `
		CREATE TABLE IF NOT EXISTS content.balances (
			user_id UUID PRIMARY KEY,
			current	INTEGER NOT NULL DEFAULT 0,
			withdrawn INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS balances_created_at_idx ON content.balances(created_at);
		CREATE INDEX IF NOT EXISTS balances_updated_at_idx ON content.balances(updated_at);
	`
}

func NewBalancePGStorage(DBDsn string) *BalancePGStorage {
	return &BalancePGStorage{
		dsn: DBDsn,
		db:  nil,
	}
}

func (bps *BalancePGStorage) InitializeBalancePGStorage(ctx context.Context, db *sql.DB) error {
	if db == nil {
		newDb, err := sql.Open("pgx", bps.dsn)
		if err != nil {
			return err
		}
		db = newDb
	}
	bps.db = db

	requests := strings.Split(getDDL(), ";")
	for _, request := range requests {
		if request != "" {
			_, err := bps.db.ExecContext(ctx, request)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (bps *BalancePGStorage) GetBalance(ctx context.Context, userId uuid.UUID) (*balance.Balance, error) {
	getBalanceFromDb := `
		SELECT current, withdrawn FROM content.balances WHERE user_id = $1;
	`
	row := bps.db.QueryRowContext(ctx, getBalanceFromDb, userId.String())

	var userBalance balance.Balance
	err := row.Scan(&userBalance.Current, &userBalance.Withdrawn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &balance.Balance{}, nil
		}
		return nil, err
	}
	return &userBalance, nil
}

func (bps *BalancePGStorage) ReduceBalance(ctx context.Context, userId uuid.UUID, amount float64, trx *transaction.Trx) error {
	spendAmountQuery := `
		UPDATE content.balances
		SET
			current = current - $2,
			withdrawn = withdraw + $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1;
	`
	_, err := trx.ExecContext(ctx, spendAmountQuery, userId, amount)
	return err
}

func (bps *BalancePGStorage) AddBalance(ctx context.Context, userId uuid.UUID, amount float64, trx transaction.Trx) error {
	insertBalanceQuery := `
		INSERT INTO content.balances (user_id, current)
		VALUES ($1, $2)
		ON CONFLICT (user_id)
		DO UPDATE SET
			current = current + $2,
			updated_at = CURRENT_TIMESTAMP
		;
	`
	_, err := trx.ExecContext(ctx, insertBalanceQuery, userId, amount)
	return err
}

func (bps *BalancePGStorage) BeginTx(ctx context.Context) (*transaction.Trx, error) {
	return transaction.BeginTx(ctx, bps.db)
}
