package main

import (
	"flag"
	"io"
	"os"
	"path/filepath"

	"github.com/is4tomj/paf/pio"
)

func pack() {
	packFlags := flag.NewFlagSet("pack", flag.ContinueOnError)
	chunkSizeFlag := packFlags.Int("chunk-size", initChunkSize, sprintf("approx. size of chunks (%d default)", initChunkSize))
	passphraseFlag := packFlags.String("passphrase", "", "passphrase to encrypt paf")
	compressionLevelFlag := packFlags.Int("compress", 0, "level of compression from 1-9, 0 is no compression")
	verboseFlag := packFlags.Bool("v", false, "verbose")

	plaintextOutFlag := packFlags.Bool("plaintext-out", false, "output data in plain text, not PaF format")

	if err := packFlags.Parse(os.Args[2:]); err != nil || len(os.Args[2:]) == 0 {
		pes(`
Pack reads one or more files and outputs the file to standard out in PAF format. This is a serial process to preserve order.

Examples:
  $ paf pack creds1.txt creds2.txt creds3.txt > creds.paf
  $ paf pack --chunk-size=2048000 creds.txt | creds.paf
`)
		os.Exit(1)
	}

	filePaths := packFlags.Args()
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
	var pass *string = nil
	if *passphraseFlag != "" {
		pass = passphraseFlag
	}

	var pafWrite func([]byte) error
	if *plaintextOutFlag == true {
		pafWrite = pio.NewPlainWriter(os.Stdout, totalSize, *verboseFlag)
	} else {
		pafWrite = pio.NewPafWriterV100(os.Stdout, pass, *compressionLevelFlag, totalSize, *verboseFlag)
	}

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
						pafWrite(buff[:i+1]) // +1 because golang is exclusive of upperbound

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

	// flush anything left in the buffer
	if buffOffset > 0 {
		pafWrite(buff[:buffOffset])
	}
	pes("\n")
}
