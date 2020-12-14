package models

import (
	"database/sql"
	"encoding/json"
	"errors"
)

// FormDB comment
type FormDB struct {
	*sql.DB
}

// GetAll comment
func (db FormDB) GetAll() (forms []Form, err error) {
	q := `SELECT id, title, updated FROM forms`
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		form := Form{}
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

// Get comment
func (db FormDB) Get(id int) (title string, formItems []FormItem, found bool, err error) {
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

// New comment
func (db FormDB) New() (id int, err error) {
	newFormItemsJSON := `[{"Label":"Text box","Type":"text","Options":null},{"Label":"Check box","Type":"checkbox","Options":null},{"Label":"Drop down select","Type":"select","Options":["option1","option2"]}]`

	q := `INSERT INTO forms (title, formitems, updated) VALUES ("New Form", ?, NOW())`
	r, err := db.Exec(q, newFormItemsJSON)
	if err != nil {
		return 0, err
	}
	formID, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(formID), nil
}

// Delete comment
func (db FormDB) Delete(id int) error {
	q := `DELETE FROM forms where id=?`
	_, err := db.Exec(q, id)
	if err != nil {
		return err
	}
	return nil
}

// Update comment
func (db FormDB) Update(id int, title string, formItems []FormItem) error {
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
