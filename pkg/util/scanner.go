package util

import (
	"bufio"
	"io"
)

const (
	ScanBufferSize    = 1024 * 1024      //1M
	ScanBufferMaxSize = 10 * 1024 * 1024 //10M
)

func NewScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, ScanBufferSize), ScanBufferMaxSize)
	return scanner
}
