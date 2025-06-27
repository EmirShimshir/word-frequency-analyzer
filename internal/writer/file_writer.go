package writer

import (
	"bufio"
	"fmt"
	"os"

	"word-frequency-analyzer/internal/models/entities"
)

type FileWriterImpl struct {
	filename string
}

func NewFileWriter(filename string) *FileWriterImpl {
	return &FileWriterImpl{
		filename: filename,
	}
}

func (w *FileWriterImpl) SaveToFile(words []entities.WordCount) error {
	file, err := os.Create(w.filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for i, word := range words {
		_, err := fmt.Fprintf(writer, "%d. %s: %d\n", i+1, word.Word, word.Count)
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}
