package extractor

import (
	"regexp"
	"strconv"
)

// RegexWordExtractor извлекает слова с помощью регулярных выражений
type RegexWordExtractorImpl struct {
	re *regexp.Regexp
}

// NewRegexWordExtractor создаёт экстрактор слов с фиксированной минимальной длиной
func NewRegexWordExtractor(minLength int) *RegexWordExtractorImpl {
	pattern := "(?i)[a-zа-я]{" + strconv.Itoa(minLength) + ",}"
	return &RegexWordExtractorImpl{
		re: regexp.MustCompile(pattern),
	}
}

// ExtractWords извлекает слова из текста с минимальной длиной
func (e *RegexWordExtractorImpl) ExtractWords(text []byte) []string {
	return e.re.FindAllString(string(text), -1)
}
