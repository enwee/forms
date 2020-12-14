package main

import (
	"net/http"
	"runtime/debug"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.PanicHandler = app.panic

	router.HandlerFunc("GET", "/", app.home)

	router.HandlerFunc("GET", "/edit", app.chooseForm)
	router.HandlerFunc("POST", "/edit", app.addRemForm)
	router.HandlerFunc("GET", "/edit/:id", app.viewForm)
	router.HandlerFunc("POST", "/edit/:id", app.editForm)

	router.HandlerFunc("GET", "/use/:id", app.useForm)
	// router.HandlerFunc("POST", "/use/:id", app.useForm)

	// router.HandlerFunc("GET", "/login", app.login)
	// router.HandlerFunc("POST", "/login", app.login)

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
