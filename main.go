package main

import (
	"bytes"

	_ "fmt"
	"log"
	"os"

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
	statement, err := db.Prepare(model.OrderFileToDB)
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

	// Storage for URLs and files to download
	filesToDownload := make(map[string]string)

	// read from DB existing orderfiles
	rows, err := db.Query(model.FilesToDownload)
	if err != nil {
		log.Printf("Error during reading filesToDownload from db: %e\n", err)
	}
	defer rows.Close()

	for rows.Next() {
		var filename string
		var url string

		err = rows.Scan(&url, &filename)
		if err != nil {
			log.Printf("Error during scaning filesToDownload row from db:%s\t%s\t%e\n", filename, url, err)

		}
		filesToDownload[filename] = url
	}

	// Saving to file
	// var f []string
	// for k, v := range filesToDownload {
	// 	f = append(f, k+"\t"+v)
	// }
	// if err := fileutil.WriteToFileStrings(outputFile, f); err != nil {
	// 	log.Printf("Error during writing outputfile %s: %e\n", outputFile, err)
	// 	return
	// }

	// download files
	downloaders.Downloader(ordersPath, filesToDownload)

	//TODO: if files not parsed, parse
	//TODO: import result to DB
	//TODO: handles for bot
	//TODO: TG-bot

}
