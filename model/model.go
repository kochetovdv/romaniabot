package model

type OrderFile struct {
	ID       string
	Date     string
	URL      string
	Filename string
	Name     string
}

type Order struct {
	ID        string
	FileID    string
	Year      string
	Number    string
	Formatted string
}