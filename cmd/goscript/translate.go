package main

import (
	"fmt"
	"go/ast"
	"tools/pkg/conv/goast"
	"tools/pkg/errors"
)

func translateMainPipe(f *ast.File, mainProc string) (*ast.File, errors.Error) {
	if fd := goast.FindFuncDecl(f.Decls, "main", nil, nil, nil); fd != nil {
		return nil, errors.NewError().SetCode(errors.Translate).SetError(fmt.Errorf("func main() found at %v", fd.Pos()))
	}
	if goast.FindFuncDecl(f.Decls, mainProc, nil, []string{"string"}, []string{"string", "error"}) == nil {
		return nil, errors.NewError().SetCode(errors.Translate).SetError(fmt.Errorf("func %s(string) (string, error) not found", mainProc))
	}
	f.Decls = append(f.Decls, generateMainPipe(mainProc))
	return f, nil
}

func translateFilePipeFromFuncLit(f ast.Expr, importSpecs []string) (*ast.File, errors.Error) {
	if !goast.ValidateFuncLit(f, []string{"string"}, []string{"string", "error"}) {
		return nil, errors.NewError().SetCode(errors.Translate).SetError(fmt.Errorf("no function literal"))
	}
	r := generateFilePipeFromFuncLit(f.(*ast.FuncLit))
	if len(importSpecs) == 0 {
		return r, nil
	}
	r.Decls = append([]ast.Decl{goast.GenerateImports(importSpecs)}, r.Decls...)
	return r, nil
}
