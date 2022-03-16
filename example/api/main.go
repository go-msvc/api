package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-msvc/api"
	"github.com/go-msvc/api/example/db"
	"github.com/go-msvc/api/example/model"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/stewelarend/logger"
)

var log = logger.New().WithLevel(logger.LevelDebug)

func main() {
	api.New(
		api.R{
			"/accounts": api.M{
				http.MethodGet: withToken(getAccounts),
			},
			"/users": api.M{
				http.MethodGet:  withToken(getUsers),
				http.MethodPost: withToken(addUser),
			},
			"/user/{id}": api.M{
				http.MethodGet: withToken(getUser),
				//				http.MethodPut:    updUser,
				//				http.MethodDelete: delUser,
			},
			"/auth/login": api.M{
				http.MethodPost: login, //open - no token required
			},
			"/auth/extend": api.M{
				http.MethodPost: withToken(extendSession), //token required in header
			},
			"/auth/logout": api.M{
				//				http.MethodGet: authLogout,
			},
		},
	).Run(":10000")
}

type ctxHandlerFunc func(ctx context.Context, httpRes http.ResponseWriter, httpReq *http.Request)

//withToken is wrapper arround a handler that requires a valid authentication token
func withToken(h ctxHandlerFunc) api.H {
	return func(httpRes http.ResponseWriter, httpReq *http.Request) {
		token := httpReq.Header.Get("X-Auth-Token")
		if token == "" {
			http.Error(httpRes, "missing header X-Auth-Token", http.StatusBadRequest)
			return
		}
		s, err := db.GetSession(token)
		if err != nil {
			log.Errorf("session error: %+v", err)
			http.Error(httpRes, fmt.Sprintf("session error: %v", err), http.StatusUnauthorized)
			return
		}
		log.Debugf("session: %+v", s)

		ctx := context.Background()
		ctx = context.WithValue(ctx, CtxSession{}, s)
		h(ctx, httpRes, httpReq)
	}
}

type CtxSession struct{}

func getUsers(ctx context.Context, httpRes http.ResponseWriter, httpReq *http.Request) {
	s, _ := ctx.Value(CtxSession{}).(model.Session)
	filter := map[string]interface{}{}
	if s.AccountID != 0 {
		filter["account_id"] = s.AccountID
	}
	users, err := db.GetUsers(filter, []string{"username"}, 10)
	if err != nil {
		http.Error(httpRes, err.Error(), http.StatusInternalServerError)
		return
	}

	httpRes.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(httpRes).Encode(users); err != nil {
		http.Error(httpRes, err.Error(), http.StatusInternalServerError)
		return
	}
}

func addUser(ctx context.Context, httpRes http.ResponseWriter, httpReq *http.Request) {
	//todo: check admin token, or account admin token

	if httpReq.Header.Get("Content-Type") != "application/json" {
		http.Error(httpRes, fmt.Sprintf("invalid Content-Type:\"%s\", expecting \"application/json\"", httpReq.Header.Get("Content-Type")), http.StatusBadRequest)
		return
	}
	var user model.User
	if err := json.NewDecoder(httpReq.Body).Decode(&user); err != nil {
		http.Error(httpRes, "cannot read JSON user data from body", http.StatusBadRequest)
		return
	}

	addedUser, err := db.AddUser(user)
	if err != nil {
		http.Error(httpRes, fmt.Sprintf("cannot add user: %+v", err), http.StatusBadRequest)
		return
	}
	httpRes.Header().Set("Content-Type", "application/json")
	json.NewEncoder(httpRes).Encode(addedUser)
}

func getUser(ctx context.Context, httpRes http.ResponseWriter, httpReq *http.Request) {
	vars := mux.Vars(httpReq)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(httpRes, "missing id", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(httpRes, fmt.Sprintf("non-integer id \"%s\"", idStr), http.StatusBadRequest)
		return
	}
	user, ok := db.GetUserByID(int(id))
	if !ok {
		http.Error(httpRes, "user not found", http.StatusNotFound)
		return
	}
	httpRes.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(httpRes).Encode(user); err != nil {
		http.Error(httpRes, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getAccounts(ctx context.Context, httpRes http.ResponseWriter, httpReq *http.Request) {
	accounts, err := db.GetAccounts()
	if err != nil {
		http.Error(httpRes, err.Error(), http.StatusInternalServerError)
		return
	}

	httpRes.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(httpRes).Encode(accounts); err != nil {
		http.Error(httpRes, err.Error(), http.StatusInternalServerError)
		return
	}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func login(httpRes http.ResponseWriter, httpReq *http.Request) {
	if httpReq.Header.Get("Content-Type") != "application/json" {
		http.Error(httpRes, fmt.Sprintf("invalid Content-Type:\"%s\", expecting \"application/json\"", httpReq.Header.Get("Content-Type")), http.StatusBadRequest)
		return
	}
	var req loginRequest
	if err := json.NewDecoder(httpReq.Body).Decode(&req); err != nil {
		http.Error(httpRes, "cannot read JSON data from body", http.StatusBadRequest)
		return
	}

	s, err := db.Login(req.Username, req.Password, time.Minute*5)
	if err != nil {
		http.Error(httpRes, err.Error(), http.StatusInternalServerError)
		return
	}
	httpRes.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(httpRes).Encode(s); err != nil {
		http.Error(httpRes, err.Error(), http.StatusInternalServerError)
		return
	}
}

type extendRequest struct {
	Minutes int `json:"minutes"`
}

func extendSession(ctx context.Context, httpRes http.ResponseWriter, httpReq *http.Request) {
	s, _ := ctx.Value(CtxSession{}).(model.Session)
	if httpReq.Header.Get("Content-Type") != "application/json" {
		http.Error(httpRes, fmt.Sprintf("invalid Content-Type:\"%s\", expecting \"application/json\"", httpReq.Header.Get("Content-Type")), http.StatusBadRequest)
		return
	}
	var req extendRequest
	if err := json.NewDecoder(httpReq.Body).Decode(&req); err != nil {
		http.Error(httpRes, "cannot read JSON data from body", http.StatusBadRequest)
		return
	}
	if req.Minutes <= 0 || req.Minutes > 1000 {
		http.Error(httpRes, fmt.Sprintf("invalid minutes:%d", req.Minutes), http.StatusBadRequest)
		return
	}
	err := db.ExtendSession(s.Token, time.Minute*time.Duration(req.Minutes))
	if err != nil {
		http.Error(httpRes, err.Error(), http.StatusNotFound)
		return
	}
	//success
}
