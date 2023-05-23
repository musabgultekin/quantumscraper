package worker

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

func extractLinksFromHTML(pageURL string, body []byte) (linkSet map[string]struct{}, err error) {
	pageURLParsed, err := url.Parse(pageURL)
	if err != nil {
		return nil, fmt.Errorf("page url parse: %w", err)
	}

	htmlLinkStrings, err := extractRawLinksFromHTML(body)
	if err != nil {
		return nil, fmt.Errorf("extract raw links from html: %w", err)
	}

	// Use a map to ensure uniqueness of links
	linkSet = make(map[string]struct{})

	// Convert to absolute
	for _, htmlLinkString := range htmlLinkStrings {
		absoluteHTMLink, err := pageURLParsed.Parse(htmlLinkString)
		if err != nil {
			// log.Println("WARN: page link parse error:", err, pageURL)
			continue
		}
		linkSet[absoluteHTMLink.String()] = struct{}{}
	}

	return
}

// func extractRawLinksFromHTML(body []byte) ([]string, error) {
// 	acceptedExtensions := []string{".asp", ".aspx", ".htm", ".html", ".jsp", ".jsx", ".php", ".php3", ".php4", ".php5", ".phtml"}
// 	var links []string

// 	// bodyReader := bytes.NewReader(body)
// 	lolhtmlWriter, err := lolhtml.NewWriter(io.Discard, &lolhtml.Handlers{
// 		ElementContentHandler: []lolhtml.ElementContentHandler{{
// 			Selector: "a",
// 			ElementHandler: func(e *lolhtml.Element) lolhtml.RewriterDirective {
// 				href, err := e.AttributeValue("href")
// 				if err != nil {
// 					log.Println("Attibute reading error", err)
// 					return lolhtml.Continue
// 				}

// 				href = strings.TrimSpace(href)

// 				if !strings.HasPrefix(href, "http") {
// 					return lolhtml.Continue
// 				}

// 				ext := filepath.Ext(href)
// 				if ext == "" || contains(acceptedExtensions, ext) {
// 					links = append(links, href)
// 				}

// 				return lolhtml.Continue
// 			},
// 		}},
// 	})
// 	if err != nil {
// 		return nil, fmt.Errorf("lolhtml writer: %w", err)
// 	}

// 	// copy from the bytes reader to lolhtml writer
// 	if _, err := lolhtmlWriter.Write(body); err != nil {
// 		return nil, fmt.Errorf("lolhtml copy err: %w", err)
// 	}

// 	// explicitly close the writer and flush the remaining content
// 	if err := lolhtmlWriter.Close(); err != nil {
// 		return nil, fmt.Errorf("lolhtml close: %w", err)
// 	}

// 	return nil, nil
// }

func extractRawLinksFromHTML(body []byte) ([]string, error) {
	var links []string

	z := html.NewTokenizer(bytes.NewReader(body))
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if z.Err() == io.EOF {
				return links, nil
			}
			return nil, z.Err()
		case html.StartTagToken:
			tagName, moreAttr := z.TagName()
			if len(tagName) == 1 && tagName[0] == 'a' {
				for moreAttr {
					var key, val []byte
					key, val, moreAttr = z.TagAttr()
					if string(key) == "href" && len(val) != 0 {
						valString := strings.TrimSpace(string(val))
						if !strings.HasPrefix(valString, "http") {
							break
						}
						acceptedExtensions := []string{".asp", ".aspx", ".htm", ".html", ".jsp", ".jsx", ".php", ".php3", ".php4", ".php5", ".phtml"}
						ext := filepath.Ext(valString)
						if ext == "" || contains(acceptedExtensions, ext) {
							links = append(links, valString)
						}
						break
					}
				}
			}
		}
	}
}

// contains checks if a slice contains a string
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
