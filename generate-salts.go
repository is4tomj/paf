//////////////////////////////////
// paf gen-salts
//////////////////////////////////

package main

import (
	"os"
	"flag"
	"sync"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/base64"
)

func genSalts() {
	genSaltsFlags := flag.NewFlagSet("gen-salts", flag.ContinueOnError)

	numSalts := genSaltsFlags.Int("num-salts", 100000000, "number of salts to generate")
	saltLength := genSaltsFlags.Int("salt-len", 20, "number of bytes")

	base64Flag := genSaltsFlags.Bool("base64", false, "output salts in base 64 [a-zA-Z0-9+/], default is hex (base16)")
	
	runSize := genSaltsFlags.Int("run-size", 10000, "number of salts to generate per run")
	numProcs := genSaltsFlags.Int("num-procs", 1 , "number of processors")

	if err := genSaltsFlags.Parse(os.Args[2:]); err != nil {
		pes(`Generate salts and print to stdout in PAF format. Each salt is represented as hex string.
Examples:
    $ time paf gen-salts --num-salts 10000000 --num-procs 4 > salts.paf
`)
		os.Exit(1)
	} else if *saltLength < 14 {
		pes("\n\nWARNING! Each salt should be at least 112 bits (14 bytes), per NIST 800-63b.\n\n")
	} else if *numSalts <= 0 || *numProcs <= 0 {
		pes("FUCK! An error occured OR the number of salts and procs must be greater than zero.\n")
		os.Exit(1)
	} else if *numSalts < *runSize {
		*runSize = *numSalts
	} else {
		pes(sprintf("Generating %s salts with %d procs.\n", prettyNum(*numSalts), *numProcs))
	}

	// The number of salts to generate is divided into "runs".
	// Each run is the number of salts that a processor should generate.
	numRuns := (*numSalts) / (*runSize)
	modRuns := (*numSalts) % (*runSize)
	totalRuns := numRuns
	if modRuns > 0 {
		totalRuns += 1
	}
	runChan := make(chan int, totalRuns)
	for i:=0; i<numRuns; i++ {
		runChan <- (*runSize)
	}
	if modRuns > 0 {
		runChan <- modRuns
	}

	encodeFunc := hex.EncodeToString
	if *base64Flag {
		encodeFunc = base64.StdEncoding.EncodeToString
	}
	
	var wg sync.WaitGroup
	wg.Add(*numProcs)
	for j:=0; j<*numProcs; j++ {
		go func(pid int) {
			defer wg.Done()

			saltBuff := make([]byte, (*runSize)*(*saltLength))
			var buff bytes.Buffer;

			for run := range runChan {
				if run > 0 {
					_, err := rand.Read(saltBuff)
					if err != nil {
						pes("FUCK! Something bad happened with getting secure prngs.\n")
						pes(err.Error())
					}

					for i:=0; i<run; i++ {
						start := i*(*saltLength)
						end := start + (*saltLength)
						salt := encodeFunc(saltBuff[start:end])
						buff.Write([]byte(salt+"\n"))
					}
					
					po(buff.Bytes())
					buff.Reset()

					pes(sprintf("\r%d%% finished", ((totalRuns - len(runChan))*100) / numRuns))
				}
			}
		}(j)
	}
	
	close(runChan) // this must come before wg.Wait
	wg.Wait()
	pes("\rComplete.                                  \n")

	
}



