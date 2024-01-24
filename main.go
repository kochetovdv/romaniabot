package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"log/slog"

	"romaniabot/model"
	"romaniabot/pkg/downloaders"
	"romaniabot/pkg/extractors"
	"romaniabot/pkg/fileutil"
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
	url = "https://cetatenie.just.ro/ordine-articolul-1-1/" // TODO: slice of links
	//	outputFile = "output.txt"
	ordersPath = "orders/"
	allowedApp = "application/pdf"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	// Initialize database
	db, err := sql.Open("sqlite", "./orders.db")
	if err != nil {
		slog.Error("Database initializing error: %e", err)
		return
	}
	defer db.Close()

	// Create OrderFiles table in the database
	_, err = db.Exec(model.CreateOrderFilesDB)
	if err != nil {
		slog.Error("Database table for OrderFiles creating error: %e", err)
		return
	}

	// Create Orders table in the database
	_, err = db.Exec(model.CreateOrdersDB)
	if err != nil {
		slog.Error("Database table for Orders creating error: %e", err)
		return
	}

	// Get <li> tags from target URL
	// LiTagsExtractor(db)
	// // Check downloaded order files in folder
	// FilesToDownloadCheck(db)
	// // Check broken URLs
	// URLsToCheck(db)
	// // Download order files
	// Download(db)
	// // Parsing orders
	ParsePDF(db)

	// TODO: IsExtension - remove if unnecessary

	// TODO: if files not parsed, parse
	// TODO: import result to DB
	// TODO: handles for bot
	// TODO: TG-bot
}

func LiTagsExtractor(db *sql.DB) {
	// Request URL
	body, err := web.GetResponseBody(url)
	if err != nil {
		log.Printf("Error during reading body response: %e\n", err)
		return
	}

	// Extract <li> tags
	reader := bytes.NewReader(body)
	z := html.NewTokenizer(reader)

	for {
		tt := z.Next()
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
	}

	// Extract target model
	orderFiles, err := extractors.OrderFiles(liTags)
	if err != nil {
		log.Printf("Error during extracting order files: %e\n", err)
		return
	}

	// Save order files to DB
	statement, err := db.Prepare(model.Insert_Order_File)
	if err != nil {
		log.Fatal(err)
	}
	defer statement.Close()

	// Execute query with specific parameters
	for _, el := range orderFiles {
		_, err := statement.Exec(el.Date, el.URL, el.Filename, el.Name)
		if err != nil {
			log.Printf("Error during insert in db %v: %e\n", el, err)
		}
	}
}

// FilesToDownloadCheck checks the downloaded files and updates the database accordingly.
func FilesToDownloadCheck(db *sql.DB) {
	// Create a slice to store the filenames that need to be downloaded
	filesToDownload := make([]string, 0)

	// Query the database to get the new filenames
	rows, err := db.QueryContext(context.Background(), model.Get_new_Filenames)
	if err != nil {
		log.Printf("Error during reading filesToDownload from db: %e\n", err)
		return
	}
	defer rows.Close()

	// Iterate over the rows returned by the query
	for rows.Next() {
		var filename string

		// Scan the filename from the row
		err = rows.Scan(&filename)
		if err != nil {
			log.Printf("Error during scanning filesToDownload row from db: %s\t%e\n", filename, err)
			continue
		}

		// Append the filename to the filesToDownload slice
		filesToDownload = append(filesToDownload, filename)
	}

	// Print the total number of files to be downloaded
	fmt.Println("Total files to download from DB:", len(filesToDownload))

	// Check the downloaded files in the specified folder
	downloadedFiles := downloaders.CheckDownloadedFiles(ordersPath, filesToDownload)
	// Print the total number of downloaded files after checking the folder
	fmt.Println("Total downloaded files after checking folder:", len(downloadedFiles))

	// If there are downloaded files, update the database
	if len(downloadedFiles) > 0 {
		// Prepare the update statement
		statement, err := db.PrepareContext(context.Background(), model.Set_is_Downloaded)
		if err != nil {
			log.Fatal(err)
		}
		defer statement.Close()

		// Iterate over the downloaded files and execute the update statement
		for _, el := range downloadedFiles {
			_, err := statement.Exec(el)
			if err != nil {
				log.Printf("Error during update in db %v: %e\n", el, err)
				continue
			}
		}
	}
}

