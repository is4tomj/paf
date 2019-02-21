package pio

// Header is a structure that defines how a paf file is stored
type Header struct {
	version   uint32
	entries   []uint64
	numBlocks uint64
	size      uint64
}

// NewHeader will create a Header struct that can be used to create a paf file
func NewHeader(numBlocks uint64) *Header {
	h := new(Header)
	h.version = 1
	h.entries = make([]uint64, numBlocks)
	h.numBlocks = numBlocks
	h.size = 8 /*sizeof offset*/ +
		8 /*sizeof numBlocks*/ +
		4 /*sizeof version*/ +
		(numBlocks * 8) /*sizeof all block offsets*/

	return h
}
