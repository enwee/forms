package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"

	"forms/models"
	// already imported in models/userDB.go to use mysql.MySQLError types
	// so no need to _ import just for the init() driver part
	// _ "github.com/go-sql-driver/mysql"
)

type form interface {
	GetAll(userid int) (forms []models.Form, err error)
	New(userid int) (id int, err error)
	Delete(id, userid int) error
	Get(id, userid int) (title string, formItems []models.FormItem, found bool, err error)
	Update(id, userid int, title string, formItems []models.FormItem) error
	Use(id int) (title, updated string, formItems []models.FormItem, found bool, err error)
	Check(id, userid int) (bool, error)
}

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	user     models.UserDB
	form
	response models.ResponseDB
	tmpl     *template.Template
	re       *regexp.Regexp
	session
}

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

	re := regexp.MustCompile(`(^(add|del|upp|dwn|txt|cxb|sel)\d+$|^opt\d+ (add|del|upp|dwn)\d+$)`)

	s := session{sid: map[string]models.User{}, uid: map[int]string{}}

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		user:     models.UserDB{DB: db},
		form:     models.FormDB{DB: db},
		response: models.ResponseDB{DB: db},
		tmpl:     tmpl,
		re:       re,
		session:  s,
	}

	// Doing TLS server here is ok but using a self signed cert is
	// not an acceptable experience when hosted for public access,
	// site will be marked unsafe - cannot verify Root CA (self signed cert).
	//
	// herokuapp.com domain already provides TLS if server is http(non TLS)
	// and that is what i am going to use because the free tier i am using
	// there is no shell access to install certbot for verification required
	// to obtain a free let's encrypt cert which can be verified.
	//
	// TLS is necessary because otherwise signup/login page request body sends
	// password in plaintext. TLS will encrypt this transmission.

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
