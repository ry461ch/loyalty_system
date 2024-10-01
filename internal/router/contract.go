package router

import "net/http"

type AuthHandlers interface {
	Register(res http.ResponseWriter, req *http.Request)
	Login(res http.ResponseWriter, req *http.Request)
}

type OrderHandlers interface {
	PostOrder(res http.ResponseWriter, req *http.Request)
	GetOrders(res http.ResponseWriter, req *http.Request)
}

type MoneyHandlers interface {
	PostWithdrawal(res http.ResponseWriter, req *http.Request)
	GetWithdrawals(res http.ResponseWriter, req *http.Request)
	GetBalance(res http.ResponseWriter, req *http.Request)
}
