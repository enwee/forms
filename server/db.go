package main

import (
	"database/sql"
	"encoding/json"
	"errors"
)

func (db database) getAll() (forms []formAttr, err error) {
	q := `SELECT id, title, updated FROM forms`
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		form := formAttr{}
		err = rows.Scan(&form.ID, &form.Title, &form.Updated)
		if err != nil {
			return nil, err
		}
		forms = append(forms, form)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return forms, nil
}

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

func (db database) new() (id int, err error) {
	q := `INSERT INTO forms (title, formitems, updated) VALUES 
	("New Form", '[{"Label":"Text box","Type":"text","Options":null},{"Label":"Check box","Type":"checkbox","Options":null},{"Label":"Drop down select","Type":"select","Options":["option1","option2"]}]', NOW())`
	r, err := db.Exec(q)
	if err != nil {
		return 0, err
	}
	formID, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(formID), nil
}

func (db database) delete(id int) error {
	q := `DELETE FROM forms where id=?`
	_, err := db.Exec(q, id)
	if err != nil {
		return err
	}
	return nil
}

func (db database) update(id int, title string, formItems []formItem) error {
	b, err := json.Marshal(formItems)
	if err != nil {
		return err
	}
	formItemsJSON := string(b)

	q := `UPDATE forms SET title=?, formitems=?, updated=NOW() where id=?`
	_, err = db.Exec(q, title, formItemsJSON, id)
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
