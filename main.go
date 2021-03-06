package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

const initChunkSize = 16777216 // 2^24
const ep = "ლ(ಠ益ಠლ)"

var po = os.Stdout.Write
var pe = os.Stderr.Write
var sprintf = fmt.Sprintf

func pes(str string) {
	pe([]byte(str))
}
func pesf(format string, args ...interface{}) {
	pe([]byte(sprintf(format, args...)))
}
func newErr(str string) error {
	return errors.New(ep + str)
}

func main() {

	// cpu profile stuff
	cpuprofile := os.Getenv("CPUPROF")
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	// content of program
	if len(os.Args) < 2 {
		pes(`
gen-creds: generate usernames-password pairs
gen-salts: generate salts
pack
cc
hash
sort
enc
dec
`)
	} else {
		switch os.Args[1] {
		case "pack":
			pack()
		case "cc":
			cc()
		case "hash":
			hash()
		case "sort":
			sortByHash()
		case "enc":
			encrypt()
		case "dec":
			decrypt()
		case "gen-creds":
			genCreds()
		case "gen-salts":
			genSalts()
		}

	}
	pes("Done, byee\n")

	// memory profile stuff
	memprofile := os.Getenv("MEMPROF")
	if memprofile != "" {
		f, err := os.Create(memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		f.Close()
	}

}
