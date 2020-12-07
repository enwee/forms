package main

import (
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/net/html"
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

func TestMakeFormPreview(t *testing.T) {
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
	view := form{"Lam's BBQ Order Form", formItems, false}

	w := httptest.NewRecorder()
	r, err := http.NewRequest("POST", "/", postBody(view, "view"))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	app.makeForm(w, r)
	resp := w.Result()
	scrapped := scrapResp(resp.Body)
	if !reflect.DeepEqual(view, scrapped) {
		t.Errorf("\nrequest:\n%+v\nscrapped:\n%+v", view, scrapped)
	}
}

func postBody(view form, action string) io.Reader {
	body := "action=" + action + "&title=" + view.Title
	for index, formItem := range view.FormItems {
		body = body + "&label=" + formItem.Label
		if formItem.Type == "select" {
			body = body + "&type=select"
			for _, option := range formItem.Options {
				body = body + "&options" + strconv.Itoa(index) + "=" + option
			}
		} else {
			body = body + "&type=" + formItem.Type
		}
	}
	return strings.NewReader(body)
}

func scrapResp(body io.Reader) form {
	scrapped := form{}
	z := html.NewTokenizer(body)
	for {
		z.Next()
		if z.Err() != nil {
			break
		}

		t := z.Token()
		if t.Type == html.EndTagToken && t.Data == "form" {
			break
		}

		if t.Type == html.StartTagToken && t.Data == "button" {
			m := map[string]string{}
			for _, attr := range t.Attr {
				m[attr.Key] = attr.Val
			}
			_, disabled := m["disabled"]
			if m["name"] == "action" {
				if m["value"] == "edit" && disabled {
					scrapped.EditMode = true
				}
				if m["value"] == "view" && !disabled {
					scrapped.EditMode = true
				}
			}
			continue
		}

		if t.Type == html.StartTagToken && t.Data == "h1" {
			z.Next() // go into text inside <h1>
			scrapped.Title = string(z.Text())
			continue
		}

		if t.Type == html.StartTagToken && t.Data == "label" {
			fI := formItem{}

			z.Next() // go into text inside <label> or if empty
			t = z.Token()
			if t.Type == html.EndTagToken && t.Data == "label" {
				fI.Type = "text"
				scrapped.FormItems = append(scrapped.FormItems, fI)
				continue
			}
			fI.Label = t.Data

			for { // scan to next <select> or <input>
				z.Next()
				t = z.Token()
				if t.Type == html.StartTagToken && (t.Data == "select" || t.Data == "input") {
					break
				}
			}

			if t.Data == "select" {
				fI.Type = t.Data
				for {
					z.Next()
					t = z.Token()
					if t.Type == html.StartTagToken && t.Data == "option" {
						z.Next() // go into text inside <option>
						fI.Options = append(fI.Options, string(z.Text()))
					} else if t.Type == html.EndTagToken && t.Data == "select" {
						break
					}
				}
			}

			if t.Data == "input" { // get the type attribuite of <input>
				for _, attr := range t.Attr {
					if attr.Key == "type" {
						fI.Type = attr.Val
					}
				}
			}

			scrapped.FormItems = append(scrapped.FormItems, fI)
		}
	}
	return scrapped
}
