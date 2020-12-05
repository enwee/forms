package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	tmpl     *template.Template
}

type formItem struct {
	Type    string
	Label   string
	Options []string
}

type form struct {
	Title     string
	FormItems []formItem
	EditMode  bool
}

func main() {
	errorLog := log.New(os.Stderr, "error:\t", log.LstdFlags|log.Lshortfile)
	infoLog := log.New(os.Stderr, "info:\t", log.LstdFlags)

	t, err := template.New("").
		Funcs(template.FuncMap{"minus1": minus1}).
		ParseGlob("ui/html/*.tmpl")
	if err != nil {
		errorLog.Fatal(err)
	}

	app := application{
		errorLog: errorLog,
		infoLog:  infoLog,
		tmpl:     t,
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	router := http.NewServeMux()
	router.HandleFunc("/", app.makeForm)
	fmt.Println("Server starting at port :", port)

	err = http.ListenAndServe(":"+port, app.handlePanic(router))
	errorLog.Fatal(err)
}

func (app *application) makeForm(w http.ResponseWriter, r *http.Request) {
	formItems := []formItem{}
	if r.Method == "POST" {
		r.ParseForm()

		title := r.FormValue("title")

		//action := r.Form["action"][0] //must check action exists else panic
		actions := strings.Split(r.FormValue("action"), " ")
		action := actions[0]
		index := 0
		err := errors.New("")
		editMode := true

		if action != "edit" && action != "view" {
			action, index, err = getAction(action)
			if err != nil {
				app.errorLog.Print(err)
				http.Error(w, "Invalid action index value", 400)
				return
			}
		}

		labels := r.Form["label"]
		inputType := r.Form["type"] //check both len same
		for i, label := range labels {
			options := []string{}
			if inputType[i] == "select" {
				opts := r.Form["options"+strconv.Itoa(i)]
				for _, option := range opts {
					options = append(options, option)
				}
			}
			formItems = append(formItems, formItem{inputType[i], label, options})
		}

		switch action {
		case "add":
			formItems = append(formItems[:index+1], formItems[index:]...)
			formItems[index+1] = formItem{"text", "", []string{}}
		case "del":
			if len(labels) == 1 {
				formItems = []formItem{{"text", "", []string{}}}
			} else {
				formItems = append(formItems[:index], formItems[index+1:]...)
			}
		case "upp":
			if index != 0 {
				formItems[index-1], formItems[index] = formItems[index], formItems[index-1]
			}
		case "dwn":
			if index != len(labels)-1 {
				formItems[index], formItems[index+1] = formItems[index+1], formItems[index]
			}
		case "opt":
			options := formItems[index].Options
			action, idx, err := getAction(actions[1])
			if err != nil {
				app.errorLog.Print(err)
				http.Error(w, "Invalid action index value", 400)
				return
			}
			switch action {
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
		case "cxb":
			formItems[index].Type = "checkbox"
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
	formItems = []formItem{
		{"select", "Order", []string{"Chicken", "Fish", "Pork", "Beef"}},
		{"select", "Qty", []string{"1", "2", "3"}},
		{"checkbox", "Chilli packs", []string{}},
		{"checkbox", "Disposable cutlery", []string{}},
		{"text", "Comments", []string{}},
		{"text", "", []string{}},
		{"text", "Name", []string{}},
		{"text", "Email", []string{}},
		{"text", "Contact", []string{}},
		{"text", "Delivery address (if any)", []string{}},
		{"checkbox", "Self collection", []string{}},
	}
	err := app.tmpl.ExecuteTemplate(w, "layout", form{"Lam's BBQ Order Form", formItems, false})
	// fields = []formItem{{"text", ""}}
	// err := app.tmpl.Execute(w, form{"", fields, true})
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
