package pio

import (
	"bytes"
	"io"
	"os"
	"sync"
	"testing"
)

func TestBufferPoolWithOrderedOutput(t *testing.T) {

	// setup pipe to test output that would normall y go to a file.
	r, w, _ := os.Pipe()

	bp := NewBufferPool(2, 2, w)

	numBuffs := 10
	buffs := make([]*bytes.Buffer, numBuffs)
	for i := 0; i < numBuffs; i++ {
		buffs[i] = bp.Get()
	}

	var wg sync.WaitGroup
	wg.Add(numBuffs)

	for i, buff := range buffs {
		go func(b *bytes.Buffer, idx int) {
			defer wg.Done()
			b.Write([]byte(sprintf("buffer %d\n", idx)))
			bp.Write(&b, idx)
		}(buff, i)
	}

	wg.Wait()

	// get the output from the pipe and close the file descriptor: w
	resBuff := &bytes.Buffer{}
	w.Close()
	io.Copy(resBuff, r)

	expRes := `buffer 0
buffer 1
buffer 2
buffer 3
buffer 4
buffer 5
buffer 6
buffer 7
buffer 8
buffer 9
`

	if resBuff.String() != expRes {
		t.Fatalf("Expected:\n%s\n\nBut, got:\n%s", expRes, resBuff.String())
	}
}
