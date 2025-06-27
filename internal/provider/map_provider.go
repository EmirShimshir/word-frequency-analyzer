package provider

import (
	"fmt"
	"sort"
	"word-frequency-analyzer/internal/models/entities"
	"word-frequency-analyzer/internal/models/ports"
)

type MapProviderImpl struct {
	extractor ports.WordExtractor
	topCount  int
}

func NewMapProvider(extractor ports.WordExtractor, topCount int) *MapProviderImpl {
	return &MapProviderImpl{
		extractor: extractor,
		topCount:  topCount,
	}
}

func (m *MapProviderImpl) BuildMap(chunk entities.Chunk) entities.Map {
	words := m.extractor.ExtractWords(chunk.Data)

	data := make(map[string]int)
	for _, word := range words {
		data[word]++
	}
	fmt.Println(data)
	return entities.Map{Data: data, Err: nil}
}

func (m *MapProviderImpl) MergeMap(dst, src entities.Map) {
	for k, v := range src.Data {
		dst.Data[k] += v
	}
}

func (m *MapProviderImpl) GetTopWords(resMap entities.Map) []entities.WordCount {
	result := make([]entities.WordCount, 0, len(resMap.Data))

	// Преобразуем map в slice для сортировки
	for word, count := range resMap.Data {
		result = append(result, entities.WordCount{
			Word:  word,
			Count: count,
		})
	}

	// Сортируем по убыванию частоты
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	// Ограничиваем количество результатов
	if len(result) > m.topCount {
		result = result[:m.topCount]
	}

	return result
}
