package usermemstorage

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/user"
)

func TestInsertUser(t *testing.T) {
	existingUser := user.User{
		ID:           uuid.New(),
		Login:        "login_1",
		PasswordHash: []byte("testPass1"),
	}

	testCases := []struct {
		testName  string
		inputUser user.User
	}{
		{
			testName: "new user",
			inputUser: user.User{
				ID:           uuid.New(),
				Login:        "login_2",
				PasswordHash: []byte("testPass"),
			},
		},
		{
			testName: "existing user",
			inputUser: user.User{
				ID:           uuid.New(),
				Login:        existingUser.Login,
				PasswordHash: []byte("testPass2"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			storage := NewUserMemStorage()
			storage.users.Store(existingUser.Login, existingUser)

			storage.InsertUser(context.TODO(), &tc.inputUser, nil)
			userInDB, _ := storage.users.Load(tc.inputUser.Login)
			assert.Equal(t, tc.inputUser, userInDB, "users not equal")
		})
	}
}

func TestGetUser(t *testing.T) {
	existingUser := user.User{
		ID:           uuid.New(),
		Login:        "login_1",
		PasswordHash: []byte("testPass1"),
	}

	testCases := []struct {
		testName     string
		login        string
		expectedUser *user.User
	}{
		{
			testName:     "existing user",
			login:        "login_1",
			expectedUser: &existingUser,
		},
		{
			testName:     "new user",
			login:        "login_2",
			expectedUser: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			storage := NewUserMemStorage()
			storage.users.Store(existingUser.Login, existingUser)

			userInDB, err := storage.GetUser(context.TODO(), tc.login)
			if userInDB == nil {
				assert.False(t, false)
			}
			if tc.expectedUser != nil {
				assert.Equal(t, *tc.expectedUser, *userInDB, "users not equal")
			} else {
				assert.ErrorIs(t, err, exceptions.NewUserNotFoundError(), "exceptions don't match")
			}
		})
	}
}
