package order

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
)

type Status int32

const (
	NEW Status = iota
	PROCESSING
	INVALID
	PROCESSED
)

func (s Status) MarshalJSON() ([]byte, error) {
	switch s {
	case NEW:
		return []byte("\"NEW\""), nil
	case PROCESSING:
		return []byte("\"PROCESSING\""), nil
	case INVALID:
		return []byte("\"INVALID\""), nil
	case PROCESSED:
		return []byte("\"PROCESSED\""), nil
	default:
		return nil, exceptions.ErrOrderBadStatusFormat
	}
}

func (s Status) Value() (driver.Value, error) {
	switch s {
	case NEW:
		return "NEW", nil
	case PROCESSING:
		return "PROCESSING", nil
	case PROCESSED:
		return "PROCESSED", nil
	case INVALID:
		return "INVALID", nil
	default:
		return nil, errors.New("invalid status")
	}
}

func (s *Status) Scan(value interface{}) error {
	if value == nil {
		*s = NEW
		return nil
	}

	if sv, err := driver.String.ConvertValue(value); err == nil {
		if v, ok := sv.(string); ok {
			switch v {
			case "NEW":
				*s = NEW
			case "PROCESSING":
				*s = PROCESSING
			case "PROCESSED":
				*s = PROCESSED
			case "INVALID":
				*s = INVALID
			default:
				return errors.New("invalid status")
			}
			return nil
		}
	}

	return errors.New("failed to scan Status")
}

func (s *Status) UnmarshalJSON(data []byte) error {
	switch {
	case bytes.Equal(data, []byte("\"REGISTERED\"")):
		*s = NEW
	case bytes.Equal(data, []byte("\"PROCESSING\"")):
		*s = PROCESSING
	case bytes.Equal(data, []byte("\"INVALID\"")):
		*s = INVALID
	case bytes.Equal(data, []byte("\"PROCESSED\"")):
		*s = PROCESSED
	default:
		return exceptions.ErrOrderBadStatusFormat
	}
	return nil
}

type Order struct {
	ID        string    `json:"number"`
	Status    Status    `json:"status"`
	Accrual   *float64  `json:"accrual,omitempty"`
	CreatedAt time.Time `json:"uploaded_at"`
}

func (o *Order) UnmarshalJSON(data []byte) error {
	type OrderAlias Order

	aliasValue := &struct {
		*OrderAlias
		ID string `json:"order"`
	}{
		OrderAlias: (*OrderAlias)(o),
	}

	if err := json.Unmarshal(data, aliasValue); err != nil {
		return err
	}

	o.ID = aliasValue.ID

	return nil
}
