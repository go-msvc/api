package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-msvc/api/example/model"
	"github.com/go-msvc/errors"
	"github.com/stewelarend/logger"
)

var log = logger.New().WithLevel(logger.LevelDebug)

func GetUsers(filter map[string]interface{}, order []string, limit int) ([]model.User, error) {
	args := []interface{}{}
	q := "select id,account_id,username from users"
	if accountID, ok := filter["account_id"].(int); ok {
		q += " where account_id=?"
		args = append(args, accountID)
	}
	if len(order) > 0 {
		q += " order by " + strings.Join(order, ",")
	}
	if limit <= 0 {
		limit = 10
	}
	q += fmt.Sprintf(" limit %d", limit)

	var stmt *sql.Stmt
	var err error
	if stmt, err = db().Prepare(q); err != nil {
		return nil, errors.Wrapf(err, "failed to prepare SQL query: %s", q)
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query")
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		err = rows.Scan(&user.ID, &user.AccountID, &user.Username)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to scan user row")
		}
		users = append(users, user)
	}
	return users, nil
} //GetUsers()

func GetUserByID(id int) (model.User, bool) {
	stmt, err := db().Prepare(`select id,account_id,username,time_created,active from users where id=?`)
	if err != nil {
		log.Errorf("failed to prepare SQL query: %+v", err)
		return model.User{}, false
	}
	defer stmt.Close()

	row := stmt.QueryRow(id)
	if row == nil {
		log.Errorf("failed to query: %+v", err)
		return model.User{}, false
	}

	var user model.User
	if err = row.Scan(&user.ID, &user.AccountID, &user.Username, &user.TimeCreated, &user.Active); err != nil {
		log.Errorf("failed to scan query result: %+v", err)
		return model.User{}, false
	}
	return user, true
}

func GetUserByUsername(username string) (model.User, bool) {
	stmt, err := db().Prepare(`select id,account_id,username,password_hash from users where username=?`)
	if err != nil {
		log.Errorf("failed to prepare SQL query: %+v", err)
		return model.User{}, false
	}
	defer stmt.Close()

	row := stmt.QueryRow(username)
	if row == nil {
		log.Errorf("failed to query: %+v", err)
		return model.User{}, false
	}

	var user model.User
	if err = row.Scan(&user.ID, &user.AccountID, &user.Username, &user.PasswordHash); err != nil {
		log.Errorf("failed to scan query result: %+v", err)
		return model.User{}, false
	}
	return user, true
}

func AddUser(user model.User) (model.User, error) {
	if user.ID != 0 {
		return model.User{}, errors.Errorf("id may not be specified for new user")
	}
	if user.Username == "" || !model.ValidUsername(user.Username) {
		return model.User{}, errors.Errorf("invalid username:\"%s\" (expecting lowercase and digits only)", user.Username)
	}
	if user.Password == "" {
		return model.User{}, errors.Errorf("missing password")
	}
	if user.AccountID <= 0 {
		return model.User{}, errors.Errorf("missing/invalid account_id:%d", user.AccountID)
	}
	_, ok := GetAccountByID(user.AccountID)
	if !ok {
		return model.User{}, errors.Errorf("unknown account_id:%d", user.AccountID)
	}

	stmt, err := db().Prepare(`insert into users set account_id=?,username=?,password_hash=?`)
	if err != nil {
		return model.User{}, errors.Wrapf(err, "failed to prepare SQL query")
	}
	defer stmt.Close()

	result, err := stmt.Exec(user.AccountID, user.Username, hash(user.Password))
	if err != nil {
		return model.User{}, errors.Wrapf(err, "failed to insert user into db")
	}
	i64, err := result.LastInsertId()
	if err != nil {
		return model.User{}, errors.Wrapf(err, "failed to insert user into db")
	}
	//define user ID in the response, but remove the specified password
	user.ID = int(i64)
	user.Password = ""
	user.TimeCreated = model.SqlTime(time.Now())
	return user, nil
}
