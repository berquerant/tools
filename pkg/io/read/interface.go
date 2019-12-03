/*
Package read provides utilities for reading
*/
package read

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"tools/pkg/errors"
)

var (
	EOI = errors.NewError().SetCode(errors.Normal).SetError(fmt.Errorf("end of iterator"))
)

type (
	Iterator interface {
		// Next yields next one.
		// return EOI if end of iterator
		Next() ([]byte, errors.Error)
	}
)

type (
	scannerIterator struct {
		sc *bufio.Scanner
	}
)

// NewScannerIterator makes an iterator that yields line
func NewScannerIterator(r io.Reader) Iterator {
	return &scannerIterator{
		sc: bufio.NewScanner(r),
	}
}

func (s *scannerIterator) Next() ([]byte, errors.Error) {
	if s.sc.Scan() {
		return s.sc.Bytes(), nil
	}
	if err := s.sc.Err(); err != nil {
		return nil, errors.NewError().SetCode(errors.System).SetError(err)
	}
	return nil, EOI
}

// Read reads all as string
func Read(r io.Reader) (string, errors.Error) {
	var (
		iter = NewScannerIterator(r)
		buf  = []string{}
	)
	for {
		b, err := iter.Next()
		if err == nil {
			buf = append(buf, string(b))
			continue
		}
		if err == EOI {
			break
		}
		return "", errors.NewError().SetCode(errors.System).SetError(err)
	}
	return strings.Join(buf, "\n"), nil
}
