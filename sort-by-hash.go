package main

import (
	"encoding/hex"
	"encoding/base64"
	"paf/pio"
	"bytes"
	"flag"
	"os"
	"sync"
)

func sortByHash() {
	sortFlags := flag.NewFlagSet("hash", flag.ContinueOnError)
	numProcs := sortFlags.Int("num-procs", 1, "number of processors")
	inputFile := sortFlags.String("input-file", "", "file to read")
	column := sortFlags.Int("col", 0, "zero-based column number with hash to sort by")

	tmpPath := sortFlags.String("tmp-dir", ".tmp", "directory to temporarily store files during sort")

	hexEnc := sortFlags.Bool("hex", true, "output hashes encoded in hex")
	base64Enc := sortFlags.Bool("base64", false, "output hashes encoded in base64")

	if err := sortFlags.Parse(os.Args[2:]); err != nil || len(os.Args[2:]) == 0 || *inputFile == "" {
		pes(`
Sort a column in a paf file with hash and print to stdout.
Examples:
    $ paf sort --num-procs 4 --input-file input.paf --col=0 --hex --tmp-dir data/tmp
`)
		os.Exit(1)
	}

	file, err := os.Open(*inputFile)
	if err != nil {
		pes(err.Error())
		os.Exit(1)
	}
	defer file.Close()

	// delete tmp dir if already exists
	pes(sprintf("Preparing tmp directory:%s\n", (*tmpPath)))
	if _, err := os.Stat(*tmpPath); os.IsExist(err) {
		err := os.RemoveAll(*tmpPath)
		if err != nil {
			pes(err.Error())
			os.Exit(1)
		}
	}
	// create tmp dir
	os.MkdirAll(*tmpPath, os.ModePerm)

	// initialize tmpfiles
	const max = 0xFFFF
	tmpFiles := make([]*TmpFile,max+1)
	for i:=0; i<=max; i++ {
		path := (*tmpPath)+sprintf("/%04x",uint16(i))
		tmpFiles[i] = &TmpFile{path: path, count: 0, mux: &sync.Mutex{}}
	}

	pes("Starting to process\n")
	pio.Process(file, chunkSize*100000, *numProcs, nil, func(pid int, chunk *pio.Chunk) {
		buff, n, err := (*chunk).Bytes(nil)
		
		if err != nil {
			pes(err.Error())
			os.Exit(1)
		} else if n < 0 {
			pes(sprintf("Shit! n is less than zero: %d.", n))
			os.Exit(1)
		} else {
			// traverse lines
			scan := pio.NewLineScanner(buff)
			for line, lineLen := scan(); lineLen > 0; line, lineLen = scan() {
				// read line
				hashEncBytes := bytes.Split(line,[]byte("\t"))[*column]
				hashDecBytes := make([]byte, 32)
				if *hexEnc {
					_, err := hex.Decode(hashDecBytes, hashEncBytes)
					if err != nil {
						pes("Shit!")
						panic(0)
					}
				} else if *base64Enc {
					_, err := base64.URLEncoding.Decode(hashDecBytes, hashEncBytes)
					if err != nil {
						pes("Shit!")
						panic(err)
					}
				}
				// get the first (most significant) two bytes
				idx := int32(uint16(hashDecBytes[0]) << 8 | uint16(hashDecBytes[1]))
				tmpFile := tmpFiles[idx]
				(*tmpFile).write(line)
			}
		}
	})

	// flush tmpFiles and create sorting channel
	tmpFileSortChan := make(chan *TmpFile, max)
	doneChan := make(chan *TmpFile, max)
	pes("Flushing tmp files.\n")
	for i:=0; i<=max; i++ {
		tmpFile := tmpFiles[i]
		(*tmpFile).flush()
	}

	// sort tmp files
	var wg sync.WaitGroup
	wg.Add(*numProcs)
	for i := 0; i < *numProcs; i++ {
		go func(pid int) {
			defer wg.Done()
			for tmpFile := range tmpFileSortChan {
				// sort
				done := len(doneChan)
				todo := len(tmpFileSortChan)
				working := max - (todo + done)
				pes(sprintf("\rSorting %d files, finished %d of %d (%d%%)\t\t\r", working, done, max, (done*100)/max))
				(*tmpFile).sort(*column)

				doneChan <- tmpFile
			}
		}(i)
	}

	close(tmpFileSortChan)
	wg.Wait()
	close(doneChan)

	// append files to stdout
	
}
