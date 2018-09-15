Paf: Parallel af
===========

Golang framework for processing data in parallel. Paf comprises libraries, tools, and file format definition for processing data in parallel.

## Tools

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

A PAF file stores data blocks followed by a header. Each data block stores data in the following format:
* 1-byte unsigned integer for the number of values in the row.
* For each value, 1-byte unsigned integer for the length of the value in bytes.

The header is at the end of the file in the following format:
* The last 64 bits are an unsigned integer indicating an offset into the file that represents the beginning of the header.
* The first 32 bits are  unsigned integer representing the PAF version number.
* The next 64 bits are an unsigned integer indicating how many blocks are in the file.
* The rest of the header comprises, a series of 64-bit unsigned integers referred to as block-offsets. Each block-offest indicates the beginning of the corresponding block in the file. The block-offsets are monotonically increaseing values.


