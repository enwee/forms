package main

import (
	"context"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"testing"

	"forms/models"

	"github.com/julienschmidt/httprouter"
)

var app application

// this is the stuff that main.go does before the handler can work
func init() {
	tmpl := template.Must(template.New("").
		Funcs(template.FuncMap{"minus1": minus1}).
		ParseGlob("../ui/html/*.tmpl"))

	re := regexp.MustCompile(`(^(add|del|upp|dwn|txt|cxb|sel)\d+$|^opt\d+ (add|del|upp|dwn)\d+$)`)

	app = application{
		errorLog: log.New(ioutil.Discard, "", 0),
		infoLog:  log.New(ioutil.Discard, "", 0),
		form:     mockDB{},
		tmpl:     tmpl,
		re:       re,
	}
}

var baseWithEmptyLabel = []models.FormItem{
	{Label: "Qty", Type: "select", Options: []string{"1", "2", "3"}},
	{Label: "Order", Type: "select", Options: []string{"Chicken", "Fish", "Pork", "Beef"}},
	{Label: "Chilli packs", Type: "checkbox", Options: nil},
	{Label: "Disposable cutlery", Type: "checkbox", Options: nil},
	{Label: "Comments", Type: "text", Options: nil},
	{Label: "", Type: "text", Options: nil},
	{Label: "Name", Type: "text", Options: nil},
	{Label: "Email", Type: "text", Options: nil},
	{Label: "Contact", Type: "text", Options: nil},
	{Label: "Delivery address (if any)", Type: "text", Options: nil},
	{Label: "Self collection", Type: "checkbox", Options: nil},
}

type test struct {
	name, action, title string
	formItems, expected []models.FormItem
	pageMode            int
	scrape              func(io.Reader) formPage
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
			formItems: []models.FormItem{
				{Label: "Order", Type: "select", Options: []string{""}},
				{Label: "Contact", Type: "text", Options: nil},
			},
		},
		{
			name:  "Empty Label with Select Options",
			title: "Lam's",
			formItems: []models.FormItem{
				{Label: "", Type: "select", Options: []string{"1", "2", "3"}},
				{Label: "Contact", Type: "text", Options: nil},
			},
		},
	}

	for _, test := range tests {
		test.action = "view"
		test.pageMode = viewMode
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
			formItems: []models.FormItem{
				{Label: "Order", Type: "select", Options: []string{""}},
				{Label: "Contact", Type: "text", Options: nil},
			},
		},
		{
			name:  "Empty Label with Select Options",
			title: "Lam's",
			formItems: []models.FormItem{
				{Label: "", Type: "select", Options: []string{"1", "2", "3"}},
				{Label: "Contact", Type: "text", Options: nil},
			},
		},
	}

	for _, test := range tests {
		test.action = "edit"
		test.pageMode = editMode
		test.scrape = scrapeEditForm
		runTest(test, t)
	}
}
func TestEditFormActions(t *testing.T) {
	tests := []test{
		{
			name:   "Add form item",
			action: "add0",
			formItems: []models.FormItem{
				{Label: "Order", Type: "text", Options: nil},
				{Label: "Contact", Type: "text", Options: nil},
			},
			expected: []models.FormItem{
				{Label: "Order", Type: "text", Options: nil},
				{Label: "", Type: "text", Options: nil},
				{Label: "Contact", Type: "text", Options: nil},
			},
		},
		{
			name:   "Delete form item",
			action: "del1",
			formItems: []models.FormItem{
				{Label: "Order", Type: "text", Options: nil},
				{Label: "", Type: "text", Options: nil},
				{Label: "Contact", Type: "text", Options: nil},
			},
			expected: []models.FormItem{
				{Label: "Order", Type: "text", Options: nil},
				{Label: "Contact", Type: "text", Options: nil},
			},
		},
		{
			name:   "Move form item up",
			action: "upp1",
			formItems: []models.FormItem{
				{Label: "1", Type: "text", Options: nil},
				{Label: "2", Type: "text", Options: nil},
				{Label: "3", Type: "text", Options: nil},
			},
			expected: []models.FormItem{
				{Label: "2", Type: "text", Options: nil},
				{Label: "1", Type: "text", Options: nil},
				{Label: "3", Type: "text", Options: nil},
			},
		},
		{
			name:   "Move form item down",
			action: "dwn1",
			formItems: []models.FormItem{
				{Label: "1", Type: "text", Options: nil},
				{Label: "2", Type: "text", Options: nil},
				{Label: "3", Type: "text", Options: nil},
			},
			expected: []models.FormItem{
				{Label: "1", Type: "text", Options: nil},
				{Label: "3", Type: "text", Options: nil},
				{Label: "2", Type: "text", Options: nil},
			},
		},
		{
			name:   "Change form item type txt/select",
			action: "sel0",
			formItems: []models.FormItem{
				{Label: "1", Type: "text", Options: nil},
			},
			expected: []models.FormItem{
				{Label: "1", Type: "select", Options: []string{""}},
			},
		},
		{
			name:   "Change form item type select/cxb",
			action: "cxb0",
			formItems: []models.FormItem{
				{Label: "1", Type: "select", Options: []string{""}},
			},
			expected: []models.FormItem{
				{Label: "1", Type: "checkbox", Options: nil},
			},
		},
		{
			name:   "Add select option item",
			action: "opt0 add0",
			formItems: []models.FormItem{
				{Label: "Order", Type: "select", Options: []string{"1", "2"}},
			},
			expected: []models.FormItem{
				{Label: "Order", Type: "select", Options: []string{"1", "", "2"}},
			},
		},
		{
			name:   "Delete select option item",
			action: "opt0 del1",
			formItems: []models.FormItem{
				{Label: "Order", Type: "select", Options: []string{"1", "2", "3"}},
			},
			expected: []models.FormItem{
				{Label: "Order", Type: "select", Options: []string{"1", "3"}},
			},
		},
		{
			name:   "Move select option item up",
			action: "opt0 upp1",
			formItems: []models.FormItem{
				{Label: "Order", Type: "select", Options: []string{"1", "2", "3"}},
			},
			expected: []models.FormItem{
				{Label: "Order", Type: "select", Options: []string{"2", "1", "3"}},
			},
		},
		{
			name:   "Move select option item down",
			action: "opt0 dwn1",
			formItems: []models.FormItem{
				{Label: "Order", Type: "select", Options: []string{"1", "2", "3"}},
			},
			expected: []models.FormItem{
				{Label: "Order", Type: "select", Options: []string{"1", "3", "2"}},
			},
		},
	}

	for _, test := range tests {
		test.pageMode = editMode
		test.scrape = scrapeEditForm
		runTest(test, t)
	}
}

func runTest(test test, t *testing.T) {
	t.Run(test.name, func(t *testing.T) {
		data := formPage{Title: test.title, FormItems: test.formItems, PageMode: test.pageMode}
		expected := formPage{Title: test.title, FormItems: test.expected, PageMode: test.pageMode}
		if test.action == "view" || test.action == "edit" {
			expected = data
		}

		w := httptest.NewRecorder()
		r, err := http.NewRequest("POST", "/1", makePostBody(data, test.action))
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.WithValue(r.Context(), contextKey("user"), models.User{ID: 0, Name: ""})
		r = r.WithContext(ctx)
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
