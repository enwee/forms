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

	"github.com/julienschmidt/httprouter"
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
		data:     mock{},
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
	scrape              func(io.Reader) editFormPage
}

func TestViewFormLayout(t *testing.T) {
	tests := []test{
		{
			name:      "Empty Label",
			title:     "Lam's BBQ Order",
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
		test.scrape = scrapeViewForm
		runTest(test, t)
	}
}

func TestEditFormLayout(t *testing.T) {
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
		test.scrape = scrapeEditForm
		runTest(test, t)
	}
}
func TestEditFormActions(t *testing.T) {
	tests := []test{
		{
			name:   "Add form item",
			action: "add0",
			formItems: []formItem{
				{"Order", "text", nil},
				{"Contact", "text", nil},
			},
			expected: []formItem{
				{"Order", "text", nil},
				{"", "text", nil},
				{"Contact", "text", nil},
			},
		},
		{
			name:   "Delete form item",
			action: "del1",
			formItems: []formItem{
				{"Order", "text", nil},
				{"", "text", nil},
				{"Contact", "text", nil},
			},
			expected: []formItem{
				{"Order", "text", nil},
				{"Contact", "text", nil},
			},
		},
		{
			name:   "Move form item up",
			action: "upp1",
			formItems: []formItem{
				{"1", "text", nil},
				{"2", "text", nil},
				{"3", "text", nil},
			},
			expected: []formItem{
				{"2", "text", nil},
				{"1", "text", nil},
				{"3", "text", nil},
			},
		},
		{
			name:   "Move form item down",
			action: "dwn1",
			formItems: []formItem{
				{"1", "text", nil},
				{"2", "text", nil},
				{"3", "text", nil},
			},
			expected: []formItem{
				{"1", "text", nil},
				{"3", "text", nil},
				{"2", "text", nil},
			},
		},
		{
			name:   "Change form item type txt/select",
			action: "sel0",
			formItems: []formItem{
				{"1", "text", nil},
			},
			expected: []formItem{
				{"1", "select", []string{""}},
			},
		},
		{
			name:   "Change form item type select/cxb",
			action: "cxb0",
			formItems: []formItem{
				{"1", "select", []string{""}},
			},
			expected: []formItem{
				{"1", "checkbox", nil},
			},
		},
		{
			name:   "Add select option item",
			action: "opt0 add0",
			formItems: []formItem{
				{"Order", "select", []string{"1", "2"}},
			},
			expected: []formItem{
				{"Order", "select", []string{"1", "", "2"}},
			},
		},
		{
			name:   "Delete select option item",
			action: "opt0 del1",
			formItems: []formItem{
				{"Order", "select", []string{"1", "2", "3"}},
			},
			expected: []formItem{
				{"Order", "select", []string{"1", "3"}},
			},
		},
		{
			name:   "Move select option item up",
			action: "opt0 upp1",
			formItems: []formItem{
				{"Order", "select", []string{"1", "2", "3"}},
			},
			expected: []formItem{
				{"Order", "select", []string{"2", "1", "3"}},
			},
		},
		{
			name:   "Move select option item down",
			action: "opt0 dwn1",
			formItems: []formItem{
				{"Order", "select", []string{"1", "2", "3"}},
			},
			expected: []formItem{
				{"Order", "select", []string{"1", "3", "2"}},
			},
		},
	}

	for _, test := range tests {
		test.editMode = true
		test.scrape = scrapeEditForm
		runTest(test, t)
	}
}

func runTest(test test, t *testing.T) {
	t.Run(test.name, func(t *testing.T) {
		data := editFormPage{test.title, "", test.formItems, test.editMode}
		expected := editFormPage{test.title, "", test.expected, test.editMode}
		if test.action == "view" || test.action == "edit" {
			expected = data
		}

		w := httptest.NewRecorder()
		r, err := http.NewRequest("POST", "/1", makePostBody(data, test.action))
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		router := httprouter.New()
		router.HandlerFunc("POST", "/:id", app.editForm)
		router.ServeHTTP(w, r)

		resp := w.Result()
		scrapped := test.scrape(resp.Body)
		if !reflect.DeepEqual(expected, scrapped) {
			t.Errorf("\nrequest(expected):\n%+v\nscrapped(received):\n%+v", data, scrapped)
		}
	})
}
