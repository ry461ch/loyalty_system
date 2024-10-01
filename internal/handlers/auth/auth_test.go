package authhandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/loyalty_system/internal/models/user"
	"github.com/ry461ch/loyalty_system/internal/services/user"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/users"
	"github.com/ry461ch/loyalty_system/pkg/authentication"
	"github.com/ry461ch/loyalty_system/pkg/logging"
)

func mockRouter(authHandlers *AuthHandlers) chi.Router {
	router := chi.NewRouter()
	router.Post("/api/user/register", authHandlers.Register)
	router.Post("/api/user/login", authHandlers.Login)
	return router
}

func TestRegister(t *testing.T) {
	logging.Initialize("INFO")
	existingUser := user.InputUser{
		Login:    "test",
		Password: "test",
	}

	secretKey := "test_secret_key"
	storage := usermemstorage.NewUserMemStorage()
	authenticator := authentication.NewAuthenticator(secretKey, time.Hour)
	userService := userservice.NewUserService(storage, authenticator)
	handlers := NewAuthHandlers(userService)
	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()
	client := resty.New()

	testCases := []struct {
		testName     string
		requestBody  *user.InputUser
		expectedCode int
	}{
		{
			testName: "successful registration",
			requestBody: &user.InputUser{
				Login:    "new",
				Password: "test",
			},
			expectedCode: http.StatusOK,
		},
		{
			testName: "user already exists",
			requestBody: &user.InputUser{
				Login:    existingUser.Login,
				Password: "test",
			},
			expectedCode: http.StatusConflict,
		},
		{
			testName:     "bad request",
			requestBody:  nil,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			userService.Register(context.TODO(), &existingUser)

			var req []byte
			if tc.requestBody != nil {
				req, _ = json.Marshal(*tc.requestBody)
			} else {
				req, _ = json.Marshal(map[string]string{})
			}

			resp, _ := client.R().
				SetHeader("Content-Type", "application/json").
				SetBody(req).
				Execute(http.MethodPost, srv.URL+"/api/user/register")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestLogin(t *testing.T) {
	logging.Initialize("INFO")
	existingUser := user.InputUser{
		Login:    "test",
		Password: "test",
	}

	secretKey := "test_secret_key"
	storage := usermemstorage.NewUserMemStorage()
	authenticator := authentication.NewAuthenticator(secretKey, time.Hour)
	userService := userservice.NewUserService(storage, authenticator)
	handlers := NewAuthHandlers(userService)
	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()
	client := resty.New()

	testCases := []struct {
		testName     string
		requestBody  *user.InputUser
		expectedCode int
	}{
		{
			testName: "successful authentication",
			requestBody: &user.InputUser{
				Login:    existingUser.Login,
				Password: existingUser.Password,
			},
			expectedCode: http.StatusOK,
		},
		{
			testName: "user not exist",
			requestBody: &user.InputUser{
				Login:    "invalid_login",
				Password: "test",
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			testName: "invalid password",
			requestBody: &user.InputUser{
				Login:    existingUser.Login,
				Password: "invalid_password",
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			testName:     "bad request",
			requestBody:  nil,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			userService.Register(context.TODO(), &existingUser)

			var req []byte
			if tc.requestBody != nil {
				req, _ = json.Marshal(*tc.requestBody)
			} else {
				req, _ = json.Marshal(map[string]string{})
			}

			resp, _ := client.R().
				SetHeader("Content-Type", "application/json").
				SetBody(req).
				Execute(http.MethodPost, srv.URL+"/api/user/login")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")
		})
	}
}
