package main

import (
	"net/http"
	"testing"
)

func TestMainServer(t *testing.T) {
	// Makes e.g 150(numRequests) concurrent requests FOR e.g 10(runs) times
	// default install of mysqld has 151 max DB connections
	// heroku clearDB mysql has about 10 max DB connections
	// test hits max DB connection limit before server limit
	const url = "http://localhost:5000"
	const numRequests = 150 // default mysqld max conn limit
	const runs = 10         // for tests to complete faster, can be 20

	// On my machine usually more than 20+ runs will hit server too many open files error
	// But the server does not crash, errors are logged and next request works
	// That's successful handling of 3000 (20 sequential, 150 concurrent) requests in short time
	// its way more than expected and designed-for use case. so GO is good.
	for i := 0; i < runs; i++ {
		errs := makeTestRequests(url, numRequests)
		if errs != 0 {
			t.Errorf("%d/%d responses NOT 200 OK", errs, numRequests)
		}
	}
}

func makeTestRequests(url string, numRequests int) (errs int) {
	c := make(chan int, numRequests)
	result := []int{}
	for i := 0; i < numRequests; i++ {
		go func() {
			r, _ := http.DefaultClient.Get(url)
			c <- r.StatusCode
		}()
	}
	for i := 0; i < numRequests; i++ {
		if code := <-c; code != 200 {
			result = append(result, code)
		}
	}
	return len(result)
}
