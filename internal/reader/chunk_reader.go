package reader

import (
	"bytes"
	"io"
	"iter"

	"word-frequency-analyzer/internal/models/entities"
)

type chunkReader struct {
	r         io.Reader
	buf       []byte // остаток от прошлого чтения
	chunkSize int
}

func newChunkReader(f io.Reader, chunkSize int) *chunkReader {
	return &chunkReader{
		r:         f,
		chunkSize: chunkSize,
	}
}

func (rc *chunkReader) lastIndexByte(data []byte) int {
	return max(bytes.LastIndexByte(data, ' '),
		bytes.LastIndexByte(data, '\n'),
		bytes.LastIndexByte(data, '\t'),
		bytes.LastIndexByte(data, '\r'))
}

func (rc *chunkReader) readChunk() ([]byte, error) {
	tmp := make([]byte, rc.chunkSize)
	n, err := rc.r.Read(tmp)
	if n == 0 && err != nil {
		return nil, err
	}
	data := append(rc.buf, tmp[:n]...)

	idx := rc.lastIndexByte(data)
	if idx == -1 {
		rc.buf = nil
		return data, err
	}

	chunk := data[:idx]
	rc.buf = data[idx+1:]

	return chunk, err
}

// Итератор-обёртка: возвращает Chunk
func (rc *chunkReader) iterator() iter.Seq[entities.Chunk] {
	return func(yield func(entities.Chunk) bool) {
		for {
			chunk, err := rc.readChunk()

			if chunk != nil {
				if !yield(entities.Chunk{Data: chunk, Err: nil}) {
					return
				}
			}

			if err == io.EOF {
				if chunk == nil && len(rc.buf) > 0 {
					if !yield(entities.Chunk{Data: rc.buf, Err: nil}) {
						return
					}
				}
				return
			}

			if err != nil {
				if !yield(entities.Chunk{Data: nil, Err: err}) {
					return
				}
				return
			}
		}
	}
}
