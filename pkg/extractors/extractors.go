//Example: "<li>Data de&nbsp;<strong>26.10.2023&nbsp;</strong>numărul:&nbsp;&nbsp;<a href="https://cetatenie.just.ro/wp-content/uploads/2022/01/Ordin-1795-P-26.10.2023-art-11.pdf">1795P</a>&nbsp;<a href="https://cetatenie.just.ro/wp-content/uploads/2022/01/ordin-1796-P-26.10.2023-art-11.pdf">1796P</a>&nbsp;&nbsp;<a href="https://cetatenie.just.ro/wp-content/uploads/2022/01/ordin-1797-din-26.10.2023-art-11.pdf">1797P</a></li>"

package extractors

import (
	"fmt"
	"log"
	"path/filepath"
	"romaniabot/model"
	"strconv"
	"strings"

	"io"
	"regexp"

	"github.com/ledongthuc/pdf"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type orderLocal struct {
	Number            uint
	Year              uint
	FullNameFormatted string
}

// InsideTags saves all tags with attributes and text between <li> tags
// Returns: "<li>Data de&nbsp;<strong>26.10.2023&nbsp;</strong>numărul:&nbsp;&nbsp;<a href="https://cetatenie.just.ro/wp-content/uploads/2022/01/Ordin-1795-P-26.10.2023-art-11.pdf">1795P</a>&nbsp;<a href="https://cetatenie.just.ro/wp-content/uploads/2022/01/ordin-1796-P-26.10.2023-art-11.pdf">1796P</a>&nbsp;&nbsp;<a href="https://cetatenie.just.ro/wp-content/uploads/2022/01/ordin-1797-din-26.10.2023-art-11.pdf">1797P</a></li>"
func InsideTags(z *html.Tokenizer, tag atom.Atom) (string, error) {
	var result strings.Builder

	if z == nil {
		return "", fmt.Errorf("error during extracting tags: html.Tokenizer nil is received")
	}

	for {
		tt := z.Next()
		token := z.Token()
		result.WriteString(token.String())

		if tt == html.ErrorToken || (tt == html.EndTagToken && token.DataAtom == tag) {
			break
		}
	}

	if strings.Contains(result.String(), "<strong>") && strings.Contains(result.String(), "href") {
		return result.String(), nil
	}
	return "", nil
}

// StrongText returns the text content within <strong> tags
// Returns: "26.10.2023"
func StrongText(htmlString string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		return "", err
	}

	var strongText string
	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "strong" {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode {
					strongText = strings.TrimSpace(c.Data)
					return
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}
	extract(doc)
	return strongText, nil
}

// Links returns a map of href to text within <a> tags
// Returns: {https://cetatenie.just.ro/wp-content/uploads/2022/01/Ordin-1795-P-26.10.2023-art-11.pdf:1795P
// https://cetatenie.just.ro/wp-content/uploads/2022/01/ordin-1796-P-26.10.2023-art-11.pdf:1796P
// https://cetatenie.just.ro/wp-content/uploads/2022/01/ordin-1797-din-26.10.2023-art-11.pdf:1797P}
func Links(htmlString string) (map[string]string, error) {
	links := make(map[string]string)

	doc, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		return nil, err
	}

	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			href := ""
			text := ""

			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href = attr.Val
					break
				}
			}

			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode {
					text += c.Data
				}
			}

			text = strings.TrimSpace(text)

			if href != "" && text != "" {
				links[href] = text
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	extract(doc)
	return links, nil
}

// OrderFiles returns a slice of model.OrderFile from a slice of strings, sorted from 2018 to 202*+.
// Returns: {26.10.2023 https://cetatenie.just.ro/wp-content/uploads/2022/01/Ordin-1795-P-26.10.2023-art-11.pdf 1795P}
// {26.10.2023 https://cetatenie.just.ro/wp-content/uploads/2022/01/ordin-1796-P-26.10.2023-art-11.pdf 1796P}
// {26.10.2023 https://cetatenie.just.ro/wp-content/uploads/2022/01/ordin-1797-din-26.10.2023-art-11.pdf 1797P}
func OrderFiles(s []string) ([]model.OrderFile, error) {
	var result []model.OrderFile

	for i := len(s) - 1; i >= 0; i-- { // _, el := range s { //without sorting
		el := s[i] // doesn`t need without sorting
		strongText, err := StrongText(el)
		if err != nil {
			fmt.Printf("Error: %e\n", err)
			continue
		}

		links, err := Links(el)
		if err != nil {
			fmt.Printf("Error: %e\n", err)
		}

		for link, name := range links {
			order := model.OrderFile{
				Date:     strongText,
				URL:      link,
				Filename: filepath.Base(link),
				Name:     name,
			}
			result = append(result, order)
		}
	}

	return result, nil
}

