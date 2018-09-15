package main

import (
	"fmt"
	"flag"
	"os"
)


const nl = byte('\n')

var inputPath = flag.String("input", "" , "input path")

var po = os.Stdout.Write
var pe = os.Stderr.Write
var sprintf = fmt.Sprintf

func pes(str string) {
	pe([]byte(str))
}

func main() {
	switch os.Args[1] {
  case "gen-creds":
		genCreds()
	}
	
	pes("Done, Byee\n")
}
