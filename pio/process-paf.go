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
	for _, chunk := range chunks {
		chunksChan <- chunk
	}
	pes(sprintf("chunks:%d\n", numChunks))

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
				pes(sprintf("\rFinished %d of %d (%d%%)", numDone, numChunks, (numDone*100)/numChunks))
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
	chunks := make([]*Chunk, numChunks)

	// Create Chunks
	entryPoint := int64(0)
	size := chunkSize
	const delta = 1024
	buff := make([]byte, delta)
	num := int64(0)

	for i := 0; true; i++ {
		// get last 1024 bytes of proposed chunk
		n, err := file.ReadAt(buff, (entryPoint+int64(size))-delta)
		if err != nil && err != io.EOF {
			return nil, fileSize, err
		}
		
		// last chunk
		if err == io.EOF {
			dataSize := int(fileSize - entryPoint)
			//chunks = append(chunks, &Chunk{entryPoint, dataSize, dataSize, file})
			chunks[i] = &Chunk{entryPoint, dataSize, dataSize, file}
			return chunks, fileSize, nil
		}
		
		j := n - 1
		for j >= 0 {
			if buff[j] == Nl {
				//chunks = append(chunks, &Chunk{entryPoint, size, size, file})
				chunks[i] = &Chunk{entryPoint, size, size, file}
				num++
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
	
	return chunks, fileSize, nil
}
