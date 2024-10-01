package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckPassword(t *testing.T) {
	inputUser := InputUser{
		Login:    "testLogin",
		Password: "testPassword",
	}
	user := New(inputUser)

	testCases := []struct {
		testName            string
		password            string
		expectedCheckResult bool
	}{
		{
			testName:            "same password",
			password:            inputUser.Password,
			expectedCheckResult: true,
		},
		{
			testName:            "invalid password",
			password:            "invalidPassword",
			expectedCheckResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			assert.Equal(t, tc.expectedCheckResult, CheckPassword(user.PasswordHash, tc.password))
		})
	}
}
