package authentication

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Authenticator struct {
	secretKey string
	tokenExp time.Duration
}

type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID
	Login  string
}

func NewAuthenticator(secretKey string, tokenExp time.Duration) *Authenticator {
	return &Authenticator{
		secretKey: secretKey,
		tokenExp: tokenExp,
	}
}

func (a *Authenticator) MakeJWT(id uuid.UUID, login string) (*string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.tokenExp)),
		},
		UserID: id,
		Login:  login,
	})

	tokenString, err := token.SignedString([]byte(a.secretKey))
	if err != nil {
		return nil, err
	}

	return &tokenString, nil
}


func (a *Authenticator) GetUserId(tokenStr string) (*uuid.UUID, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("authentication error")
		}
		return []byte(a.secretKey), nil
	})
	if err != nil {
		return nil, errors.New("authentication error")
	}

	if !token.Valid {
		return nil, errors.New("authentication error")
	}

	return &claims.UserID, nil
}
