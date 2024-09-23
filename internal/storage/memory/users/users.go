package usermemstorage

import (
	"context"
	"sync"

	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/user"
	"github.com/ry461ch/loyalty_system/internal/storage/transaction"
)

type UserMemStorage struct {
	users sync.Map // map[string]user.User
}

func NewUserMemStorage() *UserMemStorage {
	return &UserMemStorage{}
}

func (ums *UserMemStorage) InsertUser(ctx context.Context, inputUser *user.User, trx *transaction.Trx) error {
	ums.users.Store(inputUser.Login, *inputUser)
	return nil
}

func (ums *UserMemStorage) GetUser(ctx context.Context, login string) (*user.User, error) {
	val, ok := ums.users.Load(login)
	if !ok {
		return nil, exceptions.ErrUserNotFound
	}
	userInDB := val.(user.User)
	return &userInDB, nil
}

func (*UserMemStorage) BeginTx(ctx context.Context) (*transaction.Trx, error) {
	return transaction.BeginTx(ctx, nil)
}
