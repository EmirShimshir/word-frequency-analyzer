package ports

import (
	"iter"

	"word-frequency-analyzer/internal/models/entities"
)

type PipelineRunner interface {
	Run() error
}
type FileReader interface {
	ListTextFiles() ([]string, error)
	Iterator(filename string) iter.Seq[entities.Chunk]
}

// WordExtractor извлекает слова из текста
type WordExtractor interface {
	ExtractWords(text []byte) []string
}

type MapProvider interface {
	BuildMap(chunk entities.Chunk) entities.Map
	MergeMap(dst, src entities.Map)
	GetTopWords(resMap entities.Map) []entities.WordCount
}

type FileWriter interface {
	SaveToFile(words []entities.WordCount) error
}
