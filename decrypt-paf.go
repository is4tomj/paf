package main

import (
	"os"
	"io"
	"io/ioutil"
	"flag"
	"bufio"
	"paf/pio"
	"compress/flate"
	"bytes"
)

func decrypt() {
	crypFlags := flag.NewFlagSet("dec", flag.ContinueOnError)
	inputFile := crypFlags.String("input-file", "", "file to read")
	numProcs := crypFlags.Int("num-procs", 1, "number of processors")
	passphraseFlag := crypFlags.String("passphrase", "", "file to read")

	if err := crypFlags.Parse(os.Args[2:]); err != nil || len(os.Args[2:]) == 0 || *inputFile == "" {
		pes(`
Dec lets you decrypt a paf using Go's GCM cipher. The passphrase can be passed using STDIN. The result is printed to STDOUT.

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

	pes("Starting to process\n")
	pio.ProcessEncryptedPaf(file, passphrase, *numProcs, nil, func(pid int, buff []byte) {
		n := len(buff)
		if n <= 0 {
			pes(sprintf("Shit! n is less than zero: %d.", n))
			os.Exit(1)
		}

		// deflate
		b := bytes.NewReader(buff)
		r := flate.NewReader(b)
		io.Copy(os.Stdout, r)
		r.Close()
	})

}
