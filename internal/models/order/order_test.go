package order

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshal(t *testing.T) {
	accural := float64(500)
	datetime := time.Date(2020, 12, 9, 16, 9, 53, 0, &time.Location{})

	testCases := []struct {
		testName      string
		inputOrder    string
		expectedOrder Order
	}{
		{
			testName: "PROCESSED with accrual",
			inputOrder: `{
				"order": "1321",
				"status": "PROCESSED",
				"accrual": 500
			}`,
			expectedOrder: Order{
				Id:        "1321",
				Status:    PROCESSED,
				Accrual:   &accural,
				CreatedAt: datetime,
			},
		},
		{
			testName: "REGISTERED without accrual",
			inputOrder: `{
				"order": "1321",
				"status": "REGISTERED"
			}`,
			expectedOrder: Order{
				Id:     "1321",
				Status: NEW,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			var order Order
			err := json.Unmarshal([]byte(tc.inputOrder), &order)
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedOrder.Id, order.Id, "Id not equal")
			if tc.expectedOrder.Accrual != nil {
				assert.Equal(t, *tc.expectedOrder.Accrual, *order.Accrual, "Accrual not equal")
			}
			assert.Equal(t, tc.expectedOrder.Status, order.Status, "Status not equal")
		})
	}
}

func TestMarshal(t *testing.T) {
	accural := float64(500)
	datetime := time.Date(2020, 12, 9, 16, 9, 53, 0, &time.Location{})

	testCases := []struct {
		testName     string
		Order        Order
		expectedJSON string
	}{
		{
			testName: "processed with accrual",
			Order: Order{
				Id:        "1321",
				Status:    PROCESSED,
				Accrual:   &accural,
				CreatedAt: datetime,
			},
			expectedJSON: `{
				"number": "1321",
				"status": "PROCESSED",
				"accrual": 500,
				"uploaded_at": "2020-12-09T16:09:53Z"
			}`,
		},
		{
			testName: "processed without accrual",
			Order: Order{
				Id:        "1321",
				Status:    NEW,
				CreatedAt: datetime,
			},
			expectedJSON: `{
				"number": "1321",
				"status": "NEW",
				"uploaded_at": "2020-12-09T16:09:53Z"
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			resOrder, err := json.Marshal(tc.Order)
			assert.Nil(t, err)
			assert.JSONEq(t, tc.expectedJSON, string(resOrder), "result not equal")
		})
	}
}
