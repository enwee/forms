package main

import (
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var app application

func init() {
	tmpl := template.Must(template.New("").
		Funcs(template.FuncMap{"minus1": minus1}).
		ParseGlob("../ui/html/*.tmpl"))

	app = application{
		errorLog: log.New(ioutil.Discard, "", 0),
		infoLog:  log.New(ioutil.Discard, "", 0),
		tmpl:     tmpl,
	}
}

func TestMakeFormModes(t *testing.T) {
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

	tests := []struct {
		name, action string
		editMode     bool
		scrap        func(io.Reader) form
	}{
		{"Preview Mode", "view", false, scrapPreviewBody},
		{"Edit Mode", "edit", true, scrapEditBody},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := form{"Lam's BBQ Order Form", formItems, test.editMode}
			w := httptest.NewRecorder()
			r, err := http.NewRequest("POST", "/", makeBody(data, test.action))
			if err != nil {
				t.Fatal(err)
			}
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			app.makeForm(w, r)
			resp := w.Result()
			scrapped := test.scrap(resp.Body)
			if !reflect.DeepEqual(data, scrapped) {
				t.Errorf("\nrequest(expected):\n%+v\nscrapped(received):\n%+v", data, scrapped)
			}
		})
	}
}
