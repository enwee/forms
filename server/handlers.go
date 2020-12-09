package main

import (
	"net/http"
	"runtime/debug"
)

func (app *application) makeForm(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		title := r.FormValue("title")
		editMode := true
		formItems, action, opt, index, idx, err := validateForm(r)
		if err != nil {
			app.errorLog.Print(err)
			http.Error(w, "Invalid data", 400)
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
			editMode = true
		case "view":
			editMode = false
		}
		// method is POST
		err = app.tmpl.ExecuteTemplate(w, "layout", form{title, formItems, editMode})
		if err != nil {
			app.errorLog.Print(err)
			http.Error(w, "Internal Server Error", 500)
		}
		return
	}

	// method is not POST
	formItems := []formItem{
		{"Order", "select", []string{"Chicken", "Fish", "Pork", "Beef"}},
		{"Qty", "select", []string{"1", "2", "3"}},
		{"Chilli packs", "checkbox", nil},
		{"Disposable cutlery", "checkbox", nil},
		{"Comments", "text", nil},
		{"", "text", nil},
		{"Name", "text", nil},
		{"Email", "text", nil},
		{"Contact", "text", nil},
		{"Delivery address (if any)", "text", nil},
		{"Self collection", "checkbox", nil},
	}
	err := app.tmpl.ExecuteTemplate(w, "layout", form{"Lam's BBQ Order Form", formItems, false})
	if err != nil {
		app.errorLog.Print(err)
		http.Error(w, "Internal Server Error", 500)
	}
}

func (app *application) handlePanic(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				app.errorLog.Println(err, string(debug.Stack()))
				http.Error(w, "Internal Server Error", 500)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
