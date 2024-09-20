package withdrawal

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshal(t *testing.T) {
	testCases := []struct {
		testName           string
		inputWithdrawal    string
		expectedWithdrawal Withdrawal
	}{
		{
			testName: "new withdrawal",
			inputWithdrawal: `{
				"order": "1321",
				"sum": 500
			}`,
			expectedWithdrawal: Withdrawal{
				OrderId: "1321",
				Sum:     500,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			var withdrawal Withdrawal
			err := json.Unmarshal([]byte(tc.inputWithdrawal), &withdrawal)
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedWithdrawal.OrderId, withdrawal.OrderId, "order id not equal")
			assert.Equal(t, tc.expectedWithdrawal.Sum, withdrawal.Sum, "Sum not equal")
		})
	}
}
