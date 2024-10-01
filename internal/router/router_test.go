package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/loyalty_system/pkg/authentication"
	"github.com/ry461ch/loyalty_system/pkg/logging"
)

type MockAuthHandlers struct {
	pathTimesCalled map[string]int64
}

func NewMockAuthHandlers() *MockAuthHandlers {
	return &MockAuthHandlers{pathTimesCalled: map[string]int64{}}
}

func (mah *MockAuthHandlers) Register(res http.ResponseWriter, req *http.Request) {
	mah.pathTimesCalled["register"] += 1
	res.WriteHeader(http.StatusOK)
}

func (mah *MockAuthHandlers) Login(res http.ResponseWriter, req *http.Request) {
	mah.pathTimesCalled["login"] += 1
	res.WriteHeader(http.StatusOK)
}

type MockOrderHandlers struct {
	pathTimesCalled map[string]int64
}

func NewMockOrderHandlers() *MockOrderHandlers {
	return &MockOrderHandlers{pathTimesCalled: map[string]int64{}}
}

func (moh *MockOrderHandlers) PostOrder(res http.ResponseWriter, req *http.Request) {
	moh.pathTimesCalled["post_order"] += 1
	res.WriteHeader(http.StatusOK)
}

func (moh *MockOrderHandlers) GetOrders(res http.ResponseWriter, req *http.Request) {
	moh.pathTimesCalled["get_orders"] += 1
	res.WriteHeader(http.StatusOK)
}

type MockMoneyHandlers struct {
	pathTimesCalled map[string]int64
}

func NewMockMoneyHandlers() *MockMoneyHandlers {
	return &MockMoneyHandlers{pathTimesCalled: map[string]int64{}}
}

func (mmh *MockMoneyHandlers) PostWithdrawal(res http.ResponseWriter, req *http.Request) {
	mmh.pathTimesCalled["post_withdrawal"] += 1
	res.WriteHeader(http.StatusOK)
}

func (mmh *MockMoneyHandlers) GetWithdrawals(res http.ResponseWriter, req *http.Request) {
	mmh.pathTimesCalled["get_withdrawals"] += 1
	res.WriteHeader(http.StatusOK)
}

func (mmh *MockMoneyHandlers) GetBalance(res http.ResponseWriter, req *http.Request) {
	mmh.pathTimesCalled["get_balance"] += 1
	res.WriteHeader(http.StatusOK)
}

