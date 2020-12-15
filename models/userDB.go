package models

import (
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
)

// UserDB comment
type UserDB struct {
	*sql.DB
}

// New comment
func (db UserDB) New(username, pwhash string) (userid int, duplicate bool, err error) {
	q := `INSERT INTO users (name, pwhash, created) VALUES (?, ?, NOW())`
	r, err := db.Exec(q, username, pwhash)
	if err != nil {
		// Error 1062: Duplicate entry 'xyz' for key 'users.name'
		if err.(*mysql.MySQLError).Number == 1062 {
			return 0, true, nil
		}
		return
	}
	id, err := r.LastInsertId()
	if err != nil {
		return
	}
	return int(id), false, nil
}

// Get comment
func (db UserDB) Get(username string) (userid int, pwhash string, notFound bool, err error) {
	q := `SELECT id, pwhash FROM users WHERE name=?`
	row := db.QueryRow(q, username)
	err = row.Scan(&userid, &pwhash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", true, nil
		}
		return 0, "", false, err
	}
	return userid, pwhash, false, nil
}
