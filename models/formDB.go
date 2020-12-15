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
func (db FormDB) GetAll(userid int) (forms []Form, err error) {
	q := `SELECT id, title, updated FROM forms WHERE userid=?`
	rows, err := db.Query(q, userid)
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
func (db FormDB) Get(id, userid int) (title string, formItems []FormItem, found bool, err error) {
	q := `SELECT title, formitems FROM forms WHERE id=? AND userid=?`
	return db.get(q, id, userid)
}

// Use comment
func (db FormDB) Use(id int) (title string, formItems []FormItem, found bool, err error) {
	q := `SELECT title, formitems FROM forms WHERE id=?`
	return db.get(q, id)
}

func (db FormDB) get(q string, ids ...interface{}) (title string, formItems []FormItem, found bool, err error) {
	formItemsJSON := ""
	row := db.QueryRow(q, ids...)
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
func (db FormDB) New(userid int) (id int, err error) {
	newFormItemsJSON := `[{"Label":"Text box","Type":"text","Options":null},{"Label":"Check box","Type":"checkbox","Options":null},{"Label":"Drop down select","Type":"select","Options":["option1","option2"]}]`

	q := `INSERT INTO forms (title, formitems, updated, userid) VALUES ("New Form", ?, NOW(), ?)`
	r, err := db.Exec(q, newFormItemsJSON, userid)
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
func (db FormDB) Delete(id, userid int) error {
	q := `DELETE FROM forms WHERE id=? AND userid=?`
	_, err := db.Exec(q, id, userid)
	if err != nil {
		return err
	}
	return nil
}

// Update comment
func (db FormDB) Update(id, userid int, title string, formItems []FormItem) error {
	b, err := json.Marshal(formItems)
	if err != nil {
		return err
	}
	formItemsJSON := string(b)

	q := `UPDATE forms SET title=?, formitems=?, updated=NOW() WHERE id=? AND userid=?`
	_, err = db.Exec(q, title, formItemsJSON, id, userid)
	if err != nil {
		return err
	}
	return nil
}
