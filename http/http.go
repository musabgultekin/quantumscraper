package http

import (
	"crypto/tls"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

var client = &http.Client{
	Timeout: time.Second * 180,
	Transport: &http.Transport{
		Proxy: GetProxyUrl(),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func GetProxyUrl() func(*http.Request) (*url.URL, error) {
	proxyUrlStr := os.Getenv("PROXY_URL")
	proxyUrl, err := url.Parse(proxyUrlStr)
	if err != nil {
		return http.ProxyFromEnvironment
	}
	return http.ProxyURL(proxyUrl)
}

func Get(requestURI string) ([]byte, int, error) {
	// Set new request
	req, err := http.NewRequest(http.MethodGet, requestURI, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	// req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="113", "Chromium";v="113", "Not-A.Brand";v="24"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36")

	// Do request
	res, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("client do: %w", err)
	}
	defer res.Body.Close()

	body, err := handleResponse(res)
	if err != nil {
		return nil, 0, fmt.Errorf("handle response err: %w", err)
	}

	return body, res.StatusCode, nil
}

func handleResponse(res *http.Response) ([]byte, error) {
	// Check if its HTML
	contentType := res.Header.Get("Content-Type")
	if !strings.Contains(contentType, "html") {
		return nil, fmt.Errorf("not HTML")
	}

	// Check status code
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status not 200: %v", res.Status)
	}

	// Read and decode response body
	body, err := decodeResponse(res)
	if err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return body, nil
}

func decodeResponse(res *http.Response) ([]byte, error) {
	var err error
	var body []byte

	// Charset Decoding
	contentType := res.Header.Get("Content-Type")
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, fmt.Errorf("parse media type: %w", err)
	}
	if cs, ok := params["charset"]; ok {
		encoding, name := charset.Lookup(cs)
		if encoding == nil {
			return nil, fmt.Errorf("charset lookup: %v", name)
		}

		if encoding != nil {
			// If encoding is not nil, wrap body in a reader that converts from the given encoding to UTF-8.
			bodyReader := transform.NewReader(res.Body, encoding.NewDecoder())
			// Read the entire body, now decoded to UTF-8.
			body, err = io.ReadAll(bodyReader)
			if err != nil {
				return nil, fmt.Errorf("read all: %w", err)
			}
		}
	} else {
		body, err = io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("read all: %w", err)
		}
	}

	return body, nil
}
