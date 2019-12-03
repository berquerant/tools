/*
Package write provides utilities for writing
*/
package write

import (
	"bytes"
	"io"
	"tools/pkg/errors"
)

// Write writes all as string
func Write(f func(io.Writer) error) (string, errors.Error) {
	buf := &bytes.Buffer{}
	if err := f(buf); err != nil {
		return "", errors.NewError().SetCode(errors.System).SetError(err)
	}
	return buf.String(), nil
}
