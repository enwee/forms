package main

import (
	"net/http"
	"runtime/debug"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.PanicHandler = app.panic

	router.HandlerFunc("GET", "/", app.auth(app.home))

	router.HandlerFunc("GET", "/edit", app.auth(app.chooseForm))
	router.HandlerFunc("POST", "/edit", app.auth(app.addRemForm))
	router.HandlerFunc("GET", "/edit/:id", app.auth(app.viewForm))
	router.HandlerFunc("POST", "/edit/:id", app.auth(app.editForm))

	router.HandlerFunc("GET", "/use/:id", app.useForm)
	// router.HandlerFunc("POST", "/use/:id", app.useForm) // WIP

	router.HandlerFunc("GET", "/login", app.login)
	router.HandlerFunc("POST", "/login", app.login)
	router.HandlerFunc("GET", "/signup", app.signup)
	router.HandlerFunc("POST", "/signup", app.signup)
	router.HandlerFunc("GET", "/logout", app.logout)

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