// Parse the pdf file and return the list of orders
func Order(path string, orderFiles ...string) ([]model.Order, error) {
	// tager storage for data
	orders := make([]model.Order, 0, len(orderFiles))

	for _, filename := range orderFiles {
		ordersFromPDF, err := orderFromPDF(path, filename)
		if err != nil {
			fmt.Printf("error in orderFromPDf:%s\t%e\t", filename, err)
			continue
		}
		orders = append(orders, ordersFromPDF...)
	}

	return orders, nil
}

// Parse the pdf file and return the list of orders
func orderFromPDF(path string, filename string) ([]model.Order, error) {
	orders := make([]model.Order, 0)

	// Open the PDF file
	file, pdfReader, err := pdf.Open(path + filename)
	if err != nil {
		fmt.Printf("error in openning PDF-file: %s\n%e\n", filename, err)
		file.Close()
		return nil, err
	}
	defer file.Close()

	// Extract the text from the PDF
	text, err := pdfReader.GetPlainText()
	if err != nil {
		fmt.Printf("error in extracting file data: %s\n%e\n", filename, err)
		file.Close()
		return nil, err
	}

	data, _ := io.ReadAll(text)
	data2 := string(data)

	// Extract the digits using a regular expression
	fmt.Printf("String: %s\n", data2)

	//re := regexp.MustCompile(`(\d+\/\d{4})`)
	re := regexp.MustCompile(`(\d+\/[A-Za-z]{0,2}\/\d{4}|\d+\/\d{4})`)

	digits := re.FindAllString(data2, -1)
	fmt.Printf("digits: %v\t\n", digits)
	for _, digit := range digits {
		o, err := orderFromLine(digit)
		if err != nil {
			// return nil, err
			continue
		}
		order := model.Order{
			Filename:          filename,
			Year:              o.Year,
			Number:            o.Number,
			FullNameFormatted: o.FullNameFormatted,
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// orderFromLine extracts an order and year from a single line, returning a local struct
func orderFromLine(s string) (orderLocal, error) {
	// Split the input string
	parts := strings.Split(s, "/")

	// Check the length of the resulting slice for readability
	numParts := len(parts)

	// If the length of the slice is less than 2, it means there won't be an order and year
	if numParts < 2 {
		return orderLocal{}, fmt.Errorf("error during splitting in orderFromLine:%s\tresult is:%v", s, parts)
	}

	// Function for checking and extracting a number from a string
	checkAndExtractNumber := func(str string) (uint, error) {
		n, err := strconv.ParseUint(str, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("error extracting number from line: %s\t%w\t", str, err)
		}
		return uint(n), nil
	}

	// Function for checking and extracting a year from a string
	checkAndExtractYear := func(str string) (uint, error) {
		n, err := strconv.ParseUint(str, 10, 32)
		if err != nil {
			log.Printf("error extracting year from line: %s\t%e\t", str, err)
			return 0, fmt.Errorf("error extracting year from line: %s\t%w\t", str, err)
		}
		if n < 2010 || n > 2050 {
			log.Printf("error extracting year from line: %s\tinvalid year\t", str)
			return 0, fmt.Errorf("error extracting year from line: %s\tinvalid year\t", str)
		}
		return uint(n), nil
	}

	// Extract the number from the first part of the input
	number, err := checkAndExtractNumber(parts[0])
	if err != nil {
		return orderLocal{}, err
	}

	// Extract the year from the last part of the input
	year, err := checkAndExtractYear(parts[numParts-1])
	if err != nil {
		return orderLocal{}, err
	}

	// Format the full name using the extracted number and year
	fullName := strconv.Itoa(int(number)) + "/" + strconv.Itoa(int(year))

	// Create and return the order local struct
	order := orderLocal{
		Number:            number,
		Year:              year,
		FullNameFormatted: fullName,
	}

	return order, nil
}
