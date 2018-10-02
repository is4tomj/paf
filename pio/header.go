package pio

import (
)

type Header struct {
	version uint32
	entries []uint64
	numBlocks uint64
	size uint64
}

func NewHeader(numBlocks uint64) *Header {
	h := new(Header)
	h.version = 1
	h.entries = make([]uint64, numBlocks)
	h.numBlocks = numBlocks
	h.size =
		8 /*sizeof offset*/ +
		8 /*sizeof numBlocks*/ +
		4 /*sizeof version*/ +
		(numBlocks*8) /*sizeof all block offsets*/
	
	return h
}
