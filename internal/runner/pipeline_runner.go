package runner

import (
	"fmt"
	"sync"
	"word-frequency-analyzer/internal/models/entities"
	"word-frequency-analyzer/internal/models/ports"
)

type PipelineRunnerImpl struct {
	fileReader   ports.FileReader
	fileWriter   ports.FileWriter
	mapProvider  ports.MapProvider
	workersCount int
}

func NewPipelineRunner(fileReader ports.FileReader, fileWriter ports.FileWriter, mapProvider ports.MapProvider, workersCount int) *PipelineRunnerImpl {
	return &PipelineRunnerImpl{
		fileReader:   fileReader,
		fileWriter:   fileWriter,
		mapProvider:  mapProvider,
		workersCount: workersCount,
	}
}

func (r *PipelineRunnerImpl) generator(doneCh <-chan struct{}, files []string) <-chan entities.Chunk {
	dataStream := make(chan entities.Chunk)

	go func() {
		defer close(dataStream)
		for res := range r.fileReader.Iterator(files) {
			select {
			case <-doneCh:
				return
			case dataStream <- res:
			}
		}
	}()

	return dataStream
}

func (r *PipelineRunnerImpl) work(doneCh <-chan struct{}, inputCh <-chan entities.Chunk) <-chan entities.Map {
	resultStream := make(chan entities.Map)

	go func() {
		defer close(resultStream)
		for res := range inputCh {
			if res.Err != nil {
				select {
				case <-doneCh:
					return
				case resultStream <- entities.Map{nil, res.Err}: // пробрасываем ошибку дальше
				}
				continue
			}

			data := r.mapProvider.BuildMap(res)

			select {
			case <-doneCh:
				return
			case resultStream <- data:
			}
		}
	}()

	return resultStream
}

func (r *PipelineRunnerImpl) fanOut(doneCh <-chan struct{}, inputCh <-chan entities.Chunk) []<-chan entities.Map {
	resultChannels := make([]<-chan entities.Map, r.workersCount)

	for i := 0; i < r.workersCount; i++ {
		resultChannels[i] = r.work(doneCh, inputCh)
	}

	return resultChannels
}

func (r *PipelineRunnerImpl) fanIn(doneCh <-chan struct{}, channels ...<-chan entities.Map) <-chan entities.Map {
	finalStream := make(chan entities.Map)
	var wg sync.WaitGroup

	for _, ch := range channels {
		chCopy := ch
		wg.Add(1)

		go func() {
			defer wg.Done()
			for value := range chCopy {
				select {
				case <-doneCh:
					return
				case finalStream <- value:
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(finalStream)
	}()

	return finalStream
}

func (r *PipelineRunnerImpl) counter(ch <-chan entities.Map) error {
	final := entities.Map{
		Data: make(map[string]int),
		Err:  nil,
	}

	for m := range ch {
		if m.Err != nil {
			fmt.Printf("Error: %v\n", m.Err) // вывод в stdout
			continue
		}

		r.mapProvider.MergeMap(final, m)
	}

	wordsCount := r.mapProvider.GetTopWords(final)
	return r.fileWriter.SaveToFile(wordsCount)
}

func (r *PipelineRunnerImpl) Run() error {
	doneCh := make(chan struct{})
	defer close(doneCh)

	files, err := r.fileReader.ListTextFiles()
	if err != nil {
		return err
	}

	inputCh := r.generator(doneCh, files)

	// создаем N горутин work с помощью fanOut
	channels := r.fanOut(doneCh, inputCh)

	// объединяем результаты из всех каналов
	ResultCh := r.fanIn(doneCh, channels...)

	return r.counter(ResultCh)
}
