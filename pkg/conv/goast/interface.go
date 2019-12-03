package goast

import (
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"tools/pkg/errors"
	"tools/pkg/io/write"
)

// Print prints node as ast
func Print(dst io.Writer, node interface{}) errors.Error {
	if err := ast.Fprint(dst, token.NewFileSet(), node, ast.NotNilFilter); err != nil {
		return errors.NewError().SetCode(errors.IO).SetError(err)
	}
	return nil
}

// Dump dumps node as go code
func Dump(dst io.Writer, node interface{}) errors.Error {
	if err := format.Node(dst, token.NewFileSet(), node); err != nil {
		return errors.NewError().SetCode(errors.IO).SetError(err)
	}
	return nil
}

func PrintAsString(node interface{}) (string, errors.Error) {
	s, err := write.Write(func(w io.Writer) error {
		return Print(w, node)
	})
	if err != nil {
		return "", errors.NewError().SetCode(errors.IO).SetError(err)
	}
	return s, nil
}

func DumpAsString(node interface{}) (string, errors.Error) {
	s, err := write.Write(func(w io.Writer) error {
		return Dump(w, node)
	})
	if err != nil {
		return "", errors.NewError().SetCode(errors.IO).SetError(err)
	}
	return s, nil
}
