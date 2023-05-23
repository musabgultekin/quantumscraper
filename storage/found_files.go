package storage

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strconv"

	"github.com/klauspost/compress/gzip"
)

func WriteLinksToFile(links map[string]struct{}, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	w := bufio.NewWriterSize(gw, 1024*1024*16) // 16 MB buffer
	defer w.Flush()

	for key := range links {
		_, err := w.WriteString(key + "\n")
		if err != nil {
			return fmt.Errorf("failed to write key to file: %w", err)
		}
	}

	return nil
}

func WriteLinksToFileRandomFilename(links map[string]struct{}, baseDir string) error {
	filename := generateRandomFilename()

	return WriteLinksToFile(links, path.Join(baseDir, filename+".txt.gz"))
}

func generateRandomFilename() string {
	return strconv.FormatUint(rand.Uint64(), 10)
}
