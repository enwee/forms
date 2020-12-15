package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type session struct {
	sid map[string]int
	uid map[int]string
}

type logonPage struct {
	Title    string
	Username string
	Err      string
}

func (app *application) signup(w http.ResponseWriter, r *http.Request) {
	var username, userError string
	if r.Method == http.MethodPost {
		username = strings.TrimSpace(r.FormValue("username"))
		userError = validateUsername(username)
		if userError == "" {
			pw := r.FormValue("password")
			pwhash, err := bcrypt.GenerateFromPassword([]byte(pw), 12)
			if err != nil {
				app.errorLog.Print(err)
				http.Error(w, "500 Internal Server Error", 500)
				return
			}
			userid, duplicate, err := app.user.New(username, string(pwhash))
			if err != nil {
				app.errorLog.Print(err)
				http.Error(w, "500 Internal Server Error", 500)
				return
			}
			if duplicate {
				userError = "that user name is not available"
			} else {
				app.newSession(w, userid)
				http.Redirect(w, r, "/edit", http.StatusSeeOther)
				return
			}
		}
	}
	// GET also comes here directly
	err := app.tmpl.ExecuteTemplate(w, "logon", logonPage{"Sign up", username, userError})
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "500 Internal Server Error", 500)
	}
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	var username, userError string
	if r.Method == http.MethodPost {
		username = strings.TrimSpace(r.FormValue("username"))
		userError = validateUsername(username)
		if userError == "" {
			userid, pwhash, notFound, err := app.user.Get(username)
			if err != nil {
				app.errorLog.Print(err)
				http.Error(w, "500 Internal Server Error", 500)
				return
			}
			if notFound {
				userError = "invalid username or password"
			} else {
				pw := r.FormValue("password")
				if bcrypt.CompareHashAndPassword([]byte(pwhash), []byte(pw)) == nil {
					app.newSession(w, userid)
					http.Redirect(w, r, "/edit", http.StatusSeeOther)
					return
				}
				userError = "invalid username or password"
			}
		}
	}
	// // GET also comes here directly
	err := app.tmpl.ExecuteTemplate(w, "logon", logonPage{"Login", username, userError})
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "500 Internal Server Error", 500)
	}
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "sid", MaxAge: -1})
	cookie, err := r.Cookie("sid")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	userid := app.session.sid[cookie.Value]
	delete(app.session.sid, cookie.Value)
	delete(app.session.uid, userid)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return
}

func (app *application) newSession(w http.ResponseWriter, userid int) {
	u := uuid.New().String()
	c := http.Cookie{Name: "sid", Value: u, HttpOnly: true}
	http.SetCookie(w, &c)

	app.session.sid[u] = userid
	delete(app.session.sid, app.session.uid[userid]) //remove prev session if any
	app.session.uid[userid] = u
}

func (app *application) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var userid int
		c, err := r.Cookie("sid")
		if err == http.ErrNoCookie {
			userid = 0
		} else {
			userid = app.session.sid[c.Value] //userid is 0 if invalid
		}
		ctx := context.WithValue(r.Context(), userID("userid"), userid)
		r = r.WithContext(ctx)
		next(w, r)
	}
}
