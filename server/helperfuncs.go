package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"forms/models"
)

const maxUsernameLen = 8
const maxFormTitleLen = 50

func getAction(action string) (string, int, error) {
	index, err := strconv.Atoi(action[3:])
	if err != nil {
		return "", 0, err
	}
	return action[:3], index, nil
}

//for template.FuncMap
func minus1(x int) int {
	return x - 1
}

func validateUsername(username string) (err string) {
	if username == "" {
		return "user name cannot be blank"
	}
	if len(username) > maxUsernameLen {
		return "user name too long (max 8 characters)"
	}
	return ""
}

func validateTitle(r *http.Request) (title, feedback string) {
	title = strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		feedback = "Title cannot be empty"
	}
	if utf8.RuneCount([]byte(title)) > maxFormTitleLen {
		feedback = "Title is too long"
	}
	return
}

func validateForm(r *http.Request, re *regexp.Regexp) (formItems []models.FormItem, action, opt string, index, idx int, err error) {
	labels := r.Form["label"] // will get []string(nil) if doesnt exist
	inputType := r.Form["type"]
	if len(labels) != len(inputType) { // len([]string(nil) will be  0)
		err = fmt.Errorf("number of labels, types not equal")
		return
	}
	for i, label := range labels { // range []string(nil) is ok doesnt panic
		var options []string
		if !stringIs(inputType[i], "text", "checkbox", "select") {
			err = fmt.Errorf("[%s] invalid input type: [%s]", label, inputType[i])
			return
		}

		if inputType[i] == "select" {
			opts := r.Form["options"+strconv.Itoa(i)]
			for _, option := range opts {
				options = append(options, strings.TrimSpace(option))
			}
		}
		label = strings.TrimSpace(label)
		formItems = append(formItems, models.FormItem{Label: label, Type: inputType[i], Options: options})
	}

	action = r.FormValue("action")
	if !stringIs(action, "edit", "view", "choose", "auth") {
		if !re.MatchString(action) {
			err = fmt.Errorf("[%s] invalid action", action)
			return
		}
		actions := strings.Split(r.FormValue("action"), " ")
		action, index, err = getAction(actions[0])
		if index > len(formItems)-1 {
			err = fmt.Errorf("%s invalid action", actions)
			return
		}
		if action == "opt" {
			opt, idx, err = getAction(actions[1])
			if idx > len(formItems[index].Options)-1 {
				err = fmt.Errorf("%s invalid action", actions)
				return
			}
		}
	}
	return
}

func stringIs(input string, ss ...string) bool {
	for _, s := range ss {
		if input == s {
			return true
		}
	}
	return false
}
