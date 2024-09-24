package withdrawal

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
)

type Withdrawal struct {
	ID        *uuid.UUID `json:"-"`
	UserID    *uuid.UUID `json:"-"`
	OrderID   string     `json:"order"`
	Sum       float64    `json:"sum"`
	CreatedAt *time.Time `json:"processed_at"`
}

func (w *Withdrawal) UnmarshalJSON(data []byte) error {
	type WithdrawalAlias Withdrawal

	aliasValue := &struct {
		*WithdrawalAlias
	}{
		WithdrawalAlias: (*WithdrawalAlias)(w),
	}

	if err := json.Unmarshal(data, aliasValue); err != nil {
		return err
	}

	if aliasValue.Sum == 0 {
		return exceptions.ErrBalanceBadAmountFormat
	}
	if aliasValue.UserID != nil || aliasValue.ID != nil || aliasValue.CreatedAt != nil {
		return exceptions.ErrWithdrawalBadFormat
	}

	return nil
}
