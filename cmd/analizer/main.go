package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"word-frequency-analyzer/internal/extractor"
	"word-frequency-analyzer/internal/provider"
	"word-frequency-analyzer/internal/reader"
	"word-frequency-analyzer/internal/runner"
	"word-frequency-analyzer/internal/splitter"
	"word-frequency-analyzer/internal/writer"
)

const (
	defaultChunkSize = 1024 // 1КB
)

func main() {
	defaultWorkersCount := runtime.NumCPU()

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

	if *outputFile == "" {
		fmt.Println("Необходимо указать выходной файл (--output)")
		os.Exit(1)
	}

	// Создаем все компоненты
	ChunkSplitter := splitter.NewChunkSplitter(defaultChunkSize)
	fileReader := reader.NewFileReader(ChunkSplitter, *dirPath)

	wordExtractor := extractor.NewRegexWordExtractor(*minWordLen)
	mapProvider := provider.NewMapProvider(wordExtractor, *topCount)

	fileWriter := writer.NewFileWriter(*outputFile)

	pipelineRunner := runner.NewPipelineRunner(fileReader, fileWriter, mapProvider, defaultWorkersCount)

	err := pipelineRunner.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
