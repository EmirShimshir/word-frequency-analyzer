package models

// WordCount представляет пару слово-частота
type WordCount struct {
	Word  string
	Count int
}

// Result содержит отсортированный список слов
type Result struct {
	Words []WordCount
}
