package userservice

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/config"
	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/user"
	"github.com/ry461ch/loyalty_system/internal/storage"
)

type UserService struct {
	userStorage storage.UserStorage
	cfg         *config.Config
}

func NewUserService(userStorage storage.UserStorage, cfg *config.Config) *UserService {
	return &UserService{
		userStorage: userStorage,
		cfg:         cfg,
	}
}

func (us *UserService) Register(ctx context.Context, inputUser *user.InputUser) (*string, error) {
	registeredUser, err := us.userStorage.GetUser(ctx, inputUser.Login)
	if registeredUser != nil {
		return nil, exceptions.NewUserConflictError()
	}
	if !errors.Is(err, exceptions.NewUserNotFoundError()) {
		return nil, err
	}

	newUser := user.User{
		Login:        inputUser.Login,
		PasswordHash: user.GeneratePasswordHash(inputUser.Password),
		Id:           uuid.New(),
	}

	tx, err := us.userStorage.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	err = us.userStorage.InsertUser(ctx, &newUser, tx)
	if err != nil {
		return nil, err
	}
	return us.makeJWT(&newUser)
}

func (us *UserService) Authenticate(ctx context.Context, inputUser *user.InputUser) (*string, error) {
	userInDB, err := us.userStorage.GetUser(ctx, inputUser.Login)
	if err != nil {
		if errors.Is(err, exceptions.NewUserNotFoundError()) {
			return nil, exceptions.NewUserAuthenticationError()
		}
		return nil, err
	}

	if !user.CheckPassword(userInDB.PasswordHash, inputUser.Password) {
		return nil, exceptions.NewUserAuthenticationError()
	}

	return us.makeJWT(userInDB)
}

func (us *UserService) GetUserId(tokenStr string) (*uuid.UUID, error) {
	claims := &user.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, exceptions.NewUserAuthenticationError()
		}
		return []byte(us.cfg.JWTSecretKey), nil
	})
	if err != nil {
		return nil, exceptions.NewUserAuthenticationError()
	}

	if !token.Valid {
		return nil, exceptions.NewUserAuthenticationError()
	}

	return &claims.UserID, nil
}

func (us *UserService) makeJWT(requestedUser *user.User) (*string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, user.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(us.cfg.TokenExp)),
		},
		UserID: requestedUser.Id,
		Login:  requestedUser.Login,
	})

	tokenString, err := token.SignedString([]byte(us.cfg.JWTSecretKey))
	if err != nil {
		return nil, err
	}

	return &tokenString, nil
}
