package models

// Form data from other sources e.g. DB, html page
// have to be parsed into this stuct format.
type Form struct {
	ID        int
	Title     string
	FormItems []FormItem
	Updated   string
	UserID    int
}

// FormItem from other sources e.g. json, html page
// have to be parsed into this stuct format.
type FormItem struct {
	Label   string
	Type    string
	Options []string
}

// User data from other sources e.g. DB, html page
// have to be parsed into this stuct format.
type User struct {
	ID      int
	Name    string
	Pwhash  string
	Created string
}
