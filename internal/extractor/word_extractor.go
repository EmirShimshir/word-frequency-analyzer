package extractor

import (
	"regexp"
	"strconv"
)

// WordExtractor извлекает слова из текста
type WordExtractor interface {
	ExtractWords(text []byte) []string
}

// RegexWordExtractor извлекает слова с помощью регулярных выражений
type RegexWordExtractor struct {
	re *regexp.Regexp
}

// NewRegexWordExtractor создаёт экстрактор слов с фиксированной минимальной длиной
func NewRegexWordExtractor(minLength int) *RegexWordExtractor {
	pattern := "(?i)[a-zа-я]{" + strconv.Itoa(minLength) + ",}"
	return &RegexWordExtractor{
		re: regexp.MustCompile(pattern),
	}
}

// ExtractWords извлекает слова из текста с минимальной длиной
func (e *RegexWordExtractor) ExtractWords(text []byte) []string {
	return e.re.FindAllString(string(text), -1)
}
