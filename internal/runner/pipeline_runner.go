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

func (r *PipelineRunnerImpl) generator(doneCh <-chan struct{}, files []string) <-chan string {
	dataStream := make(chan string)

	go func() {
		defer close(dataStream)
		for _, file := range files {
			select {
			case <-doneCh:
				return
			case dataStream <- file:
			}
		}
	}()

	return dataStream
}

func (r *PipelineRunnerImpl) workReadFile(doneCh <-chan struct{}, inputCh <-chan string) <-chan entities.Chunk {
	resultStream := make(chan entities.Chunk)

	go func() {
		defer close(resultStream)
		for {
			select {
			case <-doneCh:
				return
			case filename, ok := <-inputCh:
				if !ok {
					return
				}
				for chunk := range r.fileReader.Iterator(filename) {
					select {
					case <-doneCh:
						return
					case resultStream <- chunk:
					}
				}
			}
		}
	}()

	return resultStream
}

func (r *PipelineRunnerImpl) fanOutReadFile(doneCh <-chan struct{}, inputCh <-chan string) []<-chan entities.Chunk {
	resultChannels := make([]<-chan entities.Chunk, r.workersCount/2)

	for i := 0; i < r.workersCount/2; i++ {
		resultChannels[i] = r.workReadFile(doneCh, inputCh)
	}

	return resultChannels
}

func (r *PipelineRunnerImpl) fanInReadFile(doneCh <-chan struct{}, channels ...<-chan entities.Chunk) <-chan entities.Chunk {
	finalStream := make(chan entities.Chunk)
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

func (r *PipelineRunnerImpl) workBuildMap(doneCh <-chan struct{}, inputCh <-chan entities.Chunk) <-chan entities.Map {
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

func (r *PipelineRunnerImpl) fanOutBuildMap(doneCh <-chan struct{}, inputCh <-chan entities.Chunk) []<-chan entities.Map {
	resultChannels := make([]<-chan entities.Map, r.workersCount/2)

	for i := 0; i < r.workersCount/2; i++ {
		resultChannels[i] = r.workBuildMap(doneCh, inputCh)
	}

	return resultChannels
}

func (r *PipelineRunnerImpl) fanInBuildMap(doneCh <-chan struct{}, channels ...<-chan entities.Map) <-chan entities.Map {
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

	// создаем N/2 горутин ReadFile с помощью fanOut
	channelsChunk := r.fanOutReadFile(doneCh, inputCh)

	// объединяем результаты из всех каналов
	channelChunk := r.fanInReadFile(doneCh, channelsChunk...)

	// создаем N/2 горутин BuildMap с помощью fanOut
	channelsMap := r.fanOutBuildMap(doneCh, channelChunk)

	// объединяем результаты из всех каналов
	ResultCh := r.fanInBuildMap(doneCh, channelsMap...)

	return r.counter(ResultCh)
}
