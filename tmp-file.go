package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"sort"
	"sync"
)

type tmpFile struct {
	Index int
	path  string
	buff  *bytes.Buffer
	mux   *sync.Mutex
	count *int
}

type line struct {
	buff      []byte
	compBytes []byte
}

var tab = byte("\t"[0])
var tabs = []byte("\t")
var nl = byte("\n"[0])

// Sort will sort lines based on a particular column
// Sort assumes that each line is valid and is not blank
func (tf *tmpFile) Sort(col int, uniq bool, decodeFunc func([]byte, []byte) (int, error)) (*bytes.Buffer, int) {
	buff, err := ioutil.ReadFile(tf.path)
	if err != nil { // according to docs, err should not be EOF if successful
		if os.IsNotExist(err) {
			return nil, 0
		}
		panic(err.Error())
	}

	// find num lines and check for issues
	lineCount := 0
	prevEoL := -1
	for i, b := range buff {
		if b == nl {

			// check for blank line
			if prevEoL == i-1 {
				panic(errors.New(ep + " line is blank"))
			}
			prevEoL = i

			// if no errors update line count
			lineCount++
		}
	}

	// parse lines
	lines := make([]line, lineCount)
	lIdx, lStart := 0, 0
	for i, b := range buff {
		if b == nl {
			// get entire line
			lineBuff := buff[lStart:i]
			lineBuffLen := len(lineBuff)
			// get bytes to compare
			cIdx, cStart, cEnd := 0, 0, 0
			for j, cb := range lineBuff {
				if cb == tab || j == lineBuffLen-1 {
					// get entire column if this column is the column we care about
					if cIdx == col {
						cEnd = j

						// if the last char is a EOL, then we need to encrement the cuttoff
						if j == lineBuffLen-1 {
							cEnd++
						}

						break // we don't care about other columns
					}

					// update column index
					cStart = j + 1
					cIdx++
				}
			}

			compBytes := make([]byte, 64)
			n, err := decodeFunc(compBytes, lineBuff[cStart:cEnd])
			if err != nil {
				pes(sprintf("\n\nSHIT! We got this string as input:%s\n\n", lineBuff[cStart:cEnd]))
				panic(err)
			}

			// create line
			lines[lIdx] = line{buff: lineBuff, compBytes: compBytes[:n]}

			// prepare for next line
			lIdx++
			lStart = i + 1
		}
	}

	numLines := len(lines)
	if numLines != tf.Count() {
		panic(sprintf("FUCK! %s sorted %d but originally %d\n", tf.path, numLines, tf.Count()))
	}

	// sort lines
	sort.Slice(lines, func(i, j int) bool {
		return bytes.Compare(lines[i].compBytes, lines[j].compBytes) == -1
	})

	// create and return new buff that is sorted
	var sortedBuff bytes.Buffer
	if uniq {
		count := 0
		if numLines > 0 {
			count++
			sortedBuff.Write(lines[0].buff)
			sortedBuff.WriteByte(nl)
			for i := 1; i < numLines; i++ {
				if !bytes.Equal(lines[i-1].compBytes, lines[i].compBytes) {
					sortedBuff.Write(lines[i].buff)
					sortedBuff.WriteByte(nl)
					count++
				}
			}
		}
		return &sortedBuff, count
	} else {
		for _, line := range lines {
			sortedBuff.Write(line.buff)
			sortedBuff.WriteByte(nl)
		}
		return &sortedBuff, numLines
	}
}

func (tf *tmpFile) Count() int {
	return *(tf.count)
}

// write will put the bites in a buffer for the tmp file
func (tf *tmpFile) Write(b []byte) {
	tf.mux.Lock()
	tf.buff.Write(b)
	tf.buff.WriteByte(nl)
	*(tf.count)++
	if tf.buff.Len() > 1024*10 { // flush if buff is getting bigger than a 10KB
		tf.Flush()
	}
	tf.mux.Unlock()
}

// flush writes the data in buff to disk.
// WARNGING: flush is not a thread-safe function!!!
func (tf *tmpFile) Flush() {
	if tf.buff.Len() > 0 {
		f, err := os.OpenFile(tf.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		f.Write(tf.buff.Bytes())
		f.Close()
		(*(tf.buff)).Reset()
	}
}
