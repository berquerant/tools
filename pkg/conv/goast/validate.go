package goast

import (
	"go/ast"
)

// TypeExprToString converts expression to string
// return empty string if expr is not for type
func TypeExprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident,
		*ast.StarExpr,
		*ast.ArrayType,
		*ast.MapType,
		*ast.InterfaceType,
		*ast.StructType,
		*ast.Ellipsis,
		*ast.SelectorExpr:
		s, err := DumpAsString(t)
		if err != nil {
			return ""
		}
		return s
	}
	return ""
}

func ValidateFieldListByType(fl *ast.FieldList, types []string) bool {
	if fl.NumFields() != len(types) {
		return false
	}
	if fl.NumFields() == 0 {
		return true
	}
	for i, p := range fl.List {
		if TypeExprToString(p.Type) != types[i] {
			return false
		}
	}
	return true
}

func ValidateFuncLit(f ast.Expr, inTypes, outTypes []string) bool {
	fl, ok := f.(*ast.FuncLit)
	if !ok {
		return false
	}
	return ValidateFieldListByType(fl.Type.Params, inTypes) ||
		ValidateFieldListByType(fl.Type.Results, outTypes)
}

func FindFuncDecl(decls []ast.Decl, name string, recvTypes, inTypes, outTypes []string) *ast.FuncDecl {
	for _, d := range decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fd.Name.Name != name ||
			!ValidateFieldListByType(fd.Recv, recvTypes) ||
			!ValidateFieldListByType(fd.Type.Params, inTypes) ||
			!ValidateFieldListByType(fd.Type.Results, outTypes) {
			continue
		}
		return fd
	}
	return nil
}
