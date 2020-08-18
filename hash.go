package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"os"
	"strings"

	"github.com/is4tomj/paf/pio"
)

func hexEncode(sum []byte) []byte {
	return []byte(hex.EncodeToString(sum))
}
func base64Encode(sum []byte) []byte {
	return []byte(base64.URLEncoding.EncodeToString(sum))
}
func binEncode(sum []byte) []byte {
	return sum
}

func hash() {
	hashFlags := flag.NewFlagSet("hash", flag.ContinueOnError)
	numProcs := hashFlags.Int("num-procs", 1, "number of processors")
	inputFile := hashFlags.String("file", "", "file to read")
	inputFields := hashFlags.String("input-fields", "usr,pwd", "name of input fields")
	outputFields := hashFlags.String("output-fields", "usr,pwd", "name of output fields")
	chunkSize := hashFlags.Int("chunk-size", initChunkSize, sprintf("approx. size of chunks (%d default)", initChunkSize))
	sha256Fields := hashFlags.String("sha256-fields", "", "run sha256 function on the fields")
	hexEnc := hashFlags.Bool("hex", false, "output hashes encoded in hex (default is binary)")
	base64Enc := hashFlags.Bool("base64", false, "output hashes encoded in base64 (default is binary")

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
		pesf(ep + " No input file specified.\n")
		os.Exit(1)
	} else if *inputFields == "" {
		pesf(ep + " No input fields specified.\n")
		os.Exit(1)
	} else if *outputFields == "" {
		pesf(ep + " No output fields specified.\n")
		os.Exit(1)
	} else if *sha256Fields == "" {
		pesf(ep + " No fields are being hashed.\n")
		os.Exit(1)
	}

	inputNames := strings.Split(*inputFields, ",")
	outputNames := strings.Split(*outputFields, ",")
	sha256Names := strings.Split(*sha256Fields, ",")
	tab := []byte("\t")
	nl := []byte("\n")

	// pick the output encoding
	encode := binEncode // default encoding
	if *hexEnc {
		encode = hexEncode
	} else if *base64Enc {
		encode = base64Encode
	}

	// open the input file
	file, err := os.Open(*inputFile)
	if err != nil {
		pes(err.Error())
		os.Exit(2)
	}
	defer file.Close()

	// crate buffer pool for efficient memory usage and ordered output
	buffPool := pio.NewBufferPool(*numProcs, *numProcs, os.Stdout)

	pio.Process(file, *chunkSize, *numProcs, nil, func(pid int, chunk *pio.Chunk) {
		buff, n, err := (*chunk).Bytes(nil)
		if err != nil {
			pes(err.Error())
			os.Exit(1)
		} else if n < 0 {
			pesf(ep+" n is less than zero: %d.", n)
			os.Exit(1)
		} else {
			w := buffPool.Get()
			scan := pio.NewLineScanner(buff)
			for line, lineLen := scan(); lineLen > 0; line, lineLen = scan() {
				// split and parse line
				res := make(map[string][]byte)
				for i, str := range bytes.Split(line, tab) {
					res[inputNames[i]] = str
				}

				// generate hashes
				for _, str := range sha256Names {
					sum := sha256.Sum256(res[str])
					res["sha256"+str] = encode(sum[:])
				}

				// generate output line
				out := make([][]byte, len(outputNames))
				for i, str := range outputNames {
					out[i] = res[str]
				}

				w.Write(bytes.Join(out, tab))
				w.Write(nl)
			}
			buffPool.Write(&w, chunk.Index)
		}
	})

	pes("\rComplete.                                  \n")

}
