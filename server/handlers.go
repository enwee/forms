package main

import (
	"net/http"
)

type userID string

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	userid := r.Context().Value(userID("userid")).(int)
	if userid == 0 {
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
