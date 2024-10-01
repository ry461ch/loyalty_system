package transaction

import (
	"context"
	"database/sql"
)

type Trx struct {
	*sql.Tx
}

func (t *Trx) Commit() error {
	if t.Tx != nil {
		return t.Tx.Commit()
	}
	return nil
}

func (t *Trx) Rollback() error {
	if t.Tx != nil {
		return t.Tx.Rollback()
	}
	return nil
}

func BeginTx(ctx context.Context, DB *sql.DB) (*Trx, error) {
	if DB != nil {
		sqlTx, err := DB.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		return &Trx{Tx: sqlTx}, nil
	}
	return &Trx{}, nil
}
