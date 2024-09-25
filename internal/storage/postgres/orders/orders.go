package orderpgstorage

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/helpers/transaction"
	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/order"
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
			accrual	DOUBLE PRECISION,
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

func (ops *OrderPGStorage) Initialize(ctx context.Context, DB *sql.DB) error {
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
			return nil, exceptions.ErrOrderNotFound
		}
		return nil, err
	}
	return &userID, nil
}

func (ops *OrderPGStorage) InsertOrder(ctx context.Context, userID uuid.UUID, orderID string, tx *transaction.Trx) error {
	insertOrderQuery := `
		INSERT INTO content.orders (id, user_id) VALUES ($1, $2);
	`

	_, err := tx.ExecContext(ctx, insertOrderQuery, orderID, userID)
	return err
}

func (ops *OrderPGStorage) UpdateOrder(ctx context.Context, order *order.Order, tx *transaction.Trx) (*uuid.UUID, error) {
	updateOrderQuery := `
		UPDATE content.orders
		SET
			status = $2,
			accrual = $3,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING user_id;
	`
	var accrual sql.NullFloat64
	if order.Accrual != nil {
		accrual = sql.NullFloat64{
			Float64: *order.Accrual,
			Valid:   true,
		}
	}

	row := tx.QueryRowContext(ctx, updateOrderQuery, order.ID, order.Status, accrual)
	var userID uuid.UUID
	err := row.Scan(&userID)
	if err != nil {
		return nil, err
	}

	return &userID, nil
}

func (ops *OrderPGStorage) GetWaitingOrders(ctx context.Context, limit int, inputCreatedAt *time.Time) ([]order.Order, error) {
	getOrdersFromDB := `
		SELECT id, status, accrual, created_at
		FROM content.orders
		WHERE 
			(status = 'NEW' OR status = 'PROCESSING') AND
			($2::TIMESTAMPTZ IS NULL OR created_at < $2::TIMESTAMPTZ)
		ORDER BY created_at DESC
		LIMIT $1;
	`

	var createdAt sql.NullTime
	if inputCreatedAt != nil {
		createdAt.Valid = true
		createdAt.Time = *inputCreatedAt
	}

	rows, err := ops.DB.QueryContext(ctx, getOrdersFromDB, limit, createdAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	waitingOrders := []order.Order{}
	for rows.Next() {
		var waitingOrder order.Order
		var accrual sql.NullFloat64

		err = rows.Scan(&waitingOrder.ID, &waitingOrder.Status, &accrual, &waitingOrder.CreatedAt)
		if err != nil {
			return nil, err
		}

		if accrual.Valid {
			waitingOrder.Accrual = &accrual.Float64
		}

		waitingOrders = append(waitingOrders, waitingOrder)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return waitingOrders, nil
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
