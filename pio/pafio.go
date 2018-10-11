package pio

import (
	"fmt"
	"os"
	"sync"
)

const Nl = byte('\n')
const Tl = byte('\n')
const Ep = "ლ(ಠ益ಠლ)"

var pe = os.Stderr.Write
var sprintf = fmt.Sprintf

func pes(str string) {
	pe([]byte(str))
}
func pesf(format string, args ...interface{}) {
	pe([]byte(sprintf(format, args)))
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
