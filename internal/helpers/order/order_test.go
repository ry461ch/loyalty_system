package orderhelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateOrderId(t *testing.T) {
	testCases := []struct {
		testName                 string
		OrderId                  string
		expectedValidationResult bool
	}{
		{
			testName:                 "valid odd order id",
			OrderId:                  "1388",
			expectedValidationResult: true,
		},
		{
			testName:                 "valid even order id",
			OrderId:                  "37218",
			expectedValidationResult: true,
		},
		{
			testName:                 "invalid odd order id",
			OrderId:                  "1323",
			expectedValidationResult: false,
		},
		{
			testName:                 "invalid even order id",
			OrderId:                  "12345",
			expectedValidationResult: false,
		},
		{
			testName:                 "invalid order id",
			OrderId:                  "1388a",
			expectedValidationResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			assert.Equal(t, tc.expectedValidationResult, ValidateOrderId(tc.OrderId), "validation error")
		})
	}
}
