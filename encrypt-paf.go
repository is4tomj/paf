package main

import (
	"os"
	"io/ioutil"
	"flag"
	"bufio"
	"paf/pio"
)

func encrypt() {
	crypFlags := flag.NewFlagSet("cryp", flag.ContinueOnError)
	inputFile := crypFlags.String("input-file", "", "file to read")
	numProcs := crypFlags.Int("num-procs", 1, "number of processors")
	passphraseFlag := crypFlags.String("passphrase", "", "file to read")
	chunkSizeFlag := crypFlags.Int("chunk-size", 100000, "approx size of each block")
	

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
	pio.Process(file, chunkSize*(*chunkSizeFlag), *numProcs, nil, func(pid int, chunk *pio.Chunk) {
		buff, n, err := (*chunk).Bytes(nil)
		
		if err != nil {
			pes(err.Error())
			os.Exit(1)
		} else if n < 0 {
			pes(sprintf("Shit! n is less than zero: %d.", n))
			os.Exit(1)
		} else {
			err := encChunkWrite(buff)
			if err != nil {
				panic(err)
			}
		}
	})

}
