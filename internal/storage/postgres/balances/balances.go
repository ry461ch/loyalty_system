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
	DB  *sql.DB
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
		DB:  nil,
	}
}

func (bps *BalancePGStorage) InitializeBalancePGStorage(ctx context.Context, DB *sql.DB) error {
	if DB == nil {
		newDB, err := sql.Open("pgx", bps.dsn)
		if err != nil {
			return err
		}
		DB = newDB
	}
	bps.DB = DB

	requests := strings.Split(getDDL(), ";")
	for _, request := range requests {
		if request != "" {
			_, err := bps.DB.ExecContext(ctx, request)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (bps *BalancePGStorage) GetBalance(ctx context.Context, userID uuid.UUID) (*balance.Balance, error) {
	getBalanceFromDB := `
		SELECT current, withdrawn FROM content.balances WHERE user_id = $1;
	`
	row := bps.DB.QueryRowContext(ctx, getBalanceFromDB, userID.String())

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

func (bps *BalancePGStorage) ReduceBalance(ctx context.Context, userID uuid.UUID, amount float64, tx *transaction.Trx) error {
	spendAmountQuery := `
		UPDATE content.balances
		SET
			current = current - $2,
			withdrawn = withdraw + $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1;
	`

	if tx == nil {
		var err error
		tx, err = bps.BeginTx(ctx)
		if err != nil {
			return err
		}
	}

	_, err := tx.ExecContext(ctx, spendAmountQuery, userID, amount)
	return err
}

func (bps *BalancePGStorage) AddBalance(ctx context.Context, userID uuid.UUID, amount float64, tx *transaction.Trx) error {
	insertBalanceQuery := `
		INSERT INTO content.balances (user_id, current)
		VALUES ($1, $2)
		ON CONFLICT (user_id)
		DO UPDATE SET
			current = current + $2,
			updated_at = CURRENT_TIMESTAMP
		;
	`
	if tx == nil {
		var err error
		tx, err = bps.BeginTx(ctx)
		if err != nil {
			return err
		}
	}

	_, err := tx.ExecContext(ctx, insertBalanceQuery, userID, amount)
	return err
}

func (bps *BalancePGStorage) BeginTx(ctx context.Context) (*transaction.Trx, error) {
	return transaction.BeginTx(ctx, bps.DB)
}
