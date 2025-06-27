package reader

import (
	"iter"
	"os"
	"path/filepath"
	"strings"

	"word-frequency-analyzer/internal/models/entities"
)

type FileReaderImpl struct {
	path      string
	chunkSize int
}

func NewFileReader(path string, chunkSize int) *FileReaderImpl {
	return &FileReaderImpl{
		path:      path,
		chunkSize: chunkSize,
	}
}

func (p *FileReaderImpl) ListTextFiles() ([]string, error) {
	var textFiles []string

	err := filepath.Walk(p.path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(filePath, ".txt") {
			textFiles = append(textFiles, filePath)
		}

		return nil
	})

	return textFiles, err
}

func (p *FileReaderImpl) Iterator(filename string) iter.Seq[entities.Chunk] {
	return func(yield func(chunk entities.Chunk) bool) {
		f, err := os.Open(filename)
		if err != nil {
			if !yield(entities.Chunk{Data: nil, Err: err}) {
				return
			}
		}
		defer f.Close()

		chunkReader := newChunkReader(f, p.chunkSize)
		for res := range chunkReader.iterator() {
			if !yield(res) {
				return
			}
		}

	}
}
