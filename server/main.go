package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
)

type data interface {
	getAll() (forms []formAttr, err error)
	new() (id int, err error)
	delete(id int) error
	get(id int) (title string, formItems []formItem, found bool, err error)
	update(id int, title string, formItems []formItem) error
}

type database struct {
	*sql.DB
}

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	data
	tmpl *template.Template
}

type formItem struct {
	Label   string
	Type    string
	Options []string
}

var re = regexp.MustCompile(`(^(add|del|upp|dwn|txt|cxb|sel)\d+$|^opt\d+ (add|del|upp|dwn)\d+$)`)

func main() {
	errorLog := log.New(os.Stderr, "error:\t", log.LstdFlags|log.Lshortfile)
	infoLog := log.New(os.Stderr, "info:\t", log.LstdFlags)

	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = "formsSvr:password@(localhost:3306)/formsapp"
	}

	db, err := openDB(dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	tmpl, err := template.New("").
		Funcs(template.FuncMap{"minus1": minus1}).
		ParseGlob("./ui/html/*.tmpl")
	if err != nil {
		errorLog.Fatal(err)
	}

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		data:     database{db},
		tmpl:     tmpl,
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	router := httprouter.New()
	router.HandlerFunc("GET", "/", app.chooseForm)
	router.HandlerFunc("GET", "/edit", app.chooseForm)
	router.HandlerFunc("POST", "/edit", app.addRemForm)
	router.HandlerFunc("GET", "/edit/:id", app.getForm)
	router.HandlerFunc("POST", "/edit/:id", app.makeForm)
	router.HandlerFunc("GET", "/favicon.ico", app.favicon)

	infoLog.Println("Server starting at port :", port)

	err = http.ListenAndServe(":"+port, app.handlePanic(router))
	errorLog.Fatal(err)
}
