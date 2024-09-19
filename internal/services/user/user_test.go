package userservice

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/loyalty_system/internal/config"
	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/user"
	"github.com/ry461ch/loyalty_system/internal/storage/memory"
)

func TestRegister(t *testing.T) {
	existingUser := user.User{
		Id:           uuid.New(),
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
			storage := memstorage.NewMemStorage()
			cfg := config.Config{
				JWTSecretKey: "test",
				TokenExp:     time.Hour,
			}
			storage.InsertUser(context.TODO(), &existingUser, nil)
			userService := NewUserService(storage, &cfg)

			tokenStr, registerErr := userService.Register(context.TODO(), &tc.inputUser)
			if tc.expectedSavingResult == nil {
				claims := &user.Claims{}
				_, err := jwt.ParseWithClaims(*tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
					if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
					}
					return []byte(cfg.JWTSecretKey), nil
				})
				assert.Nil(t, err, "invalid jwt token")
			} else {
				assert.ErrorIs(t, registerErr, tc.expectedSavingResult, "unexpected error")
			}
		})
	}
}

func TestAuthenticate(t *testing.T) {
	existingPassword := "test"
	existingUser := user.User{
		Id:           uuid.New(),
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
			expectedSavingResult: exceptions.NewUserNotFoundError(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			storage := memstorage.NewMemStorage()
			cfg := config.Config{
				JWTSecretKey: "test",
				TokenExp:     time.Hour,
			}
			storage.InsertUser(context.TODO(), &existingUser, nil)
			userService := NewUserService(storage, &cfg)

			tokenStr, authErr := userService.Authenticate(context.TODO(), &tc.inputUser)
			if tc.expectedSavingResult == nil {
				claims := &user.Claims{}
				_, err := jwt.ParseWithClaims(*tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
					if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
					}
					return []byte(cfg.JWTSecretKey), nil
				})
				assert.Nil(t, err, "invalid jwt token")
			} else {
				assert.ErrorIs(t, authErr, tc.expectedSavingResult, "unexpected error")
			}
		})
	}
}

func TestJWTValidation(t *testing.T) {
	storage := memstorage.NewMemStorage()
	cfg := config.Config{
		JWTSecretKey: "test",
	}
	userService := NewUserService(storage, &cfg)
	tokenUserId := uuid.New()

	testCases := []struct {
		testName    string
		expiresAt   time.Time
		secretKey   string
		expectedErr error
	}{
		{
			testName:    "valid token",
			expiresAt:   time.Now().Add(time.Hour),
			secretKey:   cfg.JWTSecretKey,
			expectedErr: nil,
		},
		{
			testName:    "expired token",
			expiresAt:   time.Now().Add(-time.Hour),
			secretKey:   cfg.JWTSecretKey,
			expectedErr: exceptions.NewUserAuthenticationError(),
		},
		{
			testName:    "invalid signature",
			expiresAt:   time.Now().Add(time.Hour),
			secretKey:   "invalid_secret_key",
			expectedErr: exceptions.NewUserAuthenticationError(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, user.Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(tc.expiresAt),
				},
				UserID: tokenUserId,
				Login:  "login",
			})

			tokenStr, _ := token.SignedString([]byte(tc.secretKey))
			userId, err := userService.GetUserId(tokenStr)
			if tc.expectedErr == nil {
				assert.Nil(t, err, "unexpected error")
				assert.Equal(t, tokenUserId, *userId, "user ids don't match")
			} else {
				assert.ErrorIs(t, err, tc.expectedErr, "errors don't match")
			}
		})
	}
}

func TestJWTGeneration(t *testing.T) {
	storage := memstorage.NewMemStorage()
	cfg := config.Config{
		JWTSecretKey: "test",
		TokenExp:     time.Hour,
	}
	userService := NewUserService(storage, &cfg)
	inputUser := user.User{
		Id:           uuid.New(),
		Login:        "test",
		PasswordHash: []byte("pass_hash"),
	}
	tokenStr, err := userService.makeJWT(&inputUser)
	assert.Nil(t, err, "unexpected error")

	claims := &user.Claims{}
	_, err = jwt.ParseWithClaims(*tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(cfg.JWTSecretKey), nil
	})
	assert.Nil(t, err, "invalid jwt token")
}
