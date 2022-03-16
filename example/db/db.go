package db

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

var dbConn *sql.DB

func db() *sql.DB {
	if dbConn != nil {
		return dbConn
	}
	var err error
	dbConn, err = sql.Open("mysql", "admin:admin@tcp(127.0.0.1:3306)/example")
	if err != nil {
		panic(err)
	}
	return dbConn
}