func URLsToCheck(db *sql.DB) {
	// Create an empty slice to store the URLs that need to be checked
	urlsToCheck := make([]string, 0)

	// Query the database to get the valid URLs
	rows, err := db.Query(model.Get_Valid_URLs)
	if err != nil {
		// Log an error message if there is an error querying the database
		log.Printf("Error during reading URLs to check from db: %e\n", err)
		return
	}
	defer rows.Close()

	// Iterate over the rows returned from the query
	for rows.Next() {
		var url string

		// Scan the value of the URL column into the 'url' variable
		err = rows.Scan(&url)
		if err != nil {
			// Log an error message if there is an error scanning the row
			log.Printf("Error during scanning URLs row from db: %s\t%e\n", url, err)
			continue
		}
		// Append the URL to the 'urlsToCheck' slice
		urlsToCheck = append(urlsToCheck, url)
	}

	// Print the total number of URLs to check
	fmt.Println("Total URLs to check from DB: ", len(urlsToCheck))

	// Call the 'CheckBrokenURLs' function to check the broken URLs
	brokenURLs := downloaders.CheckBrokenURLs(urlsToCheck, 2, time.Second*20)

	// Print the total number of broken URLs after pinging
	fmt.Println("Total broken URLs after ping: ", len(brokenURLs))

	// If there are no broken URLs, return
	if len(brokenURLs) == 0 {
		return
	}

	// Prepare a statement to update the broken URLs in the database
	statement, err := db.Prepare(model.Set_broken_URLs)
	if err != nil {
		log.Fatal(err)
	}
	defer statement.Close()

	// Iterate over the broken URLs and execute the update statement for each URL
	for _, el := range brokenURLs {
		_, err := statement.Exec(el)
		if err != nil {
			// Log an error message if there is an error updating the database
			log.Printf("Error during update in db %v: %e\n", el, err)
		}
	}
}

// Refactored Download function
func Download(db *sql.DB) {
	// Storage for FileNames to download
	filesToDownload := make(map[string]string)

	// Read from DB existing orderfiles
	rows, err := db.Query(model.Get_Files_to_download)
	if err != nil {
		log.Printf("Error during reading FileURLs to check from db: %e\n", err)
		return
	}
	defer rows.Close()

	// Iterate over the rows returned by the query
	for rows.Next() {
		var filename string
		var url string

		err = rows.Scan(&url, &filename)
		if err != nil {
			log.Printf("Error during scanning URLs row from db:%s\t%s\t%e\n", url, filename, err)
			continue
		}

		// Add the filename and url to the filesToDownload map
		filesToDownload[filename] = url
	}

	fmt.Println("Total Files to download from DB: ", len(filesToDownload))
	downloaders.Downloader(ordersPath, filesToDownload)

	// Update information in the database
	statement, err := db.Prepare(model.Set_is_Downloaded)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer statement.Close()

	// Execute the query with specific parameters
	for fname := range filesToDownload {
		_, err := statement.Exec(fname)
		if err != nil {
			log.Printf("Error during update in db %v: %e\n", fname, err)
			continue
		}
	}
}

// Parse PDF
func ParsePDF(db *sql.DB) {
	// Storage for FileNames which are parsed by data from DB
	parsedFiles := make([]string, 0)
	filesToParse := make([]string, 0)

	// Read from DB filenames
	rows, err := db.Query(model.Get_Files_not_parsed)
	if err != nil {
		log.Printf("Error during reading FileURLs to check from db: %e\n", err)
		return
	}
	defer rows.Close()

	// Iterate over the rows returned by the query
	for rows.Next() {
		var filename string

		err = rows.Scan(&filename)
		if err != nil {
			log.Printf("Error during scanning URLs row from db:%s\t%e\n", filename, err)
			continue
		}

		// Add the filename to the slice
		parsedFiles = append(parsedFiles, filename)
	}
	log.Println("Total Files parsed from DB: ", len(parsedFiles))

	// Get all downloaded files from folder
	filesInFolder := fileutil.GetFileListInFolder(ordersPath)
	log.Println("Total Files in folder: ", len(filesInFolder))

	// Get difference between downloaded and parsed
	for _, fileInFolder := range filesInFolder {
		isParsed := false
		for _, fileInDB := range parsedFiles {
			if fileInFolder == fileInDB {
				isParsed = true
				break
			}
		}
		if isParsed {
			filesToParse = append(filesToParse, fileInFolder)
		}

	}

	orders, err := extractors.Order(ordersPath, filesToParse...)
	if err != nil {
		log.Printf("error:%e\n", err)
	}

	fmt.Println(orders)

	// save to DB
	statement, err := db.Prepare(model.Insert_Order)
	if err != nil {
		log.Fatal(err)
	}
	defer statement.Close()

	// Execute query with specific parameters
	for _, el := range orders {
		_, err := statement.Exec(el.Filename, el.Number, el.Year, el.FullNameFormatted)
		if err != nil {
			log.Printf("Error during insert in db %v: %e\n", el, err)
		}
	}

	// Update to DB
	statement, err = db.Prepare(model.Set_is_Parsed)
	if err != nil {
		log.Fatal(err)
	}
	defer statement.Close()

	// Execute query with specific parameters
	for _, el := range orders {
		_, err := statement.Exec(el.Filename)
		if err != nil {
			log.Printf("Error during update in db %v: %e\n", el, err)
		}
	}
}
