package model

type Repository struct {
	Owner string
	Name  string
	Exists bool
	URL   string
}