package main

import (
	"net/http"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(contextKey("user")).(user)
	if user.ID == 0 {
		http.Redirect(w, r, "/edit/1", 303)
		return
	}
	http.Redirect(w, r, "/edit", 303)
}

func (app *application) favicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./ui/img/favicon.ico")
}

func (app *application) style(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./ui/css/style.css")
}
