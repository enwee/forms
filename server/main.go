package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	tmpl     *template.Template
}

type formItem struct {
	Label   string
	Type    string
	Options []string
}

type form struct {
	Title     string
	FormItems []formItem
	EditMode  bool
}

var re = regexp.MustCompile(`(^(add|del|upp|dwn|txt|cxb|sel)\d+$|^opt\d+ (add|del|upp|dwn)\d+$)`)

func main() {
	errorLog := log.New(os.Stderr, "error:\t", log.LstdFlags|log.Lshortfile)
	infoLog := log.New(os.Stderr, "info:\t", log.LstdFlags)

	tmpl, err := template.New("").
		Funcs(template.FuncMap{"minus1": minus1}).
		ParseGlob("ui/html/*.tmpl")
	if err != nil {
		errorLog.Fatal(err)
	}

	app := application{
		errorLog: errorLog,
		infoLog:  infoLog,
		tmpl:     tmpl,
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	router := http.NewServeMux()
	router.HandleFunc("/", app.makeForm)
	fmt.Println("Server starting at port :", port)

	err = http.ListenAndServe(":"+port, app.handlePanic(router))
	errorLog.Fatal(err)
}
