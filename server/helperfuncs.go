package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

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

func validateForm(r *http.Request) (formItems []formItem, action, opt string, index, idx int, err error) {
	labels := r.Form["label"] // will get []string(nil) if doesnt exist
	inputType := r.Form["type"]
	if len(labels) != len(inputType) { // len([]string(nil) will be  0)
		err = fmt.Errorf("number of labels, types not equal")
		return
	}
	for i, label := range labels { // range []string(nil) is ok doesnt panic
		var options []string
		if inputType[i] == "select" {
			opts := r.Form["options"+strconv.Itoa(i)]
			for _, option := range opts {
				options = append(options, option)
			}
		}
		formItems = append(formItems, formItem{label, inputType[i], options})
	}

	action = r.FormValue("action")
	if action != "edit" && action != "view" {
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
