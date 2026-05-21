package model

type Repository struct {
	Owner string
	Name  string
	Exists bool
	URL   string
}

type Issue struct {
	Number   int    `json:"number"`
	HTMLURL  string `json:"html_url"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	State    string `json:"state"`
}