package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"word-frequency-analyzer/internal/counter"
	"word-frequency-analyzer/internal/extractor"
	"word-frequency-analyzer/internal/models"
	"word-frequency-analyzer/internal/processor"
	"word-frequency-analyzer/internal/provider"
	"word-frequency-analyzer/internal/splitter"
)

const (
	defaultChunkSize = 50 // 1MB
)

func main() {
	// Парсим аргументы командной строки
	dirPath := flag.String("dir", "", "Путь к директории с текстовыми файлами")
	minWordLen := flag.Int("minlen", 5, "Минимальная длина слова")
	topCount := flag.Int("top", 10, "Количество топ-слов для вывода")
	outputFile := flag.String("output", "", "файл с результатами")
	flag.Parse()

	if *dirPath == "" {
		fmt.Println("Необходимо указать директорию с файлами (--dir)")
		os.Exit(1)
	}

	// Создаем все компоненты
	fileProvider := provider.NewDiskFileProvider(*dirPath)
	chunkSplitter := splitter.NewDefaultChunkSplitter(defaultChunkSize)
	wordExtractor := extractor.NewRegexWordExtractor(*minWordLen)
	wordCounter := counter.NewDefaultWordCounter()

	// Создаем процессор
	processor := processor.NewConcurrentProcessor(
		chunkSplitter,
		wordExtractor,
		wordCounter,
		defaultChunkSize,
	)

	// Получаем список файлов
	files, err := fileProvider.ListTextFiles()
	if err != nil {
		fmt.Printf("Ошибка при чтении директории: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("В указанной директории не найдено текстовых файлов (.txt)")
		os.Exit(0)
	}

	fmt.Printf("Найдено %d текстовых файлов\n", len(files))

	// Запускаем обработку
	counts, err := processor.ProcessAll(files)
	if err != nil {
		fmt.Printf("Ошибка при обработке файлов: %v\n", err)
		os.Exit(1)
	}

	// Получаем топ-N слов
	topWords := processor.GetTopWords(counts, *topCount)

	// Выводим результат
	fmt.Printf("\nТоп-%d слов с минимальной длиной %d:\n", *topCount, *minWordLen)
	for i, word := range topWords {
		fmt.Printf("%d. %s: %d\n", i+1, word.Word, word.Count)
	}

	// Записываем результат в файл, если указан путь
	if *outputFile != "" {
		err := SaveToFile(topWords, *outputFile)
		if err != nil {
			fmt.Printf("Ошибка при сохранении результатов в файл: %v\n", err)
		} else {
			fmt.Printf("\nРезультаты сохранены в файл: %s\n", *outputFile)
		}
	}
}

// SaveToFile сохраняет результаты анализа в текстовый файл
func SaveToFile(words []models.WordCount, filename string) error {
	file, err := os.Create(filename)
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
