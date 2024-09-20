package withdrawalpgstorage

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/withdrawal"
	"github.com/ry461ch/loyalty_system/internal/storage/transaction"
)

type WithdrawalPGStorage struct {
	db  *sql.DB
	dsn string
}

func getDDL() string {
	return `
		CREATE TABLE IF NOT EXISTS content.withdrawals (
			id UUID PRIMARY KEY,
			order_id VARCHAR(255) NOT NULL,
			user_id UUID NOT NULL,
			sum	INTEGER NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS withdrawals_order_id_idx ON content.withdrawals(order_id);
		CREATE INDEX IF NOT EXISTS withdrawals_user_id_idx ON content.withdrawals(user_id);
		CREATE INDEX IF NOT EXISTS withdrawals_created_at_idx ON content.withdrawals(created_at);
	`
}

func NewWithdrawalPGStorage(DBDsn string) *WithdrawalPGStorage {
	return &WithdrawalPGStorage{
		dsn: DBDsn,
		db:  nil,
	}
}

func (wps *WithdrawalPGStorage) InitializeWithdrawalPGStorage(ctx context.Context, db *sql.DB) error {
	if db == nil {
		newDb, err := sql.Open("pgx", wps.dsn)
		if err != nil {
			return err
		}
		db = newDb
	}
	wps.db = db

	requests := strings.Split(getDDL(), ";")
	for _, request := range requests {
		if request != "" {
			_, err := wps.db.ExecContext(ctx, request)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (wps *WithdrawalPGStorage) InsertWithdrawal(ctx context.Context, inputWithdrawal *withdrawal.Withdrawal, trx *transaction.Trx) error {
	insertWithdrawalQuery := `
		INSERT INTO content.withdrawals (id, order_id, user_id, sum)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING;
	`
	_, err := trx.ExecContext(ctx, insertWithdrawalQuery, *inputWithdrawal.Id, inputWithdrawal.OrderId, *inputWithdrawal.UserId, inputWithdrawal.Sum)
	return err
}

func (wps *WithdrawalPGStorage) GetWithdrawals(ctx context.Context, userId uuid.UUID) ([]withdrawal.Withdrawal, error) {
	getWithdrawalsFromDb := `
		SELECT id, order_id, sum, created_at
		FROM content.withdrawals
		WHERE user_id = $1
		ORDER BY created_at DESC;
	`
	rows, err := wps.db.QueryContext(ctx, getWithdrawalsFromDb, userId.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	withdrawals := []withdrawal.Withdrawal{}
	for rows.Next() {
		var insertedWithdrawal withdrawal.Withdrawal
		err = rows.Scan(&insertedWithdrawal.Id, &insertedWithdrawal.OrderId, &insertedWithdrawal.Sum, &insertedWithdrawal.CreatedAt)
		if err != nil {
			return nil, err
		}

		withdrawals = append(withdrawals, insertedWithdrawal)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return withdrawals, nil
}

func (wps *WithdrawalPGStorage) GetWithdrawal(ctx context.Context, Id uuid.UUID) (*withdrawal.Withdrawal, error) {
	getWithdrawalFromDb := `
		SELECT order_id, sum, created_at
		FROM content.withdrawals
		WHERE id = $1;
	`
	row := wps.db.QueryRowContext(ctx, getWithdrawalFromDb, Id)

	var withdrawalInDb withdrawal.Withdrawal
	err := row.Scan(&withdrawalInDb.OrderId, &withdrawalInDb.Sum, &withdrawalInDb.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.NewWithdrawalNotFoundError()
		}
		return nil, err
	}
	withdrawalInDb.Id = &Id
	return &withdrawalInDb, nil
}

func (wps *WithdrawalPGStorage) BeginTx(ctx context.Context) (*transaction.Trx, error) {
	return transaction.BeginTx(ctx, wps.db)
}
