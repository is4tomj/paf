package main

import (
	"flag"
	"io"
	"os"
	"paf/pio"
)

func cc() {
	ccFlags := flag.NewFlagSet("cc", flag.ContinueOnError)
	pathFlag := ccFlags.String("path", "-", "path to paf file, read from STDIN if - (default)")
	//charFlag := ccFlags.String("c", "\n", "character to count (default is \\n)")

	if err := ccFlags.Parse(os.Args[2:]); err != nil {
		pes(`
CC reads a PAF from STDIN, outputs the frequency a particular character to STDERR, and passes the input to STDOUT

Examples:
  $ paf pack creds.txt | paf cc > /dev/null # count newline characters in paf file 
  $ paf pack creds.txt | paf cc -c $'\t' > creds.paf # count tab characters in creds.txt
  $ cat creds.paf | paf cc > /dev/null # count newline characters in paf file
`)
		os.Exit(1)
	}

	f := os.Stdin
	if *pathFlag != "-" {
		file, err := os.Open(*pathFlag)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		f = file
	}

	pafReader, err := pio.NewPafReader(f, nil)
	if err != nil {
		panic(err)
	}

	count := uint64(0)
	workingBuff := make([]byte, 1024)
	for {
		buff, _, err := pafReader(&workingBuff)
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		lenBuff := len(buff)
		for i := 0; i < lenBuff; i++ {
			if buff[i] == nl {
				count++
			}
		}
		os.Stdout.Write(buff)
	}

	pes(sprintf("\ncount: %d\n", count))
}
