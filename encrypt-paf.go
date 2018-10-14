package main

import (
	"os"
	"io/ioutil"
	"flag"
	"bufio"
	"bytes"
	"paf/pio"
	"compress/flate"
)

func encrypt() {
	crypFlags := flag.NewFlagSet("cryp", flag.ContinueOnError)
	inputFile := crypFlags.String("input-file", "", "file to read")
	numProcs := crypFlags.Int("num-procs", 1, "number of processors")
	passphraseFlag := crypFlags.String("passphrase", "", "file to read")
	chunkSizeFlag := crypFlags.Int("chunk-size", 100000, "approx size of each block")
	compressLevelFlag := crypFlags.Int("compress", 0, "change from 1 (best speed) to 9 (best compression)")

	if err := crypFlags.Parse(os.Args[2:]); err != nil || len(os.Args[2:]) == 0 || *inputFile == "" {
		pes(`
Enc lets you encrypt a paf using Go's GCM cipher. The passphrase can be passed using STDIN. The result is printed to STDOUT.

Examples:
  $ echo "passphrase" | paf enc "cheeseflakes" --input-file plaintext.paf > ciphertext.epaf
  $ paf enc --passphrase "cheeseflakes" --input-file plaintext.paf > ciphertext.epaf
`)
		os.Exit(1)
	}


	passphrase := *passphraseFlag
	if passphrase == "" {
		stdin, err := ioutil.ReadAll(bufio.NewReader(os.Stdin))
		if err != nil {
			panic(err.Error())
		}
		passphrase = string(stdin)
	}

	file, err := os.Open(*inputFile)
	if err != nil {
		pes(err.Error())
		os.Exit(1)
	}
	defer file.Close()

	encChunkWrite := pio.NewEncryptedPafWriter(os.Stdout, passphrase)

	pes("Starting to process\n")
	buffers := make([]*bytes.Buffer, *numProcs)
	writers := make([]flate.Writer, *numProcs)
	pio.Process(file, chunkSize*(*chunkSizeFlag), *numProcs, func(pid int, fileSize int64) {
		var b bytes.Buffer
		buffers[pid] = &b
		w, err := flate.NewWriter(&b, *compressLevelFlag)
		if err != nil {
			panic(err)
		}
		writers[pid] = *w
		
	}, func(pid int, chunk *pio.Chunk) {
		buff, n, err := (*chunk).Bytes(nil)
		
		if err != nil {
			pes(err.Error())
			os.Exit(1)
		} else if n < 0 {
			pes(sprintf("Shit! n is less than zero: %d.", n))
			os.Exit(1)
		} else {

			w := writers[pid]
			b := buffers[pid]
			b.Reset()
			w.Reset(b)
			w.Write(buff)
			w.Flush() // this is a must
			res := (*b).Bytes()

			// encrypt and print to stdout
			err := encChunkWrite(res)
			if err != nil {
				panic(err)
			}

		}
	})

}
