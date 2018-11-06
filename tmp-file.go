package main

import (
	"io/ioutil"
	"os"
	"sync"
	"bytes"
	"errors"
	"sort"
)

type TmpFile struct {
	path string
	buff *bytes.Buffer
	mux *sync.Mutex
	count *int
}

type Line struct {
	buff []byte
	compBytes []byte
}

var tab = byte("\t"[0])
var nl = byte("\n"[0])

// Sort will sort lines based on a particular column
// Sort assumes that each line is valid and is not blank
func (tf *TmpFile)Sort(col int, decodeFunc func([]byte,[]byte)(int,error)) (*bytes.Buffer, int) {
	buff, err := ioutil.ReadFile(tf.path)
	if err != nil { // according to docs, err should not be EOF if successful
		panic(err.Error())
	}


	// find num lines and check for issues
	lineCount := 0
	prevEoL := -1
	for i, b := range buff {
		if b == nl {
			
			// check for blank line
			if prevEoL == i-1 {
				panic(errors.New("Line is blank.\n"))
			}
			prevEoL = i
			
			// if no errors update line count
			lineCount++
		}
	}
	
	// parse lines
	lines := make([]Line, lineCount)
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
						if j == lineBuffLen - 1 {
							cEnd++ 
						}

						break // we don't care about other columns
					} 
					
					// update column index
					cStart = j + 1
					cIdx++
				}
			}
			compBytes := make([]byte, 32)
			n, err := decodeFunc(compBytes, lineBuff[cStart:cEnd])
			if n != 32 {
				panic(errors.New(sprintf("Fuck! This key is not long enough (%d chars)", n)))
			}
			if err != nil {
				pes(sprintf("\n\nSHIT! We got this as the hex string:%s\n\n", lineBuff[cStart:cEnd]))
				panic(err)
			}

			// create line
			lines[lIdx] = Line{buff: lineBuff, compBytes: compBytes}

			// prepare for next line
			lIdx++
			lStart = i+1
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
	for _, line := range lines {
		sortedBuff.Write(line.buff)
		sortedBuff.WriteByte(nl)
	}

	return &sortedBuff, numLines
}

func (tf TmpFile)Count() int {
	return *(tf.count)
}

// write will put the bites in a buffer for the tmp file
func (tf TmpFile)Write(b []byte) {
	tf.mux.Lock()
	tf.buff.Write(b)
	tf.buff.WriteByte(nl)
	*(tf.count) += 1
	if tf.buff.Len() > 1024 * 10 { // flush if buff is getting bigger than a 10KB
		tf.Flush()
	}
	tf.mux.Unlock()
}

// flush writes the data in buff to disk.
// WARNGING: flush is not a thread-safe function!!!
func (tf TmpFile)Flush() {
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
