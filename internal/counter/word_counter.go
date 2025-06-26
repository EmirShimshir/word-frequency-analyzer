package counter

// WordCounter подсчитывает частоту слов
type WordCounter interface {
	Count(words []string) map[string]int
	MergeCounts(counts ...map[string]int) map[string]int
}

// DefaultWordCounter стандартная реализация подсчета слов
type DefaultWordCounter struct{}

// NewDefaultWordCounter создает новый счетчик слов
func NewDefaultWordCounter() *DefaultWordCounter {
	return &DefaultWordCounter{}
}

// Count подсчитывает частоту слов
func (c *DefaultWordCounter) Count(words []string) map[string]int {
	counts := make(map[string]int)
	for _, word := range words {
		counts[word]++
	}
	return counts
}

// MergeCounts объединяет несколько карт подсчета
func (c *DefaultWordCounter) MergeCounts(counts ...map[string]int) map[string]int {
	result := make(map[string]int)

	for _, count := range counts {
		for word, freq := range count {
			result[word] += freq
		}
	}

	return result
}
