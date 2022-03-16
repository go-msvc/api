package db

import (
	"crypto/sha1"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-msvc/api/example/model"
	"github.com/go-msvc/errors"
	"github.com/google/uuid"
)

func Login(username string, password string, duration time.Duration) (model.Session, error) {
	user, ok := GetUserByUsername(username)
	if !ok {
		return model.Session{}, errors.Errorf("unknown username:\"%s\"", username)
	}
	passwordHash := hash(password)
	if user.PasswordHash != passwordHash {
		log.Errorf("password(%s)->\"%s\" != user:%+v", password, passwordHash, user)
		return model.Session{}, errors.Errorf("wrong password")
	}

	//generate session with token and store in table (replacing any existing token for this user)
	s := model.Session{
		Token:       uuid.New().String(),
		AccountID:   user.AccountID,
		UserID:      user.ID,
		Username:    user.Username,
		TimeCreated: model.SqlTime(time.Now()),
		TimeExpire:  model.SqlTime(time.Now().Add(duration)),
	}

	//delete existing sessions for this user
	if _, err := db().Exec(fmt.Sprintf("delete from sessions where user_id=%d", user.ID)); err != nil {
		return model.Session{}, errors.Wrapf(err, "failed to delete old sessions for user")
	}
	//store this new session for the user
	stmt, err := db().Prepare("insert into sessions set token=?,account_id=?,user_id=?,username=?,time_created=?,time_expire=?")
	if err != nil {
		return model.Session{}, errors.Wrapf(err, "failed to prepare statement")
	}
	defer stmt.Close()
	result, err := stmt.Exec(s.Token, s.AccountID, s.UserID, s.Username, s.TimeCreated.String(), s.TimeExpire.String())
	if err != nil {
		return model.Session{}, errors.Wrapf(err, "failed to store new session")
	}
	if i64, err := result.RowsAffected(); err != nil || i64 != 1 {
		return model.Session{}, errors.Errorf("insert session affected %d rows: %+v", i64, err)
	}
	//successful login, return session info
	return s, nil
}

func ExtendSession(token string, duration time.Duration) error {
	stmt, err := db().Prepare("update sessions set time_expire=? where token=? and time_expire>?")
	if err != nil {
		return errors.Wrapf(err, "failed to prepare sql")
	}
	defer stmt.Close()

	now := model.SqlTime(time.Now())
	exp := model.SqlTime(time.Now().Add(duration))
	result, err := stmt.Exec(exp.String(), token, now.String())
	if err != nil {
		return errors.Wrapf(err, "failed to extend session")
	}
	if i64, _ := result.RowsAffected(); i64 != 1 {
		return errors.Errorf("extend session affected %d rows", i64)
	}
	return nil
}

func GetSession(token string) (model.Session, error) {
	stmt, err := db().Prepare("select account_id,user_id,username,time_created,time_expire from sessions where token=?")
	if err != nil {
		return model.Session{}, errors.Wrapf(err, "failed to prepare sql")
	}
	defer stmt.Close()

	row := stmt.QueryRow(token)
	if row == nil {
		return model.Session{}, errors.Errorf("session.token(%s) not found", token)
	}
	log.Debugf("row: %+v", *row)
	s := model.Session{Token: token}
	if err := row.Scan(&s.AccountID, &s.UserID, &s.Username, &s.TimeCreated, &s.TimeExpire); err != nil {
		if err == sql.ErrNoRows {
			return model.Session{}, errors.Errorf("session.token(%s) not found", token)
		}
		return model.Session{}, errors.Wrapf(err, "failed to scan session info")
	}
	if time.Time(s.TimeExpire).Before(time.Now()) {
		return model.Session{}, errors.Errorf("session.token(%s) expired at %s", token, s.TimeExpire)
	}

	//extend the session automatically for 5 minutes from now
	desiredExpTime := time.Now().Add(time.Minute * 5)
	if time.Time(s.TimeExpire).Before(desiredExpTime) {
		if err := ExtendSession(token, time.Minute*5); err != nil {
			log.Errorf("failed to auto-extend session: %+v", err)
		} else {
			log.Debugf("session.token(%s) auto-extended to %s", token, desiredExpTime.Format("2006-01-02 15:04:05"))
			s.TimeExpire = model.SqlTime(desiredExpTime)
		}
	}
	return s, nil
}

func hash(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}
