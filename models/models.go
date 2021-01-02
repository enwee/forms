package models

// Form represents the editable form
type Form struct {
	ID        int
	Title     string
	FormItems []FormItem
	Updated   string
	UserID    int
}

// FormItem is a HTML input type item e.g. <input type='textbox'>
type FormItem struct {
	Label   string
	Type    string
	Options []string
}

// User data
type User struct {
	ID      int
	Name    string
	Pwhash  string
	Created string
}

// PostResponse is the data when a users submits a form
type PostResponse struct {
	ID         int
	FormID     int
	Version    string
	Title      string
	FormKeys   []string
	FormValues []string
}

// Response is the user submission data for that version of the form
type Response struct {
	ID      int
	Version string
	Data    []string
}

// ResponseSet is all the user responses to a version of the form
type ResponseSet struct {
	Title       string
	Version     string
	TableHeader []string
	TableData   []Response
}
