package goast

import (
	"go/ast"
	"go/token"
	"strconv"
)

// GenerateImports generates import stmt
func GenerateImports(importSpecs []string) *ast.GenDecl {
	importDeclSpecs := make([]ast.Spec, len(importSpecs))
	for idx, i := range importSpecs {
		importDeclSpecs[idx] = &ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: strconv.Quote(i),
			},
		}
	}
	return &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: importDeclSpecs,
	}
}
