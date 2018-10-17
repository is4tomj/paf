package pio

import (
	"crypto/sha256"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"sync"
	"os"
	"io"
	"encoding/binary"
)

func createKey(passphrase []byte) [32]byte {
	return sha256.Sum256(passphrase)
}

func encrypt(data []byte, key [32]byte) []byte {
	block, _ := aes.NewCipher(key[:])
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	// appending the result to nonce
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}

func decrypt(data []byte, key [32]byte) []byte {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}

const maxChunks = uint64(1 << 32) - 1
// This will generate a file comprising encrypted chunks:
// length, E(chunk1), length, E(chunk2) ...
// The file is encrypted using GCM 256-bit encryption,
// which is defined in NIST 800-38D. Each chunk is encrypted
// with its own nonce. Accordingly, the number of chunks
// should never be more than 2^32 chunks.
// See [golang docs](https://golang.org/pkg/crypto/cipher/#example_NewGCM_encrypt)
func NewEncryptedPafWriter(file *os.File, passphrase string) func([]byte)error {
	if passphrase == "" {
		panic(errors.New("Fuck! No passphrase provided."))
	}

	if file == nil {
		panic(errors.New("Fuck! Output file was not valid."))
	}
	f := file

	// init values for first block written
	count := uint64(0)
	lenbuf := make([]byte, 4)
	key := createKey([]byte(passphrase))
	mux := &sync.Mutex{}

	return func(data []byte) error {
		// this line should be first
		mux.Lock()

		if count > maxChunks {
			panic("Fuck! Tried to encrypt more than 2^32 chunks.\n")
		}

		// compute cipher text
		ciphertext := encrypt(data, key) // reuse key but get a different nonce
		clen := uint32(len(ciphertext)) // chunks can never be more than 2^32 bytes

		// output results by appending to files
		binary.LittleEndian.PutUint32(lenbuf, clen)
		f.Write(lenbuf) // 4 bytes
		f.Write(ciphertext)

		// these lines should be last and in this order
		count++
		mux.Unlock()

		return nil
	}
}

type encryptedChunk struct {
	entryPoint int64
	size int
}

func ProcessEncryptedPaf(file *os.File, passphrase string, numProcs int, initFunc func(int), processFunc func(int, []byte)) {
	if passphrase == "" {
		panic(errors.New("Fuck! No passphrase provided."))
	}

	fInfo, err := file.Stat()
	if err != nil  {
		panic(errors.New("Fuck! Output file was not valid."))
	}
	f := file
	flen := fInfo.Size()

	// init values for first block written
	lenbuf := make([]byte, 4)
	key := createKey([]byte(passphrase))

	// find out how many chunks are in the file
	pes("Finding chunks in file...\r")
	numChunks := 0
	for i := int64(0); i < flen; {
		n, err := f.ReadAt(lenbuf, i)
		if n <= 0 {
			panic(errors.New("Fuck! Calculated a chunk size of zero."))
		}
		if err != nil && err != io.EOF {
			panic(err)
		}
		len := binary.LittleEndian.Uint32(lenbuf)
		i += 4
		i += int64(len)
		numChunks++
		if err == io.EOF {
			break
		}
	}

	// Find entry points and sizes
	eChunksChan := make(chan *encryptedChunk, numChunks)
	for i := int64(0); i < flen; {
		n, err := f.ReadAt(lenbuf, i)
		if n <= 0 {
			panic(errors.New("Fuck! Calculated a chunk size of zero."))
		}
		if err != nil && err != io.EOF {
			panic(err)
		}

		len := binary.LittleEndian.Uint32(lenbuf)
		eChunksChan <- &encryptedChunk{i+int64(4),int(len)}
		
		i += 4
		i += int64(len)
		
		if err == io.EOF {
			break
		}
	}

	// Spin up goroutines and decrypt
	var wg sync.WaitGroup
	wg.Add(numProcs)
	for i := 0; i < numProcs; i++ {
		go func(pid int) {
			defer wg.Done()
			if initFunc != nil {
				initFunc(pid)
			}

			for eChunk := range eChunksChan {
				numToGo := len(eChunksChan)
				numDone := numChunks - numToGo
				pes(sprintf("\rFinished %d of %d (%d%%)", numDone, numChunks, (numDone*100)/numChunks))
				if processFunc != nil {
					eDataBuff := make([]byte, eChunk.size)
					n, err := f.ReadAt(eDataBuff, eChunk.entryPoint)
					if n <= 0 {
						panic(errors.New("Fuck! Could not read eChunk."))
					}
					if err != nil && err != io.EOF {
						panic(err)
					}
					
					buff := decrypt(eDataBuff, key)
					
					processFunc(pid, buff)
				}
			}
		}(i)
	}

	// Wait
	close(eChunksChan)
	wg.Wait()
	pes(sprintf("\rFinished processing %d encrypted chunks.\t\t\n", numChunks))
}

