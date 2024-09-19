package authhandlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/user"
	"github.com/ry461ch/loyalty_system/internal/services"
)

type AuthHandlers struct {
	userService services.UserService
}

func NewAuthHandlers(userService services.UserService) *AuthHandlers {
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
	if err != nil {
		switch {
		case errors.Is(err, exceptions.NewUserConflictError()):
			res.WriteHeader(http.StatusConflict)
			return
		default:
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	res.Header().Set("Authorization", *tokenStr)
	res.WriteHeader(http.StatusOK)
}

func (ah *AuthHandlers) Authenticate(res http.ResponseWriter, req *http.Request) {
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

	tokenStr, err := ah.userService.Authenticate(req.Context(), &inputUser)
	if err != nil {
		switch {
		case errors.Is(err, exceptions.NewUserAuthenticationError()):
			res.WriteHeader(http.StatusUnauthorized)
			return
		default:
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	res.Header().Set("Authorization", *tokenStr)
	res.WriteHeader(http.StatusOK)
}
