PaF: Parallel af or Passwords are Fucked
===========

Toolkit and Golang framework for processing data, particularly credentials, in parallel. Paf comprises libraries, tools, and file format definition for processing data in parallel.

## Tools

### Generating Salts (gen-salts)

Generate salts and print to stdout in PAF format. Each salt is represented as hex string.

[NIST 800-63b](https://pages.nist.gov/800-63-3/sp800-63b.html) requires each salt to be a secure PRNG that has at least 112 bits (14 bytes).

Usage of gen-salts:
  -num-procs int
    	number of processors (default 1)
  -num-salts int
    	number of salts to generate (default 100000000)
  -run-size int
    	number of salts to generate per run (default 10000)
  -salt-len int
    	number of bytes (default 20)


Examples:

```bash
$ time paf gen-salts --num-salts 10000000 --num-procs 4 > salts.paf
```


### Generate Credentials (gen-creds)

Print random credentials to stdout.

Usage of gen-creds:
  -num-creds int
    	number of creds to generate (default 100000000)
  -num-procs int
    	number of processors (default 1)
  -run-size int
    	number of creds to generate per run (default 10000)

Examples:

```bash
$ time paf gen-creds --num-creds=100000000 --run-size=10000 --num-procs=4 > file.paf
```

## PAF File Format

A PAF file file comprises one or more records, which are delimited by newline characters. Each record comprises one or more fields. Fields are delimited by tab characters.

A PAF file may have a corresponding header file with the following format
* The *first line* is the name of the underlying data file.
* The *first 32 bits* are  unsigned integer representing the PAF version number.
* The *next 64 bits* are an unsigned integer indicating how many blocks are in the file.
* The *rest of the header* comprises, a *series of 64-bit unsigned integers* referred to as block-offsets. Each block-offest indicates the beginning of the corresponding block in the file. The block-offsets are monotonically increaseing values.


