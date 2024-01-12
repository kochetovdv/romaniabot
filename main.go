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
// layoutOrder = "02.01.2006"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Error("Error during request: %e\n", err)
	}

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

	temp, err := extractors.OrderFiles(liTags)
	if err != nil {
		log.Printf("Error during extracting order files: %e\n", err)
		return
	}
	// write to file
	if err := fileutil.WriteToFile(outputFile, liTags); err != nil {
		log.Printf("Error during writing outputfile %s: %e\n", outputFile, err)
		return
	}
	fmt.Println(temp)
}
