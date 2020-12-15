package models

// Form comment
type Form struct {
	ID        int
	Title     string
	FormItems []FormItem
	Updated   string
	UserID    int
}

// FormItem comment
type FormItem struct {
	Label   string
	Type    string
	Options []string
}

// User comment
type User struct {
	ID      int
	Name    string
	Pwhash  string
	Created string
}
