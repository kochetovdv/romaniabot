package main

import (
	"bytes"

	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"log/slog"

	"romaniabot/model"
	"romaniabot/pkg/extractors"
	"romaniabot/pkg/fileutil"

	"database/sql"
	_ "modernc.org/sqlite"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	url        = "https://cetatenie.just.ro/ordine-articolul-1-1/"
	outputFile = "output.txt"
	ordersPath = "orders/"
	allowedApp = "application/pdf" // check it out
	liTags     []string
	OrderFiles []model.OrderFile
)

const (
	createDB = `CREATE TABLE IF NOT EXISTS OrderFile (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		Date TEXT,
		URL TEXT UNIQUE,
		Filename TEXT UNIQUE,
		Name TEXT,
		IsURLBroken BOOLEAN,
		IsDownloaded BOOLEAN DEFAULT FALSE,
		CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
		UpdatedAt DATETIME DEFAULT CURRENT_TIMESTAMP);`
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

	_, err = db.Exec(createDB)
	if err != nil {
		slog.Error("Database creating error: %e", err)
	}

	// Request to URL
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Error("Error during request: %e\n", err)
	}

	// Set user-agent for https
	req.Header.Set("User-Agent", "RomanianBot/1.0")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error during connect to %s: %e\n", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error during reading body response: %e\n", err)
		return
	}

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
	temp, err := extractors.OrderFiles(liTags)
	if err != nil {
		log.Printf("Error during extracting order files: %e\n", err)
		return
	}

	// Saving to file
	if err := fileutil.WriteToFile(outputFile, liTags); err != nil {
		log.Printf("Error during writing outputfile %s: %e\n", outputFile, err)
		return
	}
	fmt.Println(temp)
}
