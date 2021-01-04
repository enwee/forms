package main

import (
	"forms/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func (app *application) useForm(w http.ResponseWriter, r *http.Request) {
	feedback := ""
	id, err := strconv.Atoi(httprouter.ParamsFromContext(r.Context()).ByName("id"))
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "400 Invalid data", 400)
		return
	}
	if r.Method == http.MethodGet {
		// Prevent demo forms from being used, currently id 1/2/3
		// Think of a better way
		// Switch this if block off to allow demo form be 'use'able
		if id == 1 || id == 2 || id == 3 {
			http.Error(w, "404 Form not found", 404)
			return
		}
		feedback = getFeedback(w, r) // get flash message if any
	}

	title, updated, formItems, found, err := app.form.Use(id)
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "500 Internal Server Error", 500)
		return
	}
	if !found {
		app.errorLog.Printf("form id:%v not found", id)
		http.Error(w, "404 Form not found", 404)
		return
	}

	if r.Method == http.MethodPost {
		version := r.FormValue("version")
		if version != updated {
			http.Error(w, "400 Invalid Data or Form has changed", 400)
			return
		}
		keys, values := []string{}, []string{}
		for index, formItem := range formItems {
			if formItem.Label == "" {
				continue
			}
			keys = append(keys, formItem.Label)
			value := strings.TrimSpace(r.FormValue(strconv.Itoa(index)))
			if formItem.Type == "checkbox" && value == "on" {
				value = "✅"
			}
			values = append(values, value)
		}

		resp := models.PostResponse{
			FormID:     id,
			Version:    version,
			Title:      title,
			FormKeys:   keys,
			FormValues: values,
		}
		if err := app.response.New(resp); err != nil {
			app.errorLog.Print(err)
			http.Error(w, "500 Internal Server Error", 500)
			return
		}
		setFeedback(w, "Response Sent")
		http.Redirect(w, r, "/use/"+strconv.Itoa(id), 303)
	}

	pageData := pageData{
		Form:     models.Form{ID: id, Title: title, FormItems: formItems, Updated: updated},
		Feedback: feedback,
	}
	err = app.tmpl.ExecuteTemplate(w, "use", pageData)
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "500 Internal Server Error", 500)
		return
	}
}

func (app *application) viewResp(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(contextKey("user")).(models.User)
	id, err := strconv.Atoi(httprouter.ParamsFromContext(r.Context()).ByName("id"))
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "400 Invalid data", 400)
		return
	}
	ok, err := app.form.Check(id, u.ID)
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "500 Internal Server Error", 500)
		return
	}
	if !ok {
		app.errorLog.Printf("form id:%v user:%v not found", id, u.Name)
		http.Error(w, "404 Form not found", 404)
		return
	}
	versions, err := app.response.Get(id)
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "500 Internal Server Error", 500)
		return
	}
	pageData := struct {
		Versions []models.ResponseSet
		models.User
		PageMode int
	}{versions, u, respMode}
	err = app.tmpl.ExecuteTemplate(w, "form", pageData)
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "500 Internal Server Error", 500)
		return
	}
}

func (app *application) delResp(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	var id int
	var err error

	if !stringIs(action, "choose", "auth") {
		action, id, err = getAction(action)
		if err != nil {
			app.errorLog.Print(err)
			http.Error(w, "400 Invalid data", 400)
			return
		}
	}

	u := r.Context().Value(contextKey("user")).(models.User)

	switch action {
	case "del":
		_ = id
		// err = app.form.Delete(id, u.ID)
		// if err != nil {
		// 	app.errorLog.Print(err)
		// 	http.Error(w, "500 Internal Server Error", 500)
		// 	return
		// }
	case "choose":
		http.Redirect(w, r, "/edit", 303)
		return
	case "auth":
		if u.ID == 0 {
			http.Redirect(w, r, "/login", 303)
			return
		}
		http.Redirect(w, r, "/logout", 303)
		return
	}

	http.Redirect(w, r, "/edit", http.StatusSeeOther)
}
