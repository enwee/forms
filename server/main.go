package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"

	"forms/models"

	_ "github.com/go-sql-driver/mysql"
)

type form interface {
	GetAll() (forms []models.Form, err error)
	New() (id int, err error)
	Delete(id int) error
	Get(id int) (title string, formItems []models.FormItem, found bool, err error)
	Update(id int, title string, formItems []models.FormItem) error
}

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	form
	tmpl *template.Template
}

// look into unglobalizing this
var re = regexp.MustCompile(`(^(add|del|upp|dwn|txt|cxb|sel)\d+$|^opt\d+ (add|del|upp|dwn)\d+$)`)

func main() {
	errorLog := log.New(os.Stderr, "error:\t", log.LstdFlags|log.Lshortfile)
	infoLog := log.New(os.Stderr, "info:\t", log.LstdFlags)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = "formsSvr:password@(localhost:3306)/formsapp"
	}

	db, err := openDB(dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	tmpl, err := template.New("").Funcs(template.FuncMap{"minus1": minus1}).ParseGlob("./ui/html/*.tmpl")
	if err != nil {
		errorLog.Fatal(err)
	}

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		form:     models.FormDB{DB: db},
		tmpl:     tmpl,
	}

	infoLog.Println("Server starting at port :", port)
	err = http.ListenAndServe(":"+port, app.routes())
	errorLog.Fatal(err)
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
