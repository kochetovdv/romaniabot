package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"time"

	"log/slog"

	"romaniabot/model"
	"romaniabot/pkg/downloaders"
	"romaniabot/pkg/extractors"
	_ "romaniabot/pkg/fileutil"
	"romaniabot/pkg/web"

	"database/sql"

	_ "modernc.org/sqlite"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	liTags     []string
	OrderFiles []model.OrderFile
)

const (
	url        = "https://cetatenie.just.ro/ordine-articolul-1-1/"
	outputFile = "output.txt"
	ordersPath = "orders/"
	allowedApp = "application/pdf"
)

func main() {
	// LOGGER init
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	// DATABASE init
	db, err := sql.Open("sqlite", "./orders.db")
	//Check for any error
	if err != nil {
		slog.Error("Database initializing error: %e", err)
	}
	defer db.Close()

	// Creating model for OrderFiles in DB
	_, err = db.Exec(model.CreateOrderFilesDB)
	if err != nil {
		slog.Error("Database table for OrderFiles creating error: %e", err)
	}

	// Creating model for Orders in DB
	_, err = db.Exec(model.CreateOrdersDB)
	if err != nil {
		slog.Error("Database table for Orders creating error: %e", err)
	}

	// Request to URL
	body, err := web.GetResponseBody(url)
	if err != nil {
		log.Printf("Error during reading body response: %e\n", err)
		return
	}

	// Extracting all <li> tags
	reader := bytes.NewReader(body)
	z := html.NewTokenizer(reader)

	for {
		tt := z.Next()
		cancel := false
		if tt == html.ErrorToken {
			err := z.Err()
			slog.Error("Error in main (html.ErrorToken):", err)
			break
		}
		if tt == html.StartTagToken && z.Token().DataAtom == atom.Li {
			tag, err := extractors.InsideTags(z, atom.Li)
			if err != nil {
				log.Printf("Error during extracting <li>: %e\n", err)
				return
			}
			if tag != "" {
				liTags = append(liTags, tag)
			}

		}
		if cancel {
			break
		}
	}

	// Extracting target model
	orderFiles, err := extractors.OrderFiles(liTags)
	if err != nil {
		log.Printf("Error during extracting order files: %e\n", err)
		return
	}

	// Saving order files to DB
	statement, err := db.Prepare(model.Insert_Order_File)
	if err != nil {
		log.Fatal(err)
	}
	defer statement.Close()

	// Выполнение запроса с конкретными параметрами
	for _, el := range orderFiles {
		_, err := statement.Exec(el.Date, el.URL, el.Filename, el.Name)
		if err != nil {
			log.Printf("Error during insert in db %v: %e\n", el, err)
		}
	}

	// TODO: отдельно запрос filename для FileCheck, отдельно url для BrokenUrls, и после уже данный FileUrls
	FilesToDownloadCheck(db)
	URLsToCheck(db)
	Download(db)

	// TODO: download order files


	// TODO: IsExtension
	// TODO: битые ссылки. Пусть возвращается мапа элементов, которые не прочитались, далее по ним пройдемся

	//TODO: if files not parsed, parse
	//TODO: import result to DB
	//TODO: handles for bot
	//TODO: TG-bot

}

func FilesToDownloadCheck(db *sql.DB) {
	// Storage for FileNames to download
	var filesToDownload []string

	// read from DB existing orderfiles
	rows, err := db.Query(model.Get_new_Filenames)
	if err != nil {
		log.Printf("Error during reading filesToDownload from db: %e\n", err)
	}
	defer rows.Close()

	for rows.Next() {
		var filename string

		err = rows.Scan(&filename)
		if err != nil {
			log.Printf("Error during scaning filesToDownload row from db:%s\t%e\n", filename, err)

		}
		filesToDownload = append(filesToDownload, filename)
	}

	fmt.Println("Total files to download from DB: ", len(filesToDownload))
	downloadedFiles := downloaders.CheckDownloadedFiles(ordersPath, filesToDownload)
	fmt.Println("Total downloaded files after checking folder: ", len(downloadedFiles))

	if len(downloadedFiles) > 0 {
		// Обновляем информацию в БД
		statement, err := db.Prepare(model.Set_is_Downloaded)
		if err != nil {
			log.Fatal(err)
		}
		defer statement.Close()

		// Выполнение запроса с конкретными параметрами
		for _, el := range downloadedFiles {
			_, err := statement.Exec(el)
			if err != nil {
				log.Printf("Error during update in db %v: %e\n", el, err)
			}
		}
	}
}

func URLsToCheck(db *sql.DB) {
	// Storage for FileNames to download
	var urlsToCheck []string

	// read from DB existing orderfiles
	rows, err := db.Query(model.Get_Valid_URLs)
	if err != nil {
		log.Printf("Error during reading URLs to check from db: %e\n", err)
	}
	defer rows.Close()

	for rows.Next() {
		var url string

		err = rows.Scan(&url)
		if err != nil {
			log.Printf("Error during scaning URLs row from db:%s\t%e\n", url, err)

		}
		urlsToCheck = append(urlsToCheck, url)
	}

	fmt.Println("Total URLs to check from DB: ", len(urlsToCheck))
	brokenURLs := downloaders.CheckBrokenURLs(urlsToCheck, 2, time.Second*20)
	fmt.Println("Total broken URLs after ping: ", len(brokenURLs))
	if len(brokenURLs) > 0 {
		// Обновляем информацию в БД
		statement, err := db.Prepare(model.Set_broken_URLs)
		if err != nil {
			log.Fatal(err)
		}
		defer statement.Close()

		// Выполнение запроса с конкретными параметрами
		for _, el := range brokenURLs {
			_, err := statement.Exec(el)
			if err != nil {
				log.Printf("Error during update in db %v: %e\n", el, err)
			}
		}
	}
}

func Download(db *sql.DB) {
	// Storage for FileNames to download
	filesToDownload := make(map[string]string)

	// read from DB existing orderfiles
	rows, err := db.Query(model.Get_Files_to_download)
	if err != nil {
		log.Printf("Error during reading FileURLs to check from db: %e\n", err)
	}
	defer rows.Close()

	for rows.Next() {
		var filename string
		var url string

		err = rows.Scan(&url, &filename)
		if err != nil {
			log.Printf("Error during scaning URLs row from db:%s\t%s\t%e\n", url, filename, err)

		}
		filesToDownload[filename] = url
	}

	fmt.Println("Total Files to download from DB: ", len(filesToDownload))
	downloaders.Downloader(ordersPath, filesToDownload)

	// Обновляем информацию в БД
	statement, err := db.Prepare(model.Set_is_Downloaded)
	if err != nil {
		log.Fatal(err)
	}
	defer statement.Close()

	// Выполнение запроса с конкретными параметрами
	for fname := range filesToDownload {
		_, err := statement.Exec(fname)
		if err != nil {
			log.Printf("Error during update in db %v: %e\n", fname, err)
		}
	}
}
