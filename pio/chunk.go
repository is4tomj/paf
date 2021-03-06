package pio

import (
	"errors"
	"io"
	"os"
)

// Chunk is a pafio privitive that comprises a subset of a file
type Chunk struct {
	EntryPoint int64 // in bytes
	Size       int   // in bytes
	DataSize   int   // in bytes
	Index      int   // order of chunk
	file       *os.File
}

// Bytes returns the bytes in a chunk
func (chunk *Chunk) Bytes(buff *[]byte) ([]byte, int, error) {
	file := chunk.file

	// check if byte buffer is already provided
	if buff == nil || cap(*buff) < chunk.Size {
		arr := make([]byte, chunk.Size+4096) // 4096 is a magic number
		buff = &arr
	}

	// get slice of buffer that is the desired size
	partial := (*buff)[:chunk.Size]

	n, err := file.ReadAt(partial, chunk.EntryPoint)
	if err != nil && err != io.EOF {
		return nil, 0, err
	}

	// n should be equal to chunk.Size, so provide error if not
	if n != chunk.Size {
		err := errors.New(sprintf("Read %d bytes, but chunk size is %d bytes.\n", n, chunk.Size))
		return partial, n, err
	}

	return partial, n, nil
}

// NewLineScanner creates a scanner that is '\n' delimited
func NewLineScanner(b []byte) func() ([]byte, int) {
	start := -1
	end := -1
	max := len(b)
	return func() ([]byte, int) {
		end++
		if end >= max {
			return nil, 0
		}
		for start = end; end < max; end++ {
			if b[end] == nl { // look for newline char
				return b[start:end], end - start
			}
		}
		return b[start:max], max - start
	}
}
