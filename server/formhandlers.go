package main

import (
	"net/http"
	"strconv"
	"strings"

	"forms/models"

	"github.com/julienschmidt/httprouter"
)

const (
	chooseMode = iota + 1
	editMode
	viewMode
)

type pageData struct {
	models.Form
	models.User
	Feedback string // change to Errs []string to use multi errs
	PageMode int
}

func (app *application) chooseForm(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(contextKey("user")).(models.User)
	forms, err := app.form.GetAll(u.ID)
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "500 Internal Server Error", 500)
		return
	}
	//overwrite the declared pageData type cos this uses []Form
	pageData := struct {
		Forms []models.Form
		models.User
		PageMode int
	}{forms, u, chooseMode}
	err = app.tmpl.ExecuteTemplate(w, "form", pageData)
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "500 Internal Server Error", 500)
		return
	}
}

func (app *application) addRemForm(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(contextKey("user")).(models.User)
	if u.ID == 0 {
		http.Redirect(w, r, "/login", 303)
		return
	}

	action := r.FormValue("action")
	var id int
	var err error

	if !stringIs(action, "add", "auth") {
		action, id, err = getAction(action)
		if err != nil {
			app.errorLog.Print(err)
			http.Error(w, "400 Invalid data", 400)
			return
		}
	}

	switch action {
	case "add":
		_, err := app.form.New(u.ID)
		if err != nil {
			app.errorLog.Print(err)
			http.Error(w, "500 Internal Server Error", 500)
			return
		}
	case "del":
		err = app.form.Delete(id, u.ID)
		if err != nil {
			app.errorLog.Print(err)
			http.Error(w, "500 Internal Server Error", 500)
			return
		}
	case "auth":
		http.Redirect(w, r, "/logout", 303)
		return
	}

	http.Redirect(w, r, "/edit", http.StatusSeeOther)
}

func (app *application) viewForm(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(contextKey("user")).(models.User)
	id, err := strconv.Atoi(httprouter.ParamsFromContext(r.Context()).ByName("id"))
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "400 Invalid data", 400)
		return
	}
	title, formItems, found, err := app.form.Get(id, u.ID)
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "500 Internal Server Error", 500)
		return
	}
	if !found {
		app.errorLog.Printf("form id:%v not found for user:%s", id, u.Name)
		http.Error(w, "404 Form not found", 404)
		return
	}

	pageData := pageData{
		Form: models.Form{ID: id, Title: title, FormItems: formItems},
		User: u, Feedback: "", PageMode: viewMode,
	}
	err = app.tmpl.ExecuteTemplate(w, "form", pageData)
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "500 Internal Server Error", 500)
		return
	}
}

func (app *application) editForm(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(contextKey("user")).(models.User)
	id, err := strconv.Atoi(httprouter.ParamsFromContext(r.Context()).ByName("id"))
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "400 Invalid data", 400)
		return
	}
	pageMode := editMode
	title, feedback := validateTitle(r)
	formItems, action, opt, index, idx, err := validateForm(r, app.re)
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "400 Invalid data", 400)
		return
	}

	switch action {
	case "add":
		formItems = append(formItems[:index+1], formItems[index:]...)
		formItems[index+1] = models.FormItem{Label: "", Type: "text", Options: nil}
	case "del":
		if len(formItems) == 1 {
			formItems = []models.FormItem{{Label: "", Type: "text", Options: nil}}
		} else {
			formItems = append(formItems[:index], formItems[index+1:]...)
		}
	case "upp":
		if index != 0 {
			formItems[index-1], formItems[index] = formItems[index], formItems[index-1]
		}
	case "dwn":
		if index != len(formItems)-1 {
			formItems[index], formItems[index+1] = formItems[index+1], formItems[index]
		}
	case "opt":
		options := formItems[index].Options
		switch opt {
		case "add":
			options = append(options[:idx+1], options[idx:]...)
			options[idx+1] = ""
		case "del":
			if len(options) == 1 {
				options = []string{""}
			} else {
				options = append(options[:idx], options[idx+1:]...)
			}
		case "upp":
			if idx != 0 {
				options[idx-1], options[idx] = options[idx], options[idx-1]
			}
		case "dwn":
			if idx != len(options)-1 {
				options[idx], options[idx+1] = options[idx+1], options[idx]
			}
		}
		formItems[index].Options = options
	case "txt":
		formItems[index].Type = "text"
		formItems[index].Options = nil
	case "cxb":
		formItems[index].Type = "checkbox"
		formItems[index].Options = nil
	case "sel":
		formItems[index].Type = "select"
		formItems[index].Options = []string{""}
	case "edit":
		pageMode = editMode // does nothing, editMode is the default, more readable than blank line
	case "view":
		if feedback != "" {
			break
		}
		pageMode = viewMode
		if u.ID == 0 {
			break
		}
		err = app.form.Update(id, u.ID, title, formItems)
		if err != nil {
			app.errorLog.Print(err)
			http.Error(w, "500 Internal Server Error", 500)
			return
		}
	case "choose":
		http.Redirect(w, r, "/edit", http.StatusSeeOther)
		return
	case "auth":
		if u.ID == 0 {
			http.Redirect(w, r, "/login", 303)
			return
		}
		http.Redirect(w, r, "/logout", 303)
		return
	}

	pageData := pageData{
		Form: models.Form{ID: id, Title: title, FormItems: formItems},
		User: u, Feedback: feedback, PageMode: pageMode,
	}
	err = app.tmpl.ExecuteTemplate(w, "form", pageData)
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "500 Internal Server Error", 500)
		return
	}
}

func (app *application) useForm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(httprouter.ParamsFromContext(r.Context()).ByName("id"))
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "400 Invalid data", 400)
		return
	}
	if r.Method == http.MethodGet {
		// Prevent demo forms from being used, currently id 1/2/3
		// Think of a better way
		// Switch this off to enter demo data
		if id == 1 || id == 2 || id == 3 {
			http.Error(w, "404 Form not found", 404)
			return
		}
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

	feedback := ""
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
			values = append(values, value)
		}

		format := "id:%v\nupdated:%v\ntitle:%v\nkeys:%#v\nvalues:%#v\n"
		app.infoLog.Printf(format, id, updated, title, keys, values)
		feedback = "Response Sent"
		// without this refresh reposts form but with it loses feedback
		http.Redirect(w, r, "/use/"+strconv.Itoa(id), 303) //loses feedback
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