func TestRouter(t *testing.T) {
	jsonContentType := "application/json"
	plainContentType := "text/plain"

	secretKey := "test_secret_key"
	authenticator := authentication.NewAuthenticator(secretKey, time.Hour)
	validTokenStr, _ := authenticator.MakeJWT(uuid.New(), "login")
	fakeAuthenticator := authentication.NewAuthenticator("fake_token", time.Hour)
	invalidTokenStr, _ := fakeAuthenticator.MakeJWT(uuid.New(), "login")
	logging.Initialize("INFO")

	client := resty.New()

	authHandlers := NewMockAuthHandlers()
	orderHandlers := NewMockOrderHandlers()
	moneyHandlers := NewMockMoneyHandlers()
	router := NewRouter(authHandlers, moneyHandlers, orderHandlers, authenticator)
	srv := httptest.NewServer(router)
	defer srv.Close()

	testCases := []struct {
		testName                string
		method                  string
		requestPath             string
		requestContentType      string
		requestAuthHeader       string
		expectedCode            int
		expectedPathTimesCalled map[string]int64
	}{
		{
			testName:                "valid registration",
			method:                  http.MethodPost,
			requestPath:             "/api/user/register",
			requestContentType:      jsonContentType,
			requestAuthHeader:       *invalidTokenStr,
			expectedCode:            http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"register": 1},
		},
		{
			testName:                "invalid registration method",
			method:                  http.MethodGet,
			requestPath:             "/api/user/register",
			requestContentType:      jsonContentType,
			requestAuthHeader:       *invalidTokenStr,
			expectedCode:            http.StatusMethodNotAllowed,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "invalid registration content type",
			method:                  http.MethodPost,
			requestPath:             "/api/user/register",
			requestContentType:      plainContentType,
			requestAuthHeader:       *invalidTokenStr,
			expectedCode:            http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "valid login",
			method:                  http.MethodPost,
			requestPath:             "/api/user/login",
			requestContentType:      jsonContentType,
			requestAuthHeader:       *invalidTokenStr,
			expectedCode:            http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"login": 1},
		},
		{
			testName:                "invalid login method",
			method:                  http.MethodGet,
			requestPath:             "/api/user/login",
			requestContentType:      jsonContentType,
			requestAuthHeader:       *invalidTokenStr,
			expectedCode:            http.StatusMethodNotAllowed,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "invalid login content type",
			method:                  http.MethodPost,
			requestPath:             "/api/user/login",
			requestContentType:      plainContentType,
			requestAuthHeader:       *invalidTokenStr,
			expectedCode:            http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "valid get orders",
			method:                  http.MethodGet,
			requestPath:             "/api/user/orders",
			requestContentType:      plainContentType,
			requestAuthHeader:       *validTokenStr,
			expectedCode:            http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"get_orders": 1},
		},
		{
			testName:                "invalid get orders content type",
			method:                  http.MethodGet,
			requestPath:             "/api/user/orders",
			requestContentType:      jsonContentType,
			requestAuthHeader:       *validTokenStr,
			expectedCode:            http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "invalid get orders token validation",
			method:                  http.MethodGet,
			requestPath:             "/api/user/orders",
			requestContentType:      plainContentType,
			requestAuthHeader:       *invalidTokenStr,
			expectedCode:            http.StatusUnauthorized,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "valid post order",
			method:                  http.MethodPost,
			requestPath:             "/api/user/orders",
			requestContentType:      plainContentType,
			requestAuthHeader:       *validTokenStr,
			expectedCode:            http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"post_order": 1},
		},
		{
			testName:                "invalid post order content type",
			method:                  http.MethodPost,
			requestPath:             "/api/user/orders",
			requestContentType:      jsonContentType,
			requestAuthHeader:       *validTokenStr,
			expectedCode:            http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "invalid post order token validation",
			method:                  http.MethodPost,
			requestPath:             "/api/user/orders",
			requestContentType:      plainContentType,
			requestAuthHeader:       *invalidTokenStr,
			expectedCode:            http.StatusUnauthorized,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "valid get balance",
			method:                  http.MethodGet,
			requestPath:             "/api/user/balance",
			requestContentType:      plainContentType,
			requestAuthHeader:       *validTokenStr,
			expectedCode:            http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"get_balance": 1},
		},
		{
			testName:                "invalid get balance method",
			method:                  http.MethodPost,
			requestPath:             "/api/user/balance",
			requestContentType:      plainContentType,
			requestAuthHeader:       *validTokenStr,
			expectedCode:            http.StatusMethodNotAllowed,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "invalid get balance content type",
			method:                  http.MethodGet,
			requestPath:             "/api/user/balance",
			requestContentType:      jsonContentType,
			requestAuthHeader:       *validTokenStr,
			expectedCode:            http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "invalid get balance token validation",
			method:                  http.MethodGet,
			requestPath:             "/api/user/balance",
			requestContentType:      plainContentType,
			requestAuthHeader:       *invalidTokenStr,
			expectedCode:            http.StatusUnauthorized,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "valid post withdrawal",
			method:                  http.MethodPost,
			requestPath:             "/api/user/balance/withdraw",
			requestContentType:      jsonContentType,
			requestAuthHeader:       *validTokenStr,
			expectedCode:            http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"post_withdrawal": 1},
		},
		{
			testName:                "invalid post withdrawal method",
			method:                  http.MethodGet,
			requestPath:             "/api/user/balance/withdraw",
			requestContentType:      jsonContentType,
			requestAuthHeader:       *validTokenStr,
			expectedCode:            http.StatusMethodNotAllowed,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "invalid post withdrawal content type",
			method:                  http.MethodPost,
			requestPath:             "/api/user/balance/withdraw",
			requestContentType:      plainContentType,
			requestAuthHeader:       *validTokenStr,
			expectedCode:            http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "invalid post withdrawal token validation",
			method:                  http.MethodPost,
			requestPath:             "/api/user/balance/withdraw",
			requestContentType:      jsonContentType,
			requestAuthHeader:       *invalidTokenStr,
			expectedCode:            http.StatusUnauthorized,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "valid get withdrawals",
			method:                  http.MethodGet,
			requestPath:             "/api/user/withdrawals",
			requestContentType:      jsonContentType,
			requestAuthHeader:       *validTokenStr,
			expectedCode:            http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"get_withdrawals": 1},
		},
		{
			testName:                "invalid get withdrawals method",
			method:                  http.MethodPost,
			requestPath:             "/api/user/withdrawals",
			requestContentType:      jsonContentType,
			requestAuthHeader:       *validTokenStr,
			expectedCode:            http.StatusMethodNotAllowed,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "invalid get withdrawals content type",
			method:                  http.MethodGet,
			requestPath:             "/api/user/withdrawals",
			requestContentType:      plainContentType,
			requestAuthHeader:       *validTokenStr,
			expectedCode:            http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "invalid get withdrawals token validation",
			method:                  http.MethodGet,
			requestPath:             "/api/user/withdrawals",
			requestContentType:      jsonContentType,
			requestAuthHeader:       *invalidTokenStr,
			expectedCode:            http.StatusUnauthorized,
			expectedPathTimesCalled: map[string]int64{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			resp, err := client.R().
				SetHeader("Content-Type", tc.requestContentType).
				SetHeader("Authorization", tc.requestAuthHeader).
				Execute(tc.method, srv.URL+tc.requestPath)
			assert.Nil(t, err, "Server returned 500")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "statuses not equal")
			timesCalled := len(authHandlers.pathTimesCalled) + len(moneyHandlers.pathTimesCalled) + len(orderHandlers.pathTimesCalled)
			assert.Equal(t, len(tc.expectedPathTimesCalled), timesCalled, "handlers time called not equal")

			pathTimesCalled := authHandlers.pathTimesCalled
			for key, val := range moneyHandlers.pathTimesCalled {
				pathTimesCalled[key] = val
			}
			for key, val := range orderHandlers.pathTimesCalled {
				pathTimesCalled[key] = val
			}

			for key, val := range pathTimesCalled {
				assert.Contains(t, tc.expectedPathTimesCalled, key, "invalid path was called")
				assert.Equal(t, tc.expectedPathTimesCalled[key], val, "necessary path was not called")
			}

			authHandlers.pathTimesCalled = map[string]int64{}
			moneyHandlers.pathTimesCalled = map[string]int64{}
			orderHandlers.pathTimesCalled = map[string]int64{}
		})
	}
}
