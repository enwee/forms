package main

import (
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type chooseFormPageItem struct {
	ID      int
	Title   string
	Updated string
}

type editFormPage struct {
	Title     string
	TitleErr  string
	FormItems []formItem
	EditMode  bool
}

func (app *application) chooseForm(w http.ResponseWriter, r *http.Request) {
	forms, err := app.data.getAll()
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "500 Internal Server Error", 500)
		return
	}
	err = app.tmpl.ExecuteTemplate(w, "choose", forms)
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "500 Internal Server Error", 500)
		return
	}
}

func (app *application) addRemForm(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	var id int
	var err error

	if action != "add" {
		action, id, err = getAction(action)
		if err != nil {
			app.errorLog.Print(err)
			http.Error(w, "400 Invalid data", 400)
			return
		}
	}

	switch action {
	case "add":
		_, err := app.data.new()
		if err != nil {
			app.errorLog.Print(err)
			http.Error(w, "500 Internal Server Error", 500)
			return
		}
	case "del":
		err = app.data.delete(id)
		if err != nil {
			app.errorLog.Print(err)
			http.Error(w, "500 Internal Server Error", 500)
			return
		}
	}

	http.Redirect(w, r, "/edit", http.StatusSeeOther)
}

func (app *application) viewForm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(httprouter.ParamsFromContext(r.Context()).ByName("id"))
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "400 Invalid data", 400)
		return
	}
	title, formItems, found, err := app.data.get(id)
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
	err = app.tmpl.ExecuteTemplate(w, "form", editFormPage{title, "", formItems, false})
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "500 Internal Server Error", 500)
		return
	}
}

func (app *application) editForm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(httprouter.ParamsFromContext(r.Context()).ByName("id"))
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "400 Invalid data", 400)
		return
	}
	editMode := true
	title, titleErr := validateTitle(r)
	formItems, action, opt, index, idx, err := validateForm(r)
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "400 Invalid data", 400)
		return
	}

	switch action {
	case "add":
		formItems = append(formItems[:index+1], formItems[index:]...)
		formItems[index+1] = formItem{"", "text", nil}
	case "del":
		if len(formItems) == 1 {
			formItems = []formItem{{"", "text", nil}}
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
		editMode = true // does nothing, editMode=true is the default
	case "view":
		if titleErr != "" {
			break
		}
		editMode = false
		err = app.data.update(id, title, formItems)
		if err != nil {
			app.errorLog.Print(err)
			http.Error(w, "500 Internal Server Error", 500)
			return
		}
	case "change":
		http.Redirect(w, r, "/edit", http.StatusSeeOther)
		return
	}
	err = app.tmpl.ExecuteTemplate(w, "form", editFormPage{title, titleErr, formItems, editMode})
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
	title, formItems, found, err := app.data.get(id)
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
	err = app.tmpl.ExecuteTemplate(w, "form.use", editFormPage{title, "", formItems, false})
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "500 Internal Server Error", 500)
		return
	}
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/edit", 303)
}

func (app *application) favicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./ui/img/favicon.ico")
}

func (app *application) style(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./ui/css/style.css")
}
