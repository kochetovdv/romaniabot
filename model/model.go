package model

import "time"

type OrderFile struct {
	Date         string    `json:"date"`
	URL          string    `json:"url"`
	Filename     string    `json:"filename"`
	Name         string    `json:"name"`
	IsURLBroken  bool      `json:"isURLBroken"`
	IsDownloaded bool      `json:"isDownloaded"`
	IsParsed     bool      `json:"isParsed"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Order struct {
	Filename          string       `json:"fileid"`
	Year              uint    `json:"year"`
	Number            uint    `json:"number"`
	FullNameFormatted string    `json:"fullnameformatted"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// type FilesToDownload struct{
// 	URL          string    `json:"url"`
// 	Filename     string    `json:"filename"`

// }
