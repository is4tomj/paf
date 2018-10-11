package main

import (
	"flag"
	"os"
	"strings"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/base64"
	"paf/pio"
)

const chunkSize = 2^24

func hash() {
	hashFlags := flag.NewFlagSet("hash", flag.ContinueOnError)
	numProcs := hashFlags.Int("num-procs", 1, "number of processors")
	inputFile := hashFlags.String("file", "", "file to read")
	inputFields := hashFlags.String("input-fields", "usr,pwd", "name of input fields")
	outputFields := hashFlags.String("output-fields", "usr,pwd", "name of output fields")

	sha256Fields := hashFlags.String("sha256-fields", "", "run sha256 function on the fields")
	hexEnc := hashFlags.Bool("hex", false, "output hashes encoded in hex")
	base64Enc := hashFlags.Bool("base64", false, "output hashes encoded in base64")
	binaryEnc := hashFlags.Bool("bin", true, "output hashes encoded in bytesn")

	if err := hashFlags.Parse(os.Args[2:]); err != nil || len(os.Args[2:]) == 0 {
		pes(`
Hash one or more columns in paf file and print to stdout.
Examples:
    $ paf hash --num-procs 4 --file input.paf \
               --input-fields="usr,pwd" \
               --sha256-fields="usr" \
               --output-fields="sha2usr,usr,pwd" > test.paf
`)
		os.Exit(1)
	} else if *inputFile == "" {
		pes(pio.Ep + " No input file specified.\n")
		os.Exit(1)
	} else if *inputFields == "" {
		pes(pio.Ep + " No input fields specified.\n")
		os.Exit(1)
	} else if *outputFields == "" {
		pes(pio.Ep + " No output fields specified.\n")
		os.Exit(1)
	} else if *sha256Fields == "" {
		pes(pio.Ep + " No fields are being hashed specified.\n")
		os.Exit(1)
	}

	inputNames := strings.Split(*inputFields, ",")
	outputNames := strings.Split(*outputFields, ",")
	sha256Names := strings.Split(*sha256Fields, ",")
	tab := []byte("\t")
	nl := []byte("\n")

	file, err := os.Open(*inputFile)
	if err != nil {
		pes(err.Error())
		os.Exit(2)
	}
	defer file.Close()

	
	pio.Process(file, chunkSize*100000, *numProcs, nil, func(pid int, chunk *pio.Chunk) {
		buff, n, err := (*chunk).Bytes(nil)
		if err != nil {
			pes(err.Error())
			os.Exit(1)
		} else if n < 0 {
			pes(sprintf("Shit! n is less than zero: %d.", n))
			os.Exit(1)
		} else {
			var w bytes.Buffer
			scan := pio.NewLineScanner(buff)
			for line, lineLen := scan(); lineLen > 0; line, lineLen = scan() {
				// split and parse line
				res := make(map[string][]byte)
				for i, str := range bytes.Split(line,tab) {
					res[inputNames[i]]=str
				}

				// generate hashes
				for _, str := range sha256Names {
					sum := sha256.Sum256(res[str])
					if *hexEnc {
						res["sha256"+str] = []byte(hex.EncodeToString(sum[:]))
					} else if *base64Enc {
						res["sha256"+str] = []byte(base64.URLEncoding.EncodeToString(sum[:]))
					} else if *binaryEnc {
						res["sha256"+str] = sum[:]
					}
				}

				// generate output line
				out := make([][]byte, len(outputNames))
				for i, str := range outputNames {
					out[i] = res[str]
				}
				
				w.Write(bytes.Join(out,tab))
				w.Write(nl)
			}
			os.Stdout.Write(w.Bytes())
			w.Reset()
		}
	})
	
	pes("\rComplete.                                  \n")

}
