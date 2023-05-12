package http

import (
	"crypto/tls"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"time"
)

var client = &fasthttp.Client{
	NoDefaultUserAgentHeader:      true,
	DisableHeaderNamesNormalizing: true,
	MaxResponseBodySize:           1024 * 1024 * 10,
	ReadTimeout:                   time.Second * 180,
	Dial:                          fasthttpproxy.FasthttpProxyHTTPDialerTimeout(time.Second * 180),
	//Dial: (&fasthttp.TCPDialer{
	//	Concurrency:      1000,
	//	DNSCacheDuration: time.Hour,
	//}).Dial,
	TLSConfig: &tls.Config{
		InsecureSkipVerify: true,
	},
}

func Get(requestURI string) ([]byte, error) {

	// Acquire request and response from pool
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()

	// Set new request
	req.SetRequestURI(requestURI)
	req.Header.Set(fasthttp.HeaderAccept, "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set(fasthttp.HeaderAcceptEncoding, "gzip, deflate, br")
	req.Header.Set(fasthttp.HeaderAcceptLanguage, "en-US,en;q=0.9")
	req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="113", "Chromium";v="113", "Not-A.Brand";v="24"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set(fasthttp.HeaderUserAgent, "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36")

	// Do request
	err := client.Do(req, res)
	if err != nil {
		return nil, fmt.Errorf("client do: %w", err)
	}

	body, err := handleResponse(res)
	if err != nil {
		return nil, fmt.Errorf("handle response: %w", err)
	}

	return body, nil
}

func handleResponse(res *fasthttp.Response) ([]byte, error) {
	// Read and decode response body
	body, err := decodeResponse(res)
	if err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return body, nil
}

func decodeResponse(res *fasthttp.Response) ([]byte, error) {
	var err error
	var body []byte

	contentEncoding := res.Header.Peek(fasthttp.HeaderContentEncoding)
	switch string(contentEncoding) {
	case "gzip":
		body, err = res.BodyGunzip()
		if err != nil {
			return nil, fmt.Errorf("gunzip: %w", err)
		}
	case "deflate":
		body, err = res.BodyInflate()
		if err != nil {
			return nil, fmt.Errorf("inflate: %w", err)
		}
	case "br":
		body, err = res.BodyUnbrotli()
		if err != nil {
			return nil, fmt.Errorf("unbrotli: %w", err)
		}
	default:
		body = res.Body()
	}

	return body, nil
}
