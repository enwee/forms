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

var baseWithEmptyLabel = []formItem{
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

type test struct {
	name, action, title string
	formItems, expected []formItem
	editMode            bool
	scrape              func(io.Reader) form
}

func TestMakeFormPreview(t *testing.T) {
	tests := []test{
		{
			name:      "Empty Label",
			title:     "Lam's BBQ Order",
			formItems: baseWithEmptyLabel,
		},
		{
			name:      "Empty Title",
			title:     "",
			formItems: baseWithEmptyLabel,
		},
		{
			name:  "Empty Option",
			title: "Lam's",
			formItems: []formItem{
				{"Order", "select", []string{""}},
				{"Contact", "text", nil},
			},
		},
		{
			name:  "Empty Label with Select Options",
			title: "Lam's",
			formItems: []formItem{
				{"", "select", []string{"1", "2", "3"}},
				{"Contact", "text", nil},
			},
		},
	}

	for _, test := range tests {
		test.action = "view"
		test.editMode = false
		test.scrape = scrapePreviewPage
		runTest(test, t)
	}
}

func TestMakeFormEdit(t *testing.T) {
	tests := []test{
		{
			name:      "Empty Label",
			title:     "Lam's BBQ Order",
			formItems: baseWithEmptyLabel,
		},
		{
			name:      "Empty Title",
			title:     "",
			formItems: baseWithEmptyLabel,
		},
		{
			name:  "Empty Option",
			title: "Lam's",
			formItems: []formItem{
				{"Order", "select", []string{""}},
				{"Contact", "text", nil},
			},
		},
		{
			name:  "Empty Label with Select Options",
			title: "Lam's",
			formItems: []formItem{
				{"", "select", []string{"1", "2", "3"}},
				{"Contact", "text", nil},
			},
		},
	}

	for _, test := range tests {
		test.action = "edit"
		test.editMode = true
		test.scrape = scrapeEditPage
		runTest(test, t)
	}
}

func runTest(test test, t *testing.T) {
	t.Run(test.name, func(t *testing.T) {
		data := form{test.title, test.formItems, test.editMode}
		w := httptest.NewRecorder()
		r, err := http.NewRequest("POST", "/", makeBody(data, test.action))
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		app.makeForm(w, r)
		resp := w.Result()
		scrapped := test.scrape(resp.Body)
		if !reflect.DeepEqual(data, scrapped) {
			t.Errorf("\nrequest(expected):\n%+v\nscrapped(received):\n%+v", data, scrapped)
		}
	})
}
