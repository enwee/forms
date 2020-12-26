package models

import (
	"database/sql"
	"encoding/json"

	"github.com/go-sql-driver/mysql"
)

// mysql statement to create the tables:
//
// CREATE TABLE `responses` (
// 	`formid` int NOT NULL,
// 	`version` datetime NOT NULL,
// 	`formvalues` text NOT NULL,
// 	`posted` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
// );

// CREATE TABLE `versions` (
// 	`formid` int NOT NULL,
// 	`version` datetime NOT NULL,
// 	`title` char(50) NOT NULL,
// 	`formkeys` text NOT NULL,
// 	PRIMARY KEY (`formid`,`version`)
// );

// ResponseDB is the database handle with functions to access users table
type ResponseDB struct {
	*sql.DB
}

// New inserts a form response into version and response table
// form title and keys are the same for a version of the form
// and can have many responses per version
func (db ResponseDB) New(r Response) error {
	// insert into versions table if first response to this formversion
	b, err := json.Marshal(r.FormKeys)
	if err != nil {
		return err
	}
	q := `INSERT INTO versions VALUES (?, ?, ?, ?)`
	_, err = db.Exec(q, r.FormID, r.Version, r.Title, string(b))
	if err != nil {
		// Error 1062: Duplicate entry
		if err.(*mysql.MySQLError).Number != 1062 {
			return err
		}
	}
	// insert into responses table - **consider to do as transaction
	b, err = json.Marshal(r.FormValues)
	if err != nil {
		return err
	}
	q = `INSERT INTO responses (formid, version, formvalues) VALUES (?, ?, ?)`
	_, err = db.Exec(q, r.FormID, r.Version, string(b))
	if err != nil {
		return err
	}
	return nil
}
