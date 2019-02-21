package pio

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

const nl = byte('\n')
const ep = "ლ(ಠ益ಠლ)"

var pe = os.Stderr.Write
var sprintf = fmt.Sprintf

func pes(str string) {
	pe([]byte(str))
}
func pesf(format string, args ...interface{}) {
	pe([]byte(sprintf(format, args...)))
}
func newErr(str string) error {
	return errors.New(ep + str)
}

// NewPafReader determines the version, encryptionFlag, compression level used.
func NewPafReader(f *os.File, passphrase *string) (func(*[]byte) ([]byte, uint, error), error) {
	hbuff := make([]byte, 7)
	_, err := f.Read(hbuff)
	if err != nil {
		panic(err)
	}

	if string(hbuff[:5]) == "pv100" {
		// get flags
		enc := uint8(hbuff[5])
		if enc == uint8(1) && passphrase == nil {
			return nil, newErr("no password to decrypt this paf file\n")
		}
		//comp := uint8(hbuff[6])

		// setup counts, offsets, counts, etc.
		count := uint(0)
		offset := int64(7)
		lenBuf := make([]byte, 4)
		mux := &sync.Mutex{}

		// setup decryption if needed
		var key [32]byte
		if passphrase != nil {
			passBytes := []byte(*passphrase)
			key = CreateKey(passBytes)
		} else {
			if enc == uint8(1) {
				panic("Fuck! This PaF file is encrypted, but no passphrase was provided.\n")
			}
		}

		var doneBuff bytes.Buffer

		// return the function
		return (func(readBuff *[]byte) ([]byte, uint, error) {
			// lock closure resources
			mux.Lock()
			defer mux.Unlock()

			// Find out how big the next chunk is
			n, err := f.Read(lenBuf)
			if err != nil {
				if err == io.EOF {
					return nil, count, err
				}
				panic(err)
			}

			// get length of the next chunk raw and uncompressed, and bump offset
			clen := binary.LittleEndian.Uint32(lenBuf)
			offset += 4

			// create raw data buffer that might be compressed or encrypted
			var cbuff []byte
			if readBuff == nil {
				cbuff = make([]byte, clen)
			} else {
				if len(*readBuff) < int(clen) {
					(*readBuff) = make([]byte, clen)
				}
				cbuff = (*readBuff)[:clen]
			}

			// get chunk and check if error or EOF
			cn, cerr := f.Read(cbuff)
			if cerr != nil {
				panic(cerr)
			}

			// make sure we get all the bytes we wanted to get
			if cn != int(clen) {
				panic(sprintf("Fuck! Reading block: Expected %d bytes, but got %d bytes\n", clen, cn))
			}

			// decrypt if needed
			if enc == uint8(1) {
				cbuff = Decrypt(cbuff, key)
			}

			// decompress
			flateReader := flate.NewReader(bytes.NewReader(cbuff))
			defer flateReader.Close()
			doneBuff.Reset()
			io.Copy(&doneBuff, flateReader)

			// update count and offsets
			offset += int64(n)
			count++

			// return chunk and error if error is EOF
			return doneBuff.Bytes(), count, nil
		}), nil
	}

	return nil, newErr("this is not a paf file")
}

// NewPafWriterV100 generates v1 PaF files.
// Format: version (5 bytes), crypto version (1 byte), compression version (1 byte), chunks.
// Each chunk is preceeded with a uint32 that indicates the length of the following chunk.
// If phasephrass is nil, then chunks are not encrypted.
// If compression level is 0, no compression
func NewPafWriterV100(file *os.File, passphrase *string, compressionLevel int, totalSize int64, verbose bool) func([]byte) error {
	f := file
	if file == nil {
		f = os.Stdout
	}

	// init values for first block written
	count := uint64(0)
	numBytes := int64(0)
	lenbuf := make([]byte, 4)
	mux := &sync.Mutex{}

	// setup encryption
	encryptFlag := 0
	var key [32]byte
	if passphrase != nil {
		encryptFlag = 1
		key = CreateKey([]byte(*passphrase))
	}

	// setup compression
	var flateBuffer bytes.Buffer
	flateWriter, err := flate.NewWriter(&flateBuffer, compressionLevel)
	if err != nil {
		panic(err)
	}

	// write the header of the file
	f.Write([]byte("pv100"))
	headerBytes := []byte{byte(uint8(encryptFlag)), byte(uint8(compressionLevel))}
	f.Write(headerBytes)

	return func(data []byte) error {
		// this line should be first
		mux.Lock()
		defer mux.Unlock()

		if count > maxChunks {
			panic("Fuck! Tried to write more than 2^32 chunks, which is a problem for encrypted PaF files.\n")
		}

		// compress if needed
		flateBuffer.Reset()
		flateWriter.Reset(&flateBuffer)
		flateWriter.Write(data)
		flateWriter.Flush()
		outBytes := flateBuffer.Bytes()

		// encrypt if needed
		if passphrase != nil {
			outBytes = Encrypt(outBytes, key)
		}

		// output results by appending to file
		clen := uint32(len(outBytes)) // length of compressed data
		binary.LittleEndian.PutUint32(lenbuf, clen)
		f.Write(lenbuf) // 4 bytes
		f.Write(outBytes)

		if verbose {
			numBytes += int64(len(data))
			pes(sprintf("\r%d%% %d of %d bytes", 100*numBytes/totalSize, numBytes, totalSize))
		}

		count++
		return nil
	}
}

// NewPlainWriter writes chunks without any additional formatting
func NewPlainWriter(file *os.File, totalSize int64, verbose bool) func([]byte) error {
	f := file
	if file == nil {
		f = os.Stdout
	}

	// init values for first block written
	count := uint64(0)
	numBytes := int64(0)
	mux := &sync.Mutex{}

	return func(data []byte) error {
		// this line should be first
		mux.Lock()
		defer mux.Unlock()

		if count > maxChunks {
			panic("Fuck! Tried to write more than 2^32 chunks, which is a problem for encrypted PaF files.\n")
		}

		// compute cipher text
		lenData := len(data)

		// output results by appending to files
		f.Write(data)

		if verbose {
			numBytes += int64(lenData)
			pes(sprintf("\r%d%% %d of %d bytes", 100*numBytes/totalSize, numBytes, totalSize))
		}

		count++
		return nil
	}
}
