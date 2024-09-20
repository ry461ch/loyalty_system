package orderpgstorage

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/order"
	"github.com/ry461ch/loyalty_system/internal/storage/transaction"
)

type OrderPGStorage struct {
	DB  *sql.DB
	dsn string
}

func getDDL() string {
	return `
		CREATE TABLE IF NOT EXISTS content.orders (
			id VARCHAR(255) PRIMARY KEY,
			status VARCHAR(255) NOT NULL default 'NEW',
			accrual	INTEGER,
			user_id UUID NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS orders_user_id_idx ON content.orders(user_id);
		CREATE INDEX IF NOT EXISTS orders_created_at_idx ON content.orders(created_at);
		CREATE INDEX IF NOT EXISTS orders_updated_at_idx ON content.orders(updated_at);
	`
}

func NewOrderPGStorage(DBDsn string) *OrderPGStorage {
	return &OrderPGStorage{
		dsn: DBDsn,
		DB:  nil,
	}
}

func (ops *OrderPGStorage) InitializeOrderPGStorage(ctx context.Context, DB *sql.DB) error {
	if DB == nil {
		newDB, err := sql.Open("pgx", ops.dsn)
		if err != nil {
			return err
		}
		DB = newDB
	}
	ops.DB = DB

	requests := strings.Split(getDDL(), ";")
	for _, request := range requests {
		if request != "" {
			_, err := ops.DB.ExecContext(ctx, request)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (ops *OrderPGStorage) GetOrderUserID(ctx context.Context, orderID string) (*uuid.UUID, error) {
	getOrderFromDB := `
		SELECT user_id FROM content.orders WHERE id = $1;
	`
	row := ops.DB.QueryRowContext(ctx, getOrderFromDB, orderID)
	var userID uuid.UUID
	err := row.Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.NewOrderNotFoundError()
		}
		return nil, err
	}
	return &userID, nil
}

func (ops *OrderPGStorage) InsertOrder(ctx context.Context, userID uuid.UUID, orderID string, tx *transaction.Trx) error {
	insertOrderQuery := `
		INSERT INTO content.orders (id, user_id) VALUES ($1, $2);
	`

	if tx == nil {
		var err error
		tx, err = ops.BeginTx(ctx)
		if err != nil {
			return err
		}
	}

	_, err := tx.ExecContext(ctx, insertOrderQuery, orderID, userID)
	return err
}

func (ops *OrderPGStorage) UpdateOrder(ctx context.Context, order *order.Order, tx *transaction.Trx) error {
	updateOrderQuery := `
		UPDATE content.orders
		SET
			status = $2,
			accrual = $3,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1;
	`
	var accrual sql.NullFloat64
	if order.Accrual != nil {
		accrual = sql.NullFloat64{
			Float64: *order.Accrual,
			Valid:   true,
		}
	}

	if tx == nil {
		var err error
		tx, err = ops.BeginTx(ctx)
		if err != nil {
			return err
		}
	}

	_, err := tx.ExecContext(ctx, updateOrderQuery, order.ID, order.Status, accrual)
	return err
}

func (ops *OrderPGStorage) GetWaitingOrderIDs(ctx context.Context) ([]string, error) {
	getOrdersFromDB := `
		SELECT id
		FROM content.orders
		WHERE status = 'NEW' OR status = 'PROCESSING'
		ORDER BY created_at ASC;
	`

	rows, err := ops.DB.QueryContext(ctx, getOrdersFromDB)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orderIDs := []string{}
	for rows.Next() {
		var orderID string

		err = rows.Scan(&orderID)
		if err != nil {
			return nil, err
		}

		orderIDs = append(orderIDs, orderID)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return orderIDs, nil
}

func (ops *OrderPGStorage) GetUserOrders(ctx context.Context, userID uuid.UUID) ([]order.Order, error) {
	getOrdersFromDB := `
		SELECT id, status, accrual, created_at
		FROM content.orders
		WHERE user_id = $1
		ORDER BY created_at DESC;
	`

	rows, err := ops.DB.QueryContext(ctx, getOrdersFromDB, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []order.Order{}
	for rows.Next() {
		var orderRow order.Order
		var accrual sql.NullFloat64

		err = rows.Scan(&orderRow.ID, &orderRow.Status, &accrual, &orderRow.CreatedAt)
		if err != nil {
			return nil, err
		}
		if accrual.Valid {
			orderRow.Accrual = &accrual.Float64
		}

		orders = append(orders, orderRow)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (ops *OrderPGStorage) BeginTx(ctx context.Context) (*transaction.Trx, error) {
	return transaction.BeginTx(ctx, ops.DB)
}
