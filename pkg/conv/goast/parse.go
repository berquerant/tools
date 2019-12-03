/*
Package goast provides utilities for golang ast
*/
package goast

import (
	"go/ast"
	"go/parser"
	"go/token"
	"tools/pkg/errors"
)

// ParseFile parses string as file
func ParseFile(src string) (*ast.File, errors.Error) {
	f, err := parser.ParseFile(token.NewFileSet(), "", src, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return nil, errors.NewError().SetCode(errors.Parse).SetError(err)
	}
	return f, nil
}

// ParseExpr parses string as expression
func ParseExpr(src string) (ast.Expr, errors.Error) {
	f, err := parser.ParseExpr(src)
	if err != nil {
		return nil, errors.NewError().SetCode(errors.Parse).SetError(err)
	}
	return f, nil
}
