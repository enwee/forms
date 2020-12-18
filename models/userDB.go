package models

import (
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
)

// mysql statement to create the table:
//
// CREATE TABLE `users` (
// 	`id` int NOT NULL PRIMARY KEY AUTO_INCREMENT,
// 	`name` char(8) NOT NULL UNIQUE,
// 	`pwhash` char(60) NOT NULL,
// 	`created` datetime NOT NULL
// );

// UserDB is the database handle with functions to access users table
type UserDB struct {
	*sql.DB
}

// New creates a new user
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

// Get a user id and pwhash
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
