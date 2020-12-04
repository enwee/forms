package main

import "strconv"

//for template.FuncMap
func minus1(x int) int {
	return x - 1
}

func getAction(action string) (string, int, error) {
	index, err := strconv.Atoi(action[3:])
	if err != nil {
		return "", 0, err
	}
	return action[:3], index, nil
}
