//////////////////////////////////
// paf gen-creds --num-creds=10
//////////////////////////////////

package main

import (
	"os"
	"flag"
	"time"
	mrand "math/rand"
	"math"
	"sync"
	"bytes"
	"strconv"
)

//////////////////////////////////
// This random string generator is from [this post](https://stackoverflow.com/a/31832326)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func randstr(src *mrand.Source, n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, (*src).Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = (*src).Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
//////////////////////////////////


//
// Modified rand function for valid special password characters according to OWASP.
// https://www.owasp.org/index.php/Password_special_characters
//

const passwordBytes =	"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~" // 95 characters
const (
	passwordIdxBits = 7                    // 6 bits to represent a password index
	passwordIdxMask = 1<<passwordIdxBits - 1 // All 1-bits, as many as passwordIdxBits
	passwordIdxMax  = 95 / passwordIdxBits   // # of password indices fitting in 63 bits
)

func randPassword(src *mrand.Source, n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for passwordIdxMax characters!
	for i, cache, remain := n-1, (*src).Int63(), passwordIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = (*src).Int63(), passwordIdxMax
		}
		if idx := int(cache & passwordIdxMask); idx < len(passwordBytes) {
			b[i] = passwordBytes[idx]
			i--
		}
		cache >>= passwordIdxBits
		remain--
	}

	return string(b)
}

//////////////////////////////////
// End of string generator code
//////////////////////////////////




func randInts(src *mrand.Source, num, min , max int) []int {
	dist := max - min + 1
	numBits := uint(math.Ceil(math.Log2(float64(dist))))
	numBitsMask := int64((1 << numBits) - 1)
	maxValues := 63 / numBits // # of values that can generated with 63 bits

	res := make([]int, num)
	for i, cache, remain := 0, (*src).Int63(), maxValues; i < num; {
		if remain == 0 {
			cache, remain = (*src).Int63(), maxValues
		}

		if val := int(cache & numBitsMask); val < dist {
			res[i] = min + val
			i++
		}
		cache >>= numBits
		remain--
	}

	return res
}

func randValsFromDic(src *mrand.Source, num int, dic []string) []string {
	dicLen := len(dic)
	numBits := uint(math.Ceil(math.Log2(float64(dicLen))))
	numBitsMask := int64((1 << numBits) - 1)
	maxValues := 63 / numBits // # of dic entries that can be generated with 64 bits
	
	res := make([]string, num)
	for i, cache, remain := 0, (*src).Int63(), maxValues; i<num; {
		if remain == 0 {
			cache, remain = (*src).Int63(), maxValues
		}

		if idx := int(cache & numBitsMask); idx < dicLen {
			res[i] = dic[idx]
			i++
		}
		cache >>= numBits
		remain--
	}
	
	return res
}

func randValsFromDicStr(src *mrand.Source, num int, dic string) []string {
	lenDic := len(dic)
	dicArr := make([]string, lenDic)
	for i:=0; i<lenDic; i++ {
		dicArr[i] = string(dic[i])
	}
	return randValsFromDic(src, num, dicArr)
}


func randBools(src *mrand.Source, num int) []bool {
	res := make([]bool, num)
	for i, cache, remain := 0, (*src).Int63(), 63; i<num; i++ {
		if remain == 0 {
			cache, remain = (*src).Int63(), 63
		}

		res[i] = cache & 1 == 1
		cache >>= 1
		remain--
	}
	
	return res
}

func prettyNum(num int) string {
	numStr := strconv.Itoa(num)
	strLen := len(numStr)
	resStr := ""
	for j, i := strLen-1, 0; i < strLen; i++ {
		if i%3 == 2 && j > 0 {
			resStr = "," + numStr[j:j+1] + resStr
		} else {
			resStr = numStr[j:j+1] + resStr
		}
		j--
	}
	return resStr
}

func genCreds() {
	genCredsFlags := flag.NewFlagSet("gen-creds", flag.ContinueOnError)
	numCreds := genCredsFlags.Int("num-creds", 100000000, "number of creds to generate")
	runSize := genCredsFlags.Int("run-size", 10000, "number of creds to generate per run")
	numProcs := genCredsFlags.Int("num-procs", 1 , "number of processors")

	if err := genCredsFlags.Parse(os.Args[2:]); err != nil {
		pes(`
Print random credentials to stdout.
Examples:
    $ paf gen-creds --num-creds 10000000 --num-procs 4 > test.paf
`)
		os.Exit(1)
	} else if *numCreds <= 0 || *numProcs <= 0 {
		pes("FUCK! An error occured OR the number of creds and procs must be greater than zero.\n")
		os.Exit(1)
	} else if *numCreds < *runSize {
		*runSize = *numCreds
	} else {
		pes(sprintf("Generating %s creds with %d procs.\n", prettyNum(*numCreds), *numProcs))
	}

	// the number of creds to generate is divided into "runs"
	// each run is the number of creds that a processor should generate
	numRuns := (*numCreds) / (*runSize)
	runChan := make(chan int, numRuns+1)
	for i:=0; i<numRuns; i++ {
		runChan <- (*runSize)
	}
	leftOverRuns := (*numCreds) % (*runSize)
	runChan <- leftOverRuns	
	
	var wg sync.WaitGroup
	wg.Add(*numProcs)
	for j:=0; j<*numProcs; j++ {
		go func(pid int) {
			defer wg.Done()

			// math.rand is not thread safe, so must create
			// separately for each thread.
			src := mrand.NewSource(time.Now().UnixNano())
			var buff bytes.Buffer;
			
			for run := range runChan {
				usernameLengths := randInts(&src, run, 1, 24)
				domainLengths := randInts(&src, run, 3, 8)
				myTLDs := randValsFromDic(&src, run, []string{"", ".com", ".io", ".org", ".gov"})
				hasDomains := randBools(&src, run)
				passwordLengths := randInts(&src, run, 8, 32)

				for i:=0; i<run; i++ {
					username := randstr(&src, usernameLengths[i])
					if hasDomains[i] == true {
						username  += "@" + randstr(&src, domainLengths[i]) + myTLDs[i]
					}
					password := randPassword(&src, passwordLengths[i])
					
					// for testing 
					buff.Write([]byte(username+"\t"+password+"\n"))
				}
				
				po(buff.Bytes())
				buff.Reset()

				pes(sprintf("\r%d%% finished", ((numRuns - len(runChan))*100) / numRuns))
			}
		}(j)
	}
	
	close(runChan) // this must come before wg.Wait
	wg.Wait()
	pes("\rComplete.                                  \n")

	
}



