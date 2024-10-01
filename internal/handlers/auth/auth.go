package authhandlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/user"
	"github.com/ry461ch/loyalty_system/pkg/logging"
)

type AuthHandlers struct {
	userService UserService
}

func NewAuthHandlers(userService UserService) *AuthHandlers {
	return &AuthHandlers{
		userService: userService,
	}
}

func (ah *AuthHandlers) Register(res http.ResponseWriter, req *http.Request) {
	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	var inputUser user.InputUser
	err = json.Unmarshal(reqBody, &inputUser)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	tokenStr, err := ah.userService.Register(req.Context(), &inputUser)
	if err == nil {
		res.Header().Set("Authorization", *tokenStr)
		res.WriteHeader(http.StatusOK)
		return
	}

	switch {
	case errors.Is(err, exceptions.ErrUserConflict):
		res.WriteHeader(http.StatusConflict)
	default:
		logging.Logger.Errorf("Register: internal error: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
	}
}

func (ah *AuthHandlers) Login(res http.ResponseWriter, req *http.Request) {
	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	var inputUser user.InputUser
	err = json.Unmarshal(reqBody, &inputUser)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	tokenStr, err := ah.userService.Login(req.Context(), &inputUser)

	if err == nil {
		res.Header().Set("Authorization", *tokenStr)
		res.WriteHeader(http.StatusOK)
		return
	}

	switch {
	case errors.Is(err, exceptions.ErrUserAuthentication):
		res.WriteHeader(http.StatusUnauthorized)
	default:
		logging.Logger.Errorf("Login: internal error: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
	}
}
