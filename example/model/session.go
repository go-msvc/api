package model

type Session struct {
	Token       string  `json:"token"`
	AccountID   int     `json:"account_id,omitempty"`
	UserID      int     `json:"user_id,omitempty"`
	Username    string  `json:"username,omitempty"`
	TimeCreated SqlTime `json:"time_created"`
	TimeExpire  SqlTime `json:"time_expire"`
}
