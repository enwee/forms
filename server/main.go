package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
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

	// temp
	// go test(&app)
	// temp

	err = http.ListenAndServe(":"+port, app.handlePanic(router))
	errorLog.Fatal(err)
}
