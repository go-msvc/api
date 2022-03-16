package model

import "regexp"

type User struct {
	ID           int     `json:"id"`
	AccountID    int     `json:"account_id"`
	Username     string  `json:"username"`
	PasswordHash string  `json:"password_hash,omitempty"`
	Password     string  `json:"password,omitempty"` //specify only when adding a new user
	Email        string  `json:"email,omitempty"`
	Phone        string  `json:"phone,omitempty"`
	TimeCreated  SqlTime `json:"time_created,omitempty"`
	Active       bool    `json:"active,omitempty"`
}

const usernamePattern = `[a-z][a-z0-9]*`

var usernameRegex = regexp.MustCompile("^" + usernamePattern + "$")

func ValidUsername(s string) bool {
	return usernameRegex.MatchString(s)
}
