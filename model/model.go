package model

import "time"

type OrderFile struct {
	ID           int       `json:"id"`
	Date         string    `json:"date"`
	URL          string    `json:"url"`
	Filename     string    `json:"filename"`
	Name         string    `json:"name"`
	IsURLBroken  bool      `json:"isURLBroken"`
	IsDownloaded bool      `json:"isDownloaded"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Order struct {
	ID                int       `json:"id"`
	FileID            int       `json:"fileid"`
	Year              string    `json:"year"`
	Number            string    `json:"number"`
	FullNameFormatted string    `json:"fullnameformatted"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// type FilesToDownload struct{
// 	URL          string    `json:"url"`
// 	Filename     string    `json:"filename"`

// }