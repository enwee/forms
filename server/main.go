package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
)

type data interface {
	get(id int) (title string, formItems []formItem, found bool, err error)
	put(id int, title string, formItems []formItem) error
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

	fileServer := http.FileServer(http.Dir("./ui/img"))

	router := http.NewServeMux()
	router.HandleFunc("/", app.makeForm)
	router.Handle("/favicon.ico", fileServer)
	infoLog.Println("Server starting at port :", port)

	err = http.ListenAndServe(":"+port, app.handlePanic(router))
	errorLog.Fatal(err)
}
