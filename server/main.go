package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
)

type application struct {
	errorLog, infoLog *log.Logger
	tmpl              *template.Template
}

type formItem struct {
	Type, Value string
}

type form struct {
	Title  string
	Fields []formItem
	Edit   bool
}

var tmpls = []string{
	"./ui/html/layout.tmpl",
	"./ui/html/form.tmpl",
	"./ui/html/form.edit.tmpl",
	"./ui/html/form.view.tmpl",
}

func main() {
	errorLog := log.New(os.Stderr, "error:\t", log.LstdFlags|log.Lshortfile)
	infoLog := log.New(os.Stderr, "info:\t", log.LstdFlags)

	ts, err := template.ParseFiles(tmpls...)
	if err != nil {
		errorLog.Fatal(err)
	}

	app := application{
		errorLog: errorLog,
		infoLog:  infoLog,
		tmpl:     ts,
	}

	router := http.NewServeMux()
	router.HandleFunc("/", app.makeForm)
	fmt.Println("Server starting at port :5000")

	err = http.ListenAndServe(":5000", app.handlePanic(router))
	errorLog.Fatal(err)
}

func (app *application) makeForm(w http.ResponseWriter, r *http.Request) {
	fields := []formItem{}
	if r.Method == "POST" {
		r.ParseForm()

		// app.infoLog.Printf("%+v", r.Form)

		title := r.FormValue("title")

		//action := r.Form["action"][0] //must check action exists else panic
		action := r.FormValue("action")
		index := 0
		err := errors.New("")
		edit := true

		if action != "edit" && action != "view" {
			index, err = strconv.Atoi(action[3:])
			if err != nil {
				app.errorLog.Print(err)
				http.Error(w, "Invalid action index value", 400)
				return
			}
			action = action[:3]
		}

		labels := r.Form["label"]
		inputType := r.Form["type"] //check both len same
		for i, value := range labels {
			fields = append(fields, formItem{inputType[i], value})
		}

		switch action {
		case "add":
			fields = append(fields, formItem{"text", ""})
			copy(fields[index+2:], fields[index+1:])
			fields[index+1] = formItem{"text", ""}
		case "del":
			if len(labels) == 1 {
				fields = []formItem{{"text", ""}}
			} else {
				fields = append(fields[:index], fields[index+1:]...)
			}
		case "upp":
			if index != 0 {
				fields[index-1], fields[index] = fields[index], fields[index-1]
			}
		case "dwn":
			if index != len(labels)-1 {
				fields[index], fields[index+1] = fields[index+1], fields[index]
			}
		case "txt":
			fields[index].Type = "text"
		case "cxb":
			fields[index].Type = "checkbox"
		case "edit":
			edit = true
		case "view":
			edit = false
		}
		// method is POST
		err = app.tmpl.Execute(w, form{title, fields, edit})
		if err != nil {
			app.errorLog.Print(err)
			http.Error(w, "Internal Server Error", 500)
		}
		return
	}

	// method is not POST
	fields = []formItem{{"text", ""}}
	err := app.tmpl.Execute(w, form{"", fields, true})
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
				app.errorLog.Println(err)
				http.Error(w, "Internal Server Error", 500)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
