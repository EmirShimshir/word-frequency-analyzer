package processor

import (
	"sort"
	"sync"

	"word-frequency-analyzer/internal/counter"
	"word-frequency-analyzer/internal/extractor"
	"word-frequency-analyzer/internal/models"
	"word-frequency-analyzer/internal/splitter"
)

// Processor обрабатывает файлы параллельно
type Processor interface {
	ProcessAll(files []string, minWordLen int) (map[string]int, error)
	GetTopWords(counts map[string]int, topCount int) []models.WordCount
}

// ConcurrentProcessor реализует параллельную обработку файлов
type ConcurrentProcessor struct {
	splitter  splitter.ChunkSplitter
	extractor extractor.WordExtractor
	counter   counter.WordCounter
	chunkSize int
}

// NewConcurrentProcessor создает новый процессор с зависимостями
func NewConcurrentProcessor(
	splitter splitter.ChunkSplitter,
	extractor extractor.WordExtractor,
	counter counter.WordCounter,
	chunkSize int,
) *ConcurrentProcessor {
	return &ConcurrentProcessor{
		splitter:  splitter,
		extractor: extractor,
		counter:   counter,
		chunkSize: chunkSize,
	}
}

// ProcessAll параллельно обрабатывает все файлы
func (p *ConcurrentProcessor) ProcessAll(files []string) (map[string]int, error) {
	var wg sync.WaitGroup
	resultCh := make(chan map[string]int, len(files))

	for _, file := range files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			// Обработка одного файла
			counts, err := p.processFile(filePath)
			if err != nil {
				// В реальном приложении здесь нужно обработать ошибку
				return
			}

			resultCh <- counts
		}(file)
	}

	// Закрываем канал после завершения всех горутин
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Собираем результаты из всех файлов
	var allCounts []map[string]int
	for counts := range resultCh {
		allCounts = append(allCounts, counts)
	}

	// Объединяем все подсчеты
	return p.counter.MergeCounts(allCounts...), nil
}

// processFile обрабатывает один файл параллельно по чанкам
func (p *ConcurrentProcessor) processFile(filePath string) (map[string]int, error) {
	chunks, err := p.splitter.SplitFile(filePath)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	chunkResultCh := make(chan map[string]int, len(chunks))

	for _, chunk := range chunks {
		wg.Add(1)
		go func(data []byte) {
			defer wg.Done()

			// Извлекаем слова из чанка
			words := p.extractor.ExtractWords(data)

			// Подсчитываем частоту слов
			counts := p.counter.Count(words)

			chunkResultCh <- counts
		}(chunk)
	}

	// Закрываем канал после завершения всех горутин
	go func() {
		wg.Wait()
		close(chunkResultCh)
	}()

	// Собираем результаты из всех чанков
	var chunkCounts []map[string]int
	for counts := range chunkResultCh {
		chunkCounts = append(chunkCounts, counts)
	}

	// Объединяем подсчеты из всех чанков
	return p.counter.MergeCounts(chunkCounts...), nil
}

// GetTopWords возвращает топ-N слов из карты частот
func (p *ConcurrentProcessor) GetTopWords(counts map[string]int, topCount int) []models.WordCount {
	result := make([]models.WordCount, 0, len(counts))

	// Преобразуем map в slice для сортировки
	for word, count := range counts {
		result = append(result, models.WordCount{
			Word:  word,
			Count: count,
		})
	}

	// Сортируем по убыванию частоты
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	// Ограничиваем количество результатов
	if len(result) > topCount {
		result = result[:topCount]
	}

	return result
}
