package pio

import (
	"fmt"
	"os"
	"sync"
	"errors"
)

const nl = byte('\n')
const ep = "ლ(ಠ益ಠლ)"

var pe = os.Stderr.Write
var sprintf = fmt.Sprintf
func pes(str string) {
	pe([]byte(str))
}
func pesf(format string, args ...interface{}) {
	pe([]byte(sprintf(format, args)))
}

type Paf struct {
	header *Header
	blocks []*Block
	mux sync.Mutex
	file *os.File
	currFileOffset uint64
	currBlockIdx uint64

}

type Block struct {
	offset uint64
	size uint32
	buffer []byte
}


func NewPaf(file *os.File, numBlocks uint64) (*Paf, error) {
	if file == nil {
		return nil, errors.New("File cannot be nil")
	} else if numBlocks == 0 {
		return nil, errors.New("Must have more than one block.")
	}
	
	p := new(Paf)
	p.header = NewHeader(numBlocks)
	p.blocks = make([]*Block, numBlocks)
	p.mux = sync.Mutex{}
	p.file = file
	p.currFileOffset = uint64(0)
	p.currBlockIdx = uint64(0)
	
	return p, nil
}

func (p Paf)CommitNewBlock(b *Block) error {
	p.mux.Lock()
	
	n, err := p.file.Write(b.buffer)
	if err != nil {
		p.blocks[p.currBlockIdx] = &Block{
			offset: p.currFileOffset,
			size: uint32(n)}
		p.currBlockIdx++
		p.currFileOffset += uint64(n)
	}
	
	p.mux.Unlock()

	if err != nil {
		return err
	}
	return nil
}

func (p Paf)Finish() error {
	p.mux.Lock()

	//headerOffset := []byte(p.currFileOffset)
	
	
	p.mux.Unlock()
	
	return nil
}

var inputMutex = &sync.Mutex{}
func AtomicWrite(buf []byte, fd *os.File, mutex *sync.Mutex) (int, error) {
	if fd == nil {
		fd = os.Stdout
	}
	if mutex == nil {
		mutex = inputMutex
	}

	mutex.Lock()
	n, err := fd.Write(buf)
	mutex.Unlock()

	return n, err
}
