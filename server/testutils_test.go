package main

import (
	"io"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// makePostBody strings together a form request body e.g key=value&key=value&....
func makePostBody(data editFormPage, action string) io.Reader {
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

// getNextToken advances to next token and returns the Token
func getNextToken(z *html.Tokenizer) html.Token {
	z.Next()
	return z.Token()
}

// getAttr returns the value of given attibuite in the Token
func getAttr(t html.Token, key string) string {
	for _, attr := range t.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return "[" + key + "] not found"
}

// checkEditMode returns the editMode based on state of Edit/Preview buttons
func checkEditMode(t html.Token, editMode bool) bool {
	m := map[string]string{}
	for _, attr := range t.Attr {
		m[attr.Key] = attr.Val
	}
	_, disabled := m["disabled"]
	if m["name"] == "action" && m["value"] == "view" {
		if !disabled {
			editMode = true
		}
	}
	if m["name"] == "action" && m["value"] == "edit" {
		if disabled {
			editMode = true
		}
	}
	return editMode
}

// scrapeViewForm screen scrapes the response Body and returns a form struct for test comparison
func scrapeViewForm(body io.Reader) editFormPage {
	scraped := editFormPage{}
	processedOptions := false
	t := html.Token{}
	z := html.NewTokenizer(body)
	// this extracts the only the Edit/Preview button state and form Title
	// dont need the loop to to keep checking for this after initial
	for {
		t = getNextToken(z)
		if z.Err() != nil {
			break
		}
		if t.Type == html.StartTagToken && t.Data == "button" {
			scraped.EditMode = checkEditMode(t, scraped.EditMode)
			continue
		}
		if t.Type == html.StartTagToken && t.Data == "h1" {
			// go into text inside <h1> or if <h1></h1>
			t = getNextToken(z)
			if t.Type == html.EndTagToken && t.Data == "h1" {
				scraped.Title = ""
				break
			}
			scraped.Title = t.Data
			break
		}
	}
	// this extracts the rest of the form items
	// label, input type, and options if <select>
	for {
		// empty <label></label> of select input type has hidden options
		// <input type=hidden name=optionsX> are scrapped until <label> tag
		// if previous iteration processedOptions, dont advance to next Token
		// and in this case process last/current Token
		if !processedOptions {
			t = getNextToken(z)
			if z.Err() != nil {
				break
			}
		}
		processedOptions = false // reset flag for current iteration

		if t.Type == html.StartTagToken && t.Data == "label" {
			fI := formItem{}
			// go into text inside <label> or if <label></label>
			t = getNextToken(z)
			if t.Type == html.EndTagToken && t.Data == "label" { // if empty
				for {
					t = getNextToken(z)
					if z.Err() != nil {
						break
					}
					// next token must be <input type="hidden" name="type">
					inputTag := t.Type == html.StartTagToken && t.Data == "input"
					attrs := getAttr(t, "type") == "hidden" && getAttr(t, "name") == "type"
					if inputTag && attrs {
						fI.Type = getAttr(t, "value")
						break
					}
				}
				if fI.Type == "select" {
					for {
						t = getNextToken(z)
						if z.Err() != nil {
							break
						}
						// next token must be <input type="hidden" name="optionsX">
						inputTag := t.Type == html.StartTagToken && t.Data == "input"
						hidden := getAttr(t, "type") == "hidden"
						optionX := getAttr(t, "name") == "options"+strconv.Itoa(len(scraped.FormItems))
						if inputTag && hidden && optionX {
							fI.Options = append(fI.Options, getAttr(t, "value"))
						} else if t.Type == html.StartTagToken && t.Data == "label" {
							// next iteration dont advance to next Token
							// process current Token on next iteration
							processedOptions = true
							break
						}
					}
				}
				scraped.FormItems = append(scraped.FormItems, fI)
				continue
			}
			fI.Label = t.Data
			// look for next <select> or <input> for input type
			for {
				t = getNextToken(z)
				if z.Err() != nil {
					break
				}
				if t.Type == html.StartTagToken && (t.Data == "select" || t.Data == "input") {
					break
				}
			}
			// <select> is followed by <option>s
			if t.Data == "select" {
				fI.Type = t.Data
				for {
					t = getNextToken(z)
					if z.Err() != nil {
						break
					}
					if t.Type == html.StartTagToken && t.Data == "option" {
						t = getNextToken(z) // go into text inside <option> or if empty
						if t.Type == html.EndTagToken && t.Data == "option" {
							fI.Options = append(fI.Options, "") // empty case
							continue
						}
						fI.Options = append(fI.Options, t.Data)
					} else if t.Type == html.EndTagToken && t.Data == "select" {
						break
					}
				}
			}
			// text or checkbox type
			if t.Data == "input" {
				fI.Type = getAttr(t, "type")
			}
			scraped.FormItems = append(scraped.FormItems, fI)
		}
	}
	return scraped
}

func scrapeEditForm(body io.Reader) editFormPage {
	scraped := editFormPage{}
	processedOptions := false
	t := html.Token{}
	z := html.NewTokenizer(body)
	// this extracts the only the Edit/Preview button state and form Title
	// dont need the loop to to keep checking for this after initial
	for {
		t = getNextToken(z)
		if z.Err() != nil {
			break
		}
		if t.Type == html.StartTagToken && t.Data == "button" {
			scraped.EditMode = checkEditMode(t, scraped.EditMode)
			continue
		}
		if t.Type == html.StartTagToken && t.Data == "h1" {
			for {
				t = getNextToken(z)
				if z.Err() != nil {
					break
				}
				if t.Type == html.StartTagToken && t.Data == "input" {
					scraped.Title = getAttr(t, "value")
				} else if t.Type == html.EndTagToken && t.Data == "h1" {
					break
				}
			}
			break
		}
	}
	// this extracts the rest of the form items
	// label, input type, and options if <select>
	for {
		// options are scrapped until next <input name!=optionsX>
		// if previous iteration processedOptions, dont advance to next Token
		// and in this case process last/current Token
		if !processedOptions {
			t = getNextToken(z)
			if z.Err() != nil {
				break
			}
		}
		processedOptions = false // reset flag for current iteration

		if t.Type == html.StartTagToken && t.Data == "input" {
			fI := formItem{}
			// look for next <input name="label">
			if getAttr(t, "name") != "label" {
				continue
			}
			fI.Label = getAttr(t, "value")
			// look for next <input name="type">
			for {
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
			// <input name="type" value="select">
			if fI.Type == "select" {
				fI.Options = []string{}
				for { // look for next <input name=optionsX>
					t = getNextToken(z)
					if z.Err() != nil {
						break
					}
					if t.Type == html.StartTagToken && t.Data == "input" {
						if getAttr(t, "name") == "options"+strconv.Itoa(len(scraped.FormItems)) {
							fI.Options = append(fI.Options, getAttr(t, "value"))
						} else {
							// next iteration dont advance to next Token
							// process current Token on next iteration
							processedOptions = true
							break
						}
					}
				}
			}
			scraped.FormItems = append(scraped.FormItems, fI)
		}
	}
	return scraped
}

type mock struct{}

func (m mock) get(id int) (title string, formItems []formItem, found bool, err error) {
	return
}

func (m mock) update(id int, title string, formItems []formItem) error {
	return nil
}

func (m mock) getAll() (forms []chooseFormPageItem, err error) {
	return
}

func (m mock) new() (id int, err error) {
	return
}

func (m mock) delete(id int) error {
	return nil
}
