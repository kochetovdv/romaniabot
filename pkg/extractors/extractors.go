//Example: "<li>Data de&nbsp;<strong>26.10.2023&nbsp;</strong>numărul:&nbsp;&nbsp;<a href="https://cetatenie.just.ro/wp-content/uploads/2022/01/Ordin-1795-P-26.10.2023-art-11.pdf">1795P</a>&nbsp;<a href="https://cetatenie.just.ro/wp-content/uploads/2022/01/ordin-1796-P-26.10.2023-art-11.pdf">1796P</a>&nbsp;&nbsp;<a href="https://cetatenie.just.ro/wp-content/uploads/2022/01/ordin-1797-din-26.10.2023-art-11.pdf">1797P</a></li>"

package extractors

import (
	"fmt"
	"path/filepath"
	"romaniabot/model"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

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
