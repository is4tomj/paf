package pio

import (
	"bytes"
	"os"
	"sync"
)

// BufferPool manages a pool of bytes.Buffer objects and supports outputting buffers in a given order
type BufferPool struct {
	pool     chan *bytes.Buffer
	mux      *sync.Mutex
	next     int // index of the next buffer to output
	outBuffs []*bytes.Buffer
	file     *os.File
}

// NewBufferPool will instantiate a new buffer bool that supports ordered output
// if outfile is nil, then outfile will be stdout
// numOutputBuffs is used to define the original number of expected output buffers
func NewBufferPool(poolSize int, numOutputBuffs int, outfile *os.File) *BufferPool {

	bp := &BufferPool{
		pool:     make(chan *bytes.Buffer, poolSize),
		mux:      &sync.Mutex{},
		next:     0,
		outBuffs: make([]*bytes.Buffer, numOutputBuffs),
		file:     outfile,
	}

	if bp.file == nil {
		bp.file = os.Stdout
	}

	return bp
}

// Get will return a pointer to a bytes.Buffer from a pool of bytes.Buffers if available
func (bp *BufferPool) Get() *bytes.Buffer {
	var buff *bytes.Buffer
	select {
	case buff = <-bp.pool:
	default:
		buff = &bytes.Buffer{}
	}
	return buff
}

// Write will write a buffer according to an index, reset the buffer, and return the buffer to the buffer pool
// if index is less than zero, then the output will not be an a particular order
func (bp *BufferPool) Write(buffPtr **bytes.Buffer, index int) {
	bp.mux.Lock()
	defer bp.mux.Unlock()

	buff := *(buffPtr)
	if index < 0 {
		bp.file.Write(buff.Bytes())
	} else {

		// allocate more buffer space if needed
		if lenOutBuffs := len(bp.outBuffs); lenOutBuffs <= index {
			diffSize := (index + 1) - lenOutBuffs
			diffSlice := make([]*bytes.Buffer, diffSize)
			bp.outBuffs = append(bp.outBuffs, diffSlice...)
		}
		bp.outBuffs[index] = buff
		*buffPtr = nil // free the bytes.Buffer pointer so that this buffer can be garbage collected if needed

		// the current buffer is the next to be printed, then print that buffer and any already received buffers immediately following the current buffer
		if index == bp.next {
			lenOutBuffs := len(bp.outBuffs)
			for ; bp.next < lenOutBuffs && bp.outBuffs[bp.next] != nil; bp.next++ {
				bp.file.Write(bp.outBuffs[bp.next].Bytes())
				bp.outBuffs[bp.next].Reset()

				// add buffer back to pool if pool is not at capacity
				select {
				case bp.pool <- bp.outBuffs[bp.next]:
					bp.outBuffs[bp.next] = nil
				default:
					bp.outBuffs[bp.next] = nil
				}
			}
		}
	}
}
