package urlloader

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/segmentio/parquet-go"
)

type CCIndex struct {
	// URLSurtkey              string    `parquet:"url_surtkey,gzip"`
	URL string `parquet:"url,gzip"`
	// URLHostName             string    `parquet:"url_host_name,gzip"`
	// URLHostTLD              string    `parquet:"url_host_tld,gzip"`
	// URLHost2ndLastPart      string    `parquet:"url_host_2nd_last_part,gzip"`
	// URLHost3rdLastPart      string    `parquet:"url_host_3rd_last_part,gzip"`
	// URLHost4thLastPart      string    `parquet:"url_host_4th_last_part,gzip"`
	// URLHost5thLastPart      string    `parquet:"url_host_5th_last_part,gzip"`
	// URLHostRegistrySuffix   string    `parquet:"url_host_registry_suffix,gzip"`
	URLHostRegisteredDomain string `parquet:"url_host_registered_domain,gzip"`
	// URLHostPrivateSuffix    string    `parquet:"url_host_private_suffix,gzip"`
	// URLHostPrivateDomain    string    `parquet:"url_host_private_domain,gzip"`
	// URLProtocol             string    `parquet:"url_protocol,gzip"`
	// URLPort                 int       `parquet:"url_port,gzip"`
	// URLPath                 string    `parquet:"url_path,gzip"`
	// URLQuery                string    `parquet:"url_query,gzip"`
	// FetchTime               time.Time `parquet:"fetch_time,gzip"`
	// FetchStatus             int16     `parquet:"fetch_status,gzip"`
	// ContentDigest           string    `parquet:"content_digest,gzip"`
	// ContentMimeType         string    `parquet:"content_mime_type,gzip"`
	// ContentMimeDetected     string    `parquet:"content_mime_detected,gzip"`
	// ContentCharset          string    `parquet:"content_charset,gzip"`
	// ContentLanguages        string    `parquet:"content_languages,gzip"`
	// WarcFilename            string    `parquet:"warc_filename,gzip"`
	// WarcRecordOffset        int       `parquet:"warc_record_offset,gzip"`
	// WarcRecordLength        int       `parquet:"warc_record_length,gzip"`
	// WarcSegment             string    `parquet:"warc_segment,gzip"`
}

type URLLoaderParquet struct {
	files           []string
	fileIndex       int
	reader          *parquet.Reader
	currentHost     string
	currentHostURLs []string
}

func NewParquet(dir string) (*URLLoaderParquet, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".parquet" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return &URLLoaderParquet{files: files}, nil
}

func (loader *URLLoaderParquet) Next() (CCIndex, error) {
	for {
		if loader.reader == nil {
			if loader.fileIndex >= len(loader.files) {
				return CCIndex{}, io.EOF
			}
			file, err := os.Open(loader.files[loader.fileIndex])
			if err != nil {
				return CCIndex{}, err
			}
			loader.fileIndex++
			loader.reader = parquet.NewReader(file)
		}

		var row CCIndex
		err := loader.reader.Read(&row)
		if err != nil {
			if err == io.EOF {
				loader.reader.Close()
				loader.reader = nil
				continue
			}
			return CCIndex{}, fmt.Errorf("read row: %w", err)
		}

		return row, nil
	}
}

func (l *URLLoaderParquet) LoadNextHostURLs() ([]string, error) {
	for {
		row, err := l.Next()
		if err != nil {
			if err == io.EOF {
				log.Println("URL Loader end of file")
				// end of file, return the URLs of the last host and reset the state
				result := l.currentHostURLs
				l.currentHostURLs = nil
				l.currentHost = ""
				return result, nil
			}

			return nil, fmt.Errorf("next url: %w", err)
		}
		if l.currentHost != "" && l.currentHost != row.URLHostRegisteredDomain {
			// host changed, return the URLs of the previous host
			result := l.currentHostURLs
			l.currentHostURLs = []string{row.URL} // start a new buffer for the new host
			l.currentHost = row.URLHostRegisteredDomain
			return result, nil
		}
		// same host, add the URL to the buffer
		l.currentHostURLs = append(l.currentHostURLs, row.URL)
		l.currentHost = row.URLHostRegisteredDomain
	}
	// This line should never be reached because the end of file is handled above
	return nil, errors.New("unexpected end of LoadNextHostURLs")
}

func (l *URLLoaderParquet) Close() error {
	return l.reader.Close()
}
