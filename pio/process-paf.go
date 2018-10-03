package pio

import (
	"io"
	"os"
	"sync"
)

// Process will loop through the chunks in a paf file and execute the function passed
func Process(inputPath string, chunkSize, numProcs int, initFunc func(int, int64), processFunc func(int, *Chunk)) {
	chunks, fileSize, err := findDataChunks(inputPath, chunkSize, numProcs+1)
	if err != nil {
		pesf(err.Error())
		return
	}

	// Spin up goroutines
	var wg sync.WaitGroup
	wg.Add(numProcs)
	for i := 0; i < numProcs; i++ {
		go func(pid int) {
			defer wg.Done()
			if initFunc != nil {
				initFunc(pid, fileSize)
			}

			for chunk := range chunks {
				if processFunc != nil {
					processFunc(pid, chunk)
				}
			}
		}(i)
	}

	// Wait to close
	wg.Wait()

}

// FindDataChunks finds the chunks in a paf file
func findDataChunks(inputPath string, chunkSize, chanSize int) (chan *Chunk, int64, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, 0, err
	}

	const maxInt = 2147483647
	if chunkSize <= 1024 {
		panic(sprintf("%s Chunk size cannot be less than one KB.", ep))
	}

	fileStats, err := file.Stat()
	if err != nil {
		panic(err)
	}
	fileSize := fileStats.Size()

	// Create Chunks
	chunks := make(chan *Chunk, chanSize)
	entryPoint := int64(0)
	size := chunkSize
	const delta = 1024
	buff := make([]byte, delta)
	num := int64(0)
	go func() {
		defer close(chunks)
		defer file.Close()
		for true {
			// get last 1024 bytes of proposed chunk
			n, err := file.ReadAt(buff, (entryPoint+int64(size))-delta)
			if err != nil && err != io.EOF {
				pesf(err.Error() + "\n")
				return
			}

			// last chunk
			if err == io.EOF {
				dataSize := int(fileSize - entryPoint)
				chunks <- &Chunk{inputPath, entryPoint, dataSize, dataSize}
				return
			}

			j := n - 1
			for j >= 0 {
				if buff[j] == nl {
					chunks <- &Chunk{inputPath, entryPoint, size, size}
					num++
					entryPoint = entryPoint + int64(size)
					size = chunkSize
					break
				}
				j--
				size--
			}
			// If no '\n' is in the last 1024 bytes, then go back to
			// the top of the loop and get the next 1024 to find '\n'.
		}
	}()

	return chunks, fileSize, nil
}
