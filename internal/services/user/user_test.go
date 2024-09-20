package userservice

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/user"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/users"
	"github.com/ry461ch/loyalty_system/pkg/authentication"
)

func TestRegister(t *testing.T) {
	existingUser := user.User{
		ID:           uuid.New(),
		Login:        "login_1",
		PasswordHash: user.GeneratePasswordHash("test"),
	}

	testCases := []struct {
		testName             string
		inputUser            user.InputUser
		expectedSavingResult error
	}{
		{
			testName: "successfully saved",
			inputUser: user.InputUser{
				Login:    "login_2",
				Password: "testPass",
			},
			expectedSavingResult: nil,
		},
		{
			testName: "user already exists",
			inputUser: user.InputUser{
				Login:    existingUser.Login,
				Password: "testPass",
			},
			expectedSavingResult: exceptions.NewUserConflictError(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			secretKey := "test"
			storage := usermemstorage.NewUserMemStorage()
			authenticator := authentication.NewAuthenticator(secretKey, time.Hour)
			storage.InsertUser(context.TODO(), &existingUser, nil)
			userService := NewUserService(storage, authenticator)

			tokenStr, registerErr := userService.Register(context.TODO(), &tc.inputUser)
			if tc.expectedSavingResult == nil {
				claims := &authentication.Claims{}
				_, err := jwt.ParseWithClaims(*tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
					if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
					}
					return []byte(secretKey), nil
				})
				assert.Nil(t, err, "invalid jwt token")
			} else {
				assert.ErrorIs(t, registerErr, tc.expectedSavingResult, "unexpected error")
			}
		})
	}
}

func TestLogin(t *testing.T) {
	existingPassword := "test"
	existingUser := user.User{
		ID:           uuid.New(),
		Login:        "login_1",
		PasswordHash: user.GeneratePasswordHash(existingPassword),
	}

	testCases := []struct {
		testName             string
		inputUser            user.InputUser
		expectedSavingResult error
	}{
		{
			testName: "successfully authenticated",
			inputUser: user.InputUser{
				Login:    existingUser.Login,
				Password: existingPassword,
			},
			expectedSavingResult: nil,
		},
		{
			testName: "bad password",
			inputUser: user.InputUser{
				Login:    existingUser.Login,
				Password: "invalid_password",
			},
			expectedSavingResult: exceptions.NewUserAuthenticationError(),
		},
		{
			testName: "user doesn't exist",
			inputUser: user.InputUser{
				Login:    "login_2",
				Password: existingPassword,
			},
			expectedSavingResult: exceptions.NewUserAuthenticationError(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			secretKey := "test"
			storage := usermemstorage.NewUserMemStorage()
			authenticator := authentication.NewAuthenticator(secretKey, time.Hour)
			storage.InsertUser(context.TODO(), &existingUser, nil)
			userService := NewUserService(storage, authenticator)

			tokenStr, authErr := userService.Login(context.TODO(), &tc.inputUser)
			if tc.expectedSavingResult == nil {
				claims := &authentication.Claims{}
				_, err := jwt.ParseWithClaims(*tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
					if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
					}
					return []byte(secretKey), nil
				})
				assert.Nil(t, err, "invalid jwt token")
			} else {
				assert.ErrorIs(t, authErr, tc.expectedSavingResult, "unexpected error")
			}
		})
	}
}
