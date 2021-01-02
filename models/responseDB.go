package models

import (
	"database/sql"
	"encoding/json"

	"github.com/go-sql-driver/mysql"
)

// mysql statement to create the tables:
//
// CREATE TABLE `versions` (
// 	`formid` int NOT NULL,
// 	`version` datetime NOT NULL,
// 	`title` char(50) NOT NULL,
// 	`formkeys` text NOT NULL,
// 	PRIMARY KEY (`formid`,`version`)
// );
//
// CREATE TABLE `responses` (
//	`id` int NOT NULL AUTO_INCREMENT,
// 	`formvalues` text NOT NULL,
// 	`created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
//  `formid` int NOT NULL,
//  `version` datetime NOT NULL,
//  PRIMARY KEY (`id`)
// );

// ResponseDB is the database handle with functions to access users table
type ResponseDB struct {
	*sql.DB
}

// New inserts a form response into version and response table
// form title and keys are the same for a version of the form
// and can have many responses per version
func (db ResponseDB) New(r PostResponse) error {
	// insert into versions table if first response to this formversion
	b, err := json.Marshal(r.FormKeys)
	if err != nil {
		return err
	}
	q := `INSERT INTO versions (formid, version, title, formkeys) VALUES (?, ?, ?, ?)`
	_, err = db.Exec(q, r.FormID, r.Version, r.Title, string(b))
	if err != nil {
		// Error 1062: Duplicate entry 'xyz' for key 'versions.PRIMARY'
		if err.(*mysql.MySQLError).Number != 1062 {
			return err
		}
	}
	// insert into responses table - **consider how to do as transaction
	b, err = json.Marshal(r.FormValues)
	if err != nil {
		return err
	}
	q = `INSERT INTO responses (formvalues, formid, version) VALUES (?, ?, ?)`
	_, err = db.Exec(q, string(b), r.FormID, r.Version)
	if err != nil {
		return err
	}
	return nil
}

// Get all past responses to the form (by id)
func (db ResponseDB) Get(id int) (versions []ResponseSet, err error) {
	// get versions
	q := `SELECT version, title, formkeys FROM versions WHERE formid=?`
	rows, err := db.Query(q, id)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var v ResponseSet
		formKeysJSON := ""
		err = rows.Scan(&v.Version, &v.Title, &formKeysJSON)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal([]byte(formKeysJSON), &v.TableHeader)
		if err != nil {
			return
		}
		v.TableHeader = append(v.TableHeader, "created")
		versions = append(versions, v)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(versions) == 0 {
		return nil, nil
	}
	// get responses
	indexer := 0
	curVersion := versions[0].Version
	q = `SELECT id, formvalues, created, version FROM responses WHERE formid=?`
	rows, err = db.Query(q, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var r Response
		formValuesJSON := ""
		created := ""
		err = rows.Scan(&r.ID, &formValuesJSON, &created, &r.Version)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal([]byte(formValuesJSON), &r.Data)
		if err != nil {
			return
		}
		r.Data = append(r.Data, created)
		if r.Version != curVersion {
			indexer++
			curVersion = r.Version
		}
		versions[indexer].TableData = append(versions[indexer].TableData, r)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return versions, nil
}
