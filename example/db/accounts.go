package db

import (
	"github.com/go-msvc/api/example/model"
	"github.com/go-msvc/errors"
)

func GetAccounts() ([]model.Account, error) {
	stmt, err := db().Prepare(`select id,name from accounts`)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to prepare SQL query")
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query")
	}
	defer rows.Close()

	var accounts []model.Account
	for rows.Next() {
		var account model.Account
		err = rows.Scan(&account.ID, &account.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to scan account row")
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
} //GetAccounts()

func GetAccountByID(id int) (model.Account, bool) {
	stmt, err := db().Prepare(`select id,name from accounts where id=?`)
	if err != nil {
		log.Errorf("failed to prepare SQL query: %+v", err)
		return model.Account{}, false
	}
	defer stmt.Close()

	row := stmt.QueryRow(id)
	if row == nil {
		log.Errorf("failed to query: %+v", err)
		return model.Account{}, false
	}

	var account model.Account
	if err = row.Scan(&account.ID, &account.Name); err != nil {
		log.Errorf("failed to scan query result: %+v", err)
		return model.Account{}, false
	}
	return account, true
}
