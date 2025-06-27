package reader

import (
	"iter"
	"os"
	"path/filepath"
	"strings"

	"word-frequency-analyzer/internal/models/entities"
	"word-frequency-analyzer/internal/models/ports"
)

type FileReaderImpl struct {
	path        string
	chunkReader ports.ChunkSplitter
}

func NewFileReader(chunkReader ports.ChunkSplitter, path string) *FileReaderImpl {
	return &FileReaderImpl{
		path:        path,
		chunkReader: chunkReader,
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

func (p *FileReaderImpl) Iterator(files []string) iter.Seq[entities.Chunk] {
	return func(yield func(chunk entities.Chunk) bool) {
		for _, filename := range files {
			f, err := os.Open(filename)
			if err != nil {
				if !yield(entities.Chunk{Data: nil, Err: err}) {
					return
				}
				continue
			}

			func() {
				defer f.Close()

				p.chunkReader.Init(f)
				for res := range p.chunkReader.Iterator() {
					if !yield(res) {
						return
					}
				}
			}()
		}
	}
}
