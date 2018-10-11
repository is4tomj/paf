package main

import (
	//"io/ioutil"
	"os"
	"sync"
	"bytes"
	"paf/pio"
)

type TmpFile struct {
	path string
	count int
	buff bytes.Buffer
	mux *sync.Mutex

	sortedBuff []byte
}

var tab = []byte("\t")
func (tf *TmpFile)sort(col int) {
	/*
	buff, err := ioutil.ReadAll()
	if err != nil { // according to docs, err should not be EOF if successful
		pes(err.Error())
	}


	lines := bytes.Split(buff, tab)
	*/
}

func (tf *TmpFile)write(b []byte) {
	tf.mux.Lock()
	tf.buff.Write(b)
	tf.buff.WriteByte(pio.Nl)
	tf.count += 1
	if tf.count >= 10 {
		f, err := os.OpenFile(tf.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			pes(err.Error())
			os.Exit(1)
		}
		defer f.Close()
		f.Write(tf.buff.Bytes())
		tf.buff.Reset()
		tf.count = 0
	}
	tf.mux.Unlock()
}

func (tf *TmpFile)flush() {
	tf.mux.Lock()
	if tf.count > 0 {
		f, err := os.OpenFile(tf.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			pes(err.Error())
			os.Exit(1)
		}
		defer f.Close()
		f.Write(tf.buff.Bytes())
		tf.buff.Reset()
		tf.count = 0
	}
	tf.mux.Unlock()
}
