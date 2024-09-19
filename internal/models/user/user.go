package user

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
)

type InputUser struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (u *InputUser) UnmarshalJSON(data []byte) error {
	type UserAlias InputUser

	aliasValue := &struct {
		*UserAlias
	}{
		UserAlias: (*UserAlias)(u),
	}

	if err := json.Unmarshal(data, aliasValue); err != nil {
		return err
	}

	if aliasValue.Login == "" || aliasValue.Password == "" {
		return exceptions.NewUserBadFormatError()
	}

	return nil
}

type User struct {
	Id           uuid.UUID `json:"uid"`
	Login        string    `json:"login"`
	PasswordHash []byte    `json:"-"`
}

type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID
	Login  string
}

func New(inputUser InputUser) *User {
	return &User{
		Id:           uuid.New(),
		Login:        inputUser.Login,
		PasswordHash: GeneratePasswordHash(inputUser.Password),
	}
}

func GeneratePasswordHash(password string) []byte {
	h := sha256.New()
	h.Write([]byte(password))
	hash := h.Sum(nil)
	return hash
}

func CheckPassword(passHash []byte, password string) bool {
	return bytes.Equal(passHash, GeneratePasswordHash(password))
}
