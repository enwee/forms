package main

import (
	"net/http"
	"runtime/debug"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.PanicHandler = app.panic

	router.HandlerFunc("GET", "/", app.chooseForm)
	router.HandlerFunc("GET", "/edit", app.chooseForm)
	router.HandlerFunc("POST", "/edit", app.addRemForm)
	router.HandlerFunc("GET", "/edit/:id", app.getForm)
	router.HandlerFunc("POST", "/edit/:id", app.makeForm)
	router.HandlerFunc("GET", "/favicon.ico", app.favicon)
	router.HandlerFunc("GET", "/style.css", app.style)

	return router
}

func (app *application) panic(w http.ResponseWriter, r *http.Request, err interface{}) {
	if err != nil {
		app.errorLog.Println(err, string(debug.Stack()))
		http.Error(w, "500 Internal Server Error", 500)
	}
}
