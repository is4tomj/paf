package pio

import (
	"io"
	"os"
	"sync"
)

// Process will loop through the chunks in a paf file and execute the function passed
func Process(file *os.File, chunkSize, numProcs int, initFunc func(int, int64), processFunc func(int, *Chunk)) {
	chunks, fileSize, err := findDataChunks(file, chunkSize)
	if err != nil {
		pesf(err.Error())
		return
	}

	// Create chunkChan
	numChunks := len(chunks)
	chunksChan := make(chan *Chunk, numChunks)
	for i:=0; i<numChunks; i++ {
		chunksChan <- chunks[i]
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

			for chunk := range chunksChan {
				numToGo := len(chunksChan)
				numDone := numChunks - numToGo
				pes(sprintf("\rprocessed %d of %d (%d%%) chunks", numDone, numChunks, (numDone*100)/numChunks))
				if processFunc != nil {
					processFunc(pid, chunk)
				}
			}
		}(i)
	}

	// Close and wait
	close(chunksChan)
	wg.Wait()
	pes(sprintf("\rFinished processing %d chunks.                       \n", numChunks))
}

// FindDataChunks finds the chunks in a paf file
func findDataChunks(file *os.File, chunkSize int) ([]*Chunk, int64, error) {

	const maxInt = 2147483647
	if chunkSize <= 1024 {
		panic(sprintf("%s Chunk size cannot be less than one KB.", Ep))
	}

	fileStats, err := file.Stat()
	if err != nil {
		panic(err)
	}
	fileSize := fileStats.Size()

	// Calculate number of chunks
	numChunks := (fileSize / int64(chunkSize)) + int64(1)
	chunks := make([]*Chunk, numChunks*2)

	// Create Chunks
	entryPoint := int64(0)
	size := chunkSize
	const delta = 1024
	buff := make([]byte, delta)

	i := 0
	for ; true; i++ {
		// get last 1024 bytes of proposed chunk
		n, err := file.ReadAt(buff, (entryPoint+int64(size))-delta)
		if err != nil && err != io.EOF {
			return nil, fileSize, err
		}
		
		// last chunk
		if err == io.EOF {
			dataSize := int(fileSize - entryPoint)
			chunks[i] = &Chunk{entryPoint, dataSize, dataSize, file}
			break
		}
		
		j := n - 1
		for j >= 0 {
			if buff[j] == Nl {
				chunks[i] = &Chunk{entryPoint, size, size, file}
				entryPoint = entryPoint + int64(size)
				size = chunkSize
				break
			}
			j--
			size--
			// If no '\n' is in the last 1024 bytes, then go back to
			// the top of the loop and get the next 1024 to find '\n'.
		}
	}
	return chunks[:i+1], fileSize, nil
}
