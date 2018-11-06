package crypto

import (
	"crypto/rand"
)

// Base62 RNG
const (
	base62Bytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	base62IdxBits = 6                    // 6 bits to represent a letter index
	base62IdxMask = 1<<base62IdxBits - 1 // All 1-bits, as many as base62IdxBits
	base62IdxMax  = 63 / base62IdxBits   // # of letter indices fitting in 63 bits
)
func NewBase62Generator(cacheLength int) func(int)string {
	cache := make([]byte, cacheLength)
	remain := 0
	
	return func(length int) string {
		
	}
}

str(src *mrand.Source, n int) string {
	return randstr(src, n, base62IdxBits, base62IdxMask, base62IdxMax, base62Bytes)
}

func randstr(src *rand.Source, n, idxBits, idxMask, idxMax int, dic string) string {
	b := make([]byte, n)
	lenDic := len(dic)
	// A src.Int63() generates 63 random bits, enough for idxMax characters!
	for i, cache, remain := n-1, (*src).Int63(), idxMax; i >= 0; {
		if remain == 0 {
			cache, remain = (*src).Int63(), idxMax
		}
		if idx := int(cache & idxMask); idx < lenDic {
			b[i] = dic[idx]
			i--
		}
		cache >>= idxBits
		remain--
	}

	return string(b)
}



func sprngStrs(num int, lenth int, encodeFunc func([]byte)string, buff []byte) ([]byte, error) {
	if encodeFunc == nil {
		encodeFunc = hex.EncodeToString
	}
	
	totalLength := num*length	
	if buff == nil || len(buff) < totalLength {
		buff = make([]byte, totalLength)
	}

	
	n, err := rand.Read(buff)
	if err != nil {
		return nil, err
	} else if n < totalLength {
		return nil, errors.New("Fuck! Not enough random bytes were read.")
	}
	
	salt := encodeFunc(saltBuff[start:end])
	return buff, nil
}
