package main

import (
	"io"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

func makeBody(data form, action string) io.Reader {
	body := "action=" + action + "&title=" + data.Title
	for index, formItem := range data.FormItems {
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

func getNextToken(z *html.Tokenizer) html.Token {
	z.Next()
	return z.Token()
}

func getAttr(t html.Token, key string) string {
	for _, attr := range t.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return "[" + key + "] not found"
}

func checkEditMode(t html.Token) bool {
	m := map[string]string{}
	for _, attr := range t.Attr {
		m[attr.Key] = attr.Val
	}

	_, disabled := m["disabled"]
	nameAction := m["name"] == "action"
	editDisabled := m["value"] == "edit" && disabled
	previewDisabled := m["value"] == "view" && disabled

	if nameAction && editDisabled {
		return true
	}
	if nameAction && !previewDisabled {
		return true
	}
	return false
}

func scrapPreviewBody(body io.Reader) form {
	scrapped := form{}
	z := html.NewTokenizer(body)
	for {
		t := getNextToken(z)
		if z.Err() != nil {
			break
		}
		if t.Type == html.StartTagToken && t.Data == "button" {
			scrapped.EditMode = checkEditMode(t)
			continue
		}
		if t.Type == html.StartTagToken && t.Data == "h1" {
			t = getNextToken(z) // go into text inside <h1>
			scrapped.Title = t.Data
			break
		}
	}

	for {
		t := getNextToken(z)
		if z.Err() != nil {
			break
		}
		if t.Type == html.StartTagToken && t.Data == "label" {
			fI := formItem{}

			t = getNextToken(z) // go into text inside <label> or if empty
			if t.Type == html.EndTagToken && t.Data == "label" {
				fI.Type = "text"
				scrapped.FormItems = append(scrapped.FormItems, fI)
				continue
			}
			fI.Label = t.Data

			for { // scan to next <select> or <input>
				t = getNextToken(z)
				if z.Err() != nil {
					break
				}
				if t.Type == html.StartTagToken && (t.Data == "select" || t.Data == "input") {
					break
				}
			}

			if t.Data == "select" {
				fI.Type = t.Data
				for {
					t = getNextToken(z)
					if z.Err() != nil {
						break
					}
					if t.Type == html.StartTagToken && t.Data == "option" {
						t = getNextToken(z) // go into text inside <option>
						fI.Options = append(fI.Options, t.Data)
					} else if t.Type == html.EndTagToken && t.Data == "select" {
						break
					}
				}
			}

			if t.Data == "input" {
				fI.Type = getAttr(t, "type")
			}

			scrapped.FormItems = append(scrapped.FormItems, fI)
		}
	}
	return scrapped
}

func scrapEditBody(body io.Reader) form {
	scrapped := form{}
	didOptions := false
	t := html.Token{}
	z := html.NewTokenizer(body)
	for {
		t = getNextToken(z)
		if z.Err() != nil {
			break
		}
		if t.Type == html.StartTagToken && t.Data == "button" {
			scrapped.EditMode = checkEditMode(t)
			continue
		}
		if t.Type == html.StartTagToken && t.Data == "h1" {
			for {
				t = getNextToken(z)
				if z.Err() != nil {
					break
				}
				if t.Type == html.StartTagToken && t.Data == "input" {
					scrapped.Title = getAttr(t, "value")
				} else if t.Type == html.EndTagToken && t.Data == "h1" {
					break
				}
			}
			break
		}
	}

	for {
		if !didOptions {
			t = getNextToken(z)
			if z.Err() != nil {
				break
			}
		}
		didOptions = false // cant think of a way to reverse z.Next

		if t.Type == html.StartTagToken && t.Data == "input" {
			fI := formItem{}

			if getAttr(t, "name") != "label" {
				continue
			}
			fI.Label = getAttr(t, "value")

			for { // scan to next <input>
				t = getNextToken(z)
				if z.Err() != nil {
					break
				}
				if t.Type == html.StartTagToken && t.Data == "input" {
					if getAttr(t, "name") == "type" {
						fI.Type = getAttr(t, "value")
						break
					}
				}
			}

			if fI.Type == "select" {
				fI.Options = []string{}
				for { // scan to next <input>
					t = getNextToken(z)
					if z.Err() != nil {
						break
					}
					if t.Type == html.StartTagToken && t.Data == "input" {
						if getAttr(t, "name") == "options"+strconv.Itoa(len(scrapped.FormItems)) {
							fI.Options = append(fI.Options, getAttr(t, "value"))
						} else {
							didOptions = true // have when to next wanted token
							break
						}
					}
				}
			}
			scrapped.FormItems = append(scrapped.FormItems, fI)
		}
	}
	return scrapped
}
