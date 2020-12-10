package main

import (
	"database/sql"
	"encoding/json"

	"errors"
)

func (db database) get(id int) (title string, formItems []formItem, found bool, err error) {
	formItemsJSON := ""

	q := `SELECT title, formitems FROM forms WHERE id=?`
	row := db.QueryRow(q, id)
	err = row.Scan(&title, &formItemsJSON)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return
		}
		err = nil
		return
	}
	found = true

	err = json.Unmarshal([]byte(formItemsJSON), &formItems)
	if err != nil {
		return
	}
	return
}

func (db database) put(id int, title string, formItems []formItem) error {
	b, err := json.Marshal(formItems)
	if err != nil {
		return err
	}
	formItemsJSON := string(b)

	q := `INSERT INTO forms VALUES (?, ?, ?, UTC_TIMESTAMP()) 
	ON DUPLICATE KEY UPDATE title=?, formitems=?, updated=UTC_TIMESTAMP()`
	_, err = db.Exec(q, id, title, formItemsJSON, title, formItemsJSON)
	if err != nil {
		return err
	}
	return nil
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
