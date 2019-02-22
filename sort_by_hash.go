package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"paf/pio"
	"sync"
	"time"
)

func deleteTmpDir(path string) {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		os.RemoveAll(path)
	}
}

func sortByHash() {
	sortFlags := flag.NewFlagSet("hash", flag.ContinueOnError)
	numProcs := sortFlags.Int("num-procs", 1, "number of processors")
	chunkSize := sortFlags.Int("chunk-size", initChunkSize, sprintf("approx. size of chunks (%d default)", initChunkSize))
	inputFile := sortFlags.String("input-file", "", "file to read")
	column := sortFlags.Int("col", 0, "zero-based column number with hash to sort by")

	skipPresortFlag := sortFlags.Bool("skip-presort", false, "do not presort (tmp files already generated)")

	tmpPath := sortFlags.String("tmp-dir", ".tmp", "directory to temporarily store files during sort")

	base64Enc := sortFlags.Bool("base64", false, "output hashes encoded in base64 instead of hex (default)")

	if err := sortFlags.Parse(os.Args[2:]); err != nil || len(os.Args[2:]) == 0 || *inputFile == "" {
		pes(`
Sort a column in a paf file with hash and print to stdout.
Examples:
    $ paf sort --num-procs 4 --input-file input.paf --col=0 --tmp-falsedir data/tmp
`)
		os.Exit(1)
	}

	file, err := os.Open(*inputFile)
	if err != nil {
		pes(err.Error())
		os.Exit(1)
	}
	defer file.Close()

	// delete (if needed) and make tmp dir
	pes(sprintf("Preparing tmp directory:%s\n", (*tmpPath)))
	deleteTmpDir(*tmpPath)
	os.MkdirAll(*tmpPath, os.ModePerm)

	// initialize tmpfiles
	const max = 0xFFFF
	tmpFiles := make([]*tmpFile, max+1)
	for i := 0; i <= max; i++ {
		path := (*tmpPath) + sprintf("/%04x", uint16(i))
		count := 0
		tmpFiles[i] = &tmpFile{Index: i, path: path, buff: &bytes.Buffer{}, mux: &sync.Mutex{}, count: &count}
	}

	// set decode function for comparisons
	decodeFunc := hex.Decode
	if *base64Enc {
		decodeFunc = base64.URLEncoding.Decode
	}

	start := time.Now()
	if *skipPresortFlag {
		pes("Skipping presort\n")

	} else {
		pes("Presorting into tmp files.\n")
		pio.Process(file, *chunkSize, *numProcs, nil, func(pid int, chunk *pio.Chunk) {
			buff, n, err := (*chunk).Bytes(nil)

			if err != nil {
				pes(err.Error())
				os.Exit(1)
			} else if n < 0 {
				pes(sprintf("Shit! n is less than zero: %d.", n))
				os.Exit(1)
			} else {
				// traverse lines
				scan := pio.NewLineScanner(buff)
				hashDecBytes := make([]byte, 64)
				for line, lineLen := scan(); lineLen > 0; line, lineLen = scan() {

					// read line
					hashEncBytes := bytes.Split(line, []byte("\t"))[*column]
					n, err := decodeFunc(hashDecBytes, hashEncBytes)
					if err != nil {
						pesf("\n\nn: %d\n\n", n)
						panic(err)
					}
					// get the first (most significant) two bytes
					idx := int32(uint16(hashDecBytes[0])<<8 | uint16(hashDecBytes[1]))
					tf := *(tmpFiles[idx])
					tf.Write(line)
				}
			}
		})
	}

	// flush tmpFiles and create sorting channel
	tmpFileSortChan := make(chan *tmpFile, max+1)
	doneChan := make(chan *tmpFile, max+1)

	pes("Flushing tmp files.\n")
	for i := 0; i <= max; i++ {
		tmpFile := tmpFiles[i]
		if *skipPresortFlag == false {
			(*tmpFile).Flush()
		}
		tmpFileSortChan <- tmpFile
	}
	pes(sprintf("  finished in %.2f minutes.\n", time.Now().Sub(start).Minutes()))

	// sort tmp files
	start = time.Now()
	for i := 0; i < *numProcs; i++ {
		go func(pid int) {
			for tmpFile := range tmpFileSortChan {
				// sort
				tf := *tmpFile
				sortedBuff, n := tf.Sort(*column, decodeFunc)
				if n > 0 {
					if err = ioutil.WriteFile(tf.path+".sorted", sortedBuff.Bytes(), 0644); err != nil {
						panic(err)
					}

					// delete original tmp file
					if err = os.Remove(tf.path); err != nil {
						panic(err)
					}
				}

				doneChan <- tmpFile
			}
		}(i)
	}
	// close here so that Go runtime does not think there is a deadlock issue
	close(tmpFileSortChan)

	// This next part runs in the background while the tmp files are being sorted
	// and outputs the next sorted tmp file in order.
	doneBuff := make([]*tmpFile, max+1)
	idx := 0
	for i := 0; i <= max; i++ {
		doneTF := <-doneChan
		doneBuff[doneTF.Index] = doneTF

		for ; idx <= max && doneBuff[idx] != nil; idx++ {
			pes(sprintf("\rSorting %d%%", (idx*100)/max))

			tf := doneBuff[idx]

			sortFileName := tf.path + ".sorted"
			_, err := os.Stat(sortFileName)
			if os.IsExist(err) {
				in, err := os.Open(sortFileName)
				if err != nil {
					panic(err)
				}

				_, err = io.Copy(os.Stdout, in)
				if err != nil {
					panic(err)
				}
				in.Close()

				// delete sorted tmp file
				if err = os.Remove(sortFileName); err != nil {
					panic(err)
				}
			}
		}
	}
	close(doneChan)

	pes(sprintf("\n  finished in %.2f minutes.\n", time.Now().Sub(start).Minutes()))

	deleteTmpDir(*tmpPath)
}
