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

func extractLinksFromHTML(pageURL string, body []byte) (fullURLs []string, err error) {
	pageURLParsed, err := url.Parse(pageURL)
	if err != nil {
		return nil, fmt.Errorf("page url parse: %w", err)
	}

	htmlLinkStrings, err := extractRawLinksFromHTML(body)
	if err != nil {
		return nil, fmt.Errorf("extract raw links from html: %w", err)
	}

	// Convert to absolute
	for _, htmlLinkString := range htmlLinkStrings {
		absoluteHTMLink, err := pageURLParsed.Parse(htmlLinkString)
		if err != nil {
			// log.Println("WARN: page link parse error:", err, pageURL)
			continue
		}
		fullURLs = append(fullURLs, absoluteHTMLink.String())
	}

	return
}

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
