package splitter

import (
	"io/ioutil"
)

// ChunkSplitter делит файл на части для параллельной обработки
type ChunkSplitter interface {
	SplitFile(path string) ([][]byte, error)
}

// DefaultChunkSplitter — быстрая реализация без перекрытия и без строк
type DefaultChunkSplitter struct {
	chunkSize int
}

// NewDefaultChunkSplitter создаёт сплиттер
func NewDefaultChunkSplitter(chunkSize int) *DefaultChunkSplitter {
	return &DefaultChunkSplitter{
		chunkSize: chunkSize,
	}
}

// isSpace проверяет, является ли байт пробелом/переводом строки и т.п.
func isSpace(b byte) bool {
	// Только ASCII-пробельные символы
	return b == ' ' || b == '\n' || b == '\t' || b == '\r'
}

// SplitFile делит данные, не разрывая слова и не переходя к строкам
func (s *DefaultChunkSplitter) SplitFile(path string) ([][]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var chunks [][]byte
	start := 0

	for start < len(data) {
		end := start + s.chunkSize
		if end >= len(data) {
			end = len(data)
			chunks = append(chunks, data[start:end])
			break
		}

		// Найдём последний пробел до end
		split := end
		for split > start && !isSpace(data[split]) {
			split--
		}

		// если пробела не нашли — жертвуем словом, идём до следующего пробела
		if split == start {
			split = end
			for split < len(data) && !isSpace(data[split]) {
				split++
			}
		}

		// trim правые пробелы
		for split > start && isSpace(data[split-1]) {
			split--
		}

		chunks = append(chunks, data[start:split])

		// сдвигаем старт: пропускаем пробелы
		start = split
		for start < len(data) && isSpace(data[start]) {
			start++
		}
	}

	return chunks, nil
}
