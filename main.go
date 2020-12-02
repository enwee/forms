package main

import "net/http"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Forms"))
	})
	http.ListenAndServe(":5000", nil)
}
