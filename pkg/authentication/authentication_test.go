package authentication

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
)

func TestJWTValidation(t *testing.T) {
	secretKey := "test"
	authenticator := Authenticator{
		secretKey: secretKey,
		tokenExp: time.Hour,
	}
	tokenUserId := uuid.New()

	testCases := []struct {
		testName    string
		expiresAt   time.Time
		secretKey   string
		expectedErr bool
	}{
		{
			testName:    "valid token",
			expiresAt:   time.Now().Add(time.Hour),
			secretKey:   secretKey,
			expectedErr: false,
		},
		{
			testName:    "expired token",
			expiresAt:   time.Now().Add(-time.Hour),
			secretKey:   secretKey,
			expectedErr: true,
		},
		{
			testName:    "invalid signature",
			expiresAt:   time.Now().Add(time.Hour),
			secretKey:   "invalid_secret_key",
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(tc.expiresAt),
				},
				UserID: tokenUserId,
				Login:  "login",
			})

			tokenStr, _ := token.SignedString([]byte(tc.secretKey))
			userId, err := authenticator.GetUserId(tokenStr)
			if !tc.expectedErr {
				assert.Nil(t, err, "unexpected error")
				assert.Equal(t, tokenUserId, *userId, "user ids don't match")
			} else {
				assert.Error(t, err, "sholud be an error")
			}
		})
	}
}

func TestJWTGeneration(t *testing.T) {
	secretKey := "test"
	authenticator := Authenticator{
		secretKey: 	  secretKey,
		tokenExp:     time.Hour,
	}
	inputId := uuid.New()
	login := "login"

	tokenStr, err := authenticator.MakeJWT(inputId, login)
	assert.Nil(t, err, "unexpected error")

	claims := &Claims{}
	_, err = jwt.ParseWithClaims(*tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	assert.Nil(t, err, "invalid jwt token")
}
