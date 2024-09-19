package withdrawal

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
)

type Withdrawal struct {
	Id             uuid.UUID `json:"-"`
	OrderId        string    `json:"number"`
	Sum       float64     `json:"sum"`
	CreatedAt time.Time `json:"processed_at"`
}

func (w *Withdrawal) UnmarshalJSON(data []byte) error {
	type WithdrawalAlias Withdrawal

	aliasValue := &struct {
		*WithdrawalAlias
		OrderId string `json:"order"`
	}{
		WithdrawalAlias: (*WithdrawalAlias)(w),
	}

	if err := json.Unmarshal(data, aliasValue); err != nil {
		return err
	}

	if aliasValue.Sum == 0 {
		return exceptions.NewBalanceBadAmountFormatError()
	}

	w.OrderId = aliasValue.OrderId
	w.CreatedAt = time.Now()

	return nil
}
