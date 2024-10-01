package orderhelpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateOrderID(t *testing.T) {
	testCases := []struct {
		testName                 string
		OrderID                  string
		expectedValidationResult bool
	}{
		{
			testName:                 "valid odd order id",
			OrderID:                  "1388",
			expectedValidationResult: true,
		},
		{
			testName:                 "valid even order id",
			OrderID:                  "37218",
			expectedValidationResult: true,
		},
		{
			testName:                 "invalid odd order id",
			OrderID:                  "1323",
			expectedValidationResult: false,
		},
		{
			testName:                 "invalid even order id",
			OrderID:                  "12345",
			expectedValidationResult: false,
		},
		{
			testName:                 "invalid order id",
			OrderID:                  "1388a",
			expectedValidationResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			assert.Equal(t, tc.expectedValidationResult, ValidateOrderID(tc.OrderID), "validation error")
		})
	}
}
