package userservice

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/user"
	"github.com/ry461ch/loyalty_system/internal/storage"
	"github.com/ry461ch/loyalty_system/pkg/authentication"
)

type UserService struct {
	userStorage   storage.UserStorage
	authenticator *authentication.Authenticator
}

func NewUserService(userStorage storage.UserStorage, authenticator *authentication.Authenticator) *UserService {
	return &UserService{
		userStorage:   userStorage,
		authenticator: authenticator,
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
	return us.authenticator.MakeJWT(newUser.Id, newUser.Login)
}

func (us *UserService) Login(ctx context.Context, inputUser *user.InputUser) (*string, error) {
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

	return us.authenticator.MakeJWT(userInDB.Id, userInDB.Login)
}
