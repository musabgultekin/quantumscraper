package http

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"os"
	"time"
)

var client = &fasthttp.Client{
	Name:        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36",
	ReadTimeout: time.Second * 180,
	Dial:        fasthttpproxy.FasthttpHTTPDialer(os.Getenv("HTTP_PROXY")),
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
	req.Header.Set(fasthttp.HeaderAccept, "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set(fasthttp.HeaderAcceptEncoding, "gzip, deflate, br")

	// Do request
	err := client.Do(req, res)
	if err != nil {
		return nil, err
	}

	body, err := handleResponse(res)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func handleResponse(res *fasthttp.Response) ([]byte, error) {
	// Read and decode response body
	body, err := decodeResponse(res)
	if err != nil {
		return nil, err
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
