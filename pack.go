package main

import (
	"flag"
	"io"
	"os"
	"path/filepath"
)

func cat() {
	catFlags := flag.NewFlagSet("cryp", flag.ContinueOnError)
	chunkSizeFlag := catFlags.Int("chunk-size", initChunkSize, sprintf("approx. size of chunks (%d default)", initChunkSize))
	verboseFlag := catFlags.Bool("v", false, "verbose")

	if err := catFlags.Parse(os.Args[2:]); err != nil || len(os.Args[2:]) == 0 {
		pes(`
Cat reads a file and outputs the file to standard out in PAF format

Examples:
  $ paf cat creds1.txt creds2.txt creds3.txt > creds.paf
  $ paf cat --num-procs=4 --chunk-size=2048000 creds.txt | creds.paf
`)
		os.Exit(1)
	}

	filePaths := catFlags.Args()
	if len(filePaths) == 0 {
		pes("No file names or paths given\n")
		os.Exit(1)
	}

	// expand all the paths and calculate total size of raw data
	var allFilePaths []string
	totalSize := int64(0)
	for _, path := range filePaths {
		matchedFiles, err := filepath.Glob(path)
		if err != nil {
			panic(err)
		}
		for _, mf := range matchedFiles {
			stat, err := os.Stat(mf)
			if err != nil {
				panic(nil)
			}
			if stat.Mode().IsRegular() {
				allFilePaths = append(allFilePaths, mf)
				totalSize += stat.Size()
			}
		}
	}

	pes(sprintf("chunk-size:%d\n", *chunkSizeFlag))
	buff := make([]byte, *chunkSizeFlag)
	buffOffset := 0
	fileOffset := int64(0)
	bytesRead := int64(0)
	for _, path := range allFilePaths {
		f, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		// closing file at end of loop

		for true {
			n, err := f.ReadAt(buff[buffOffset:], fileOffset)
			if err != nil && err != io.EOF {
				panic(err)
			}
			buffOffset += n
			fileOffset += int64(n)

			if buffOffset == *chunkSizeFlag {
				// find last \n character
				for i := buffOffset - 1; true; i-- {
					if i < 0 {
						panic("A record is bigger than a chunk\n")
					}

					if buff[i] == nl {
						// output buffer until the last \n char
						os.Stdout.Write(buff[:i+1]) // +1 because golang is exclusive of upperbound

						if *verboseFlag {
							// +1 because golang is exclusive of upper bound
							// +1 because i starts at buffOffset-1
							bytesRead += int64(i + 1)
							pes(sprintf("\r%d%% %d of %d bytes", 100*bytesRead/totalSize, bytesRead, totalSize))
						}

						// copy remaining bytes in buff to the front of buff
						copy(buff, buff[i+1:])
						buffOffset = (*chunkSizeFlag) - (i + 1)
						break
					}
				}
			}

			// check if at EOF but buffer is not full
			if err == io.EOF {
				fileOffset = 0
				break
			}
		}

		// do not defer this
		f.Close()
	}

	os.Stdout.Write(buff[:buffOffset]) // flush anything left in the buffer

	if *verboseFlag {
		// update status
		bytesRead += int64(buffOffset)
		pes(sprintf("\r%d%% %d of %d bytes\n", 100*bytesRead/totalSize, bytesRead, totalSize))
	}
}
