package main

import (
	"go/ast"
	"go/token"
	"strconv"
)

// generateMainPipe generates main function from specified function
//
// signature:
// func mainProc(string) (string, error)
// translation:
// func main() {
//   sc := bufio.NewScanner(os.Stdin)
//   for sc.Scan() {
//     line := sc.Text()
//     s, err := Main(line)
//     if err != nil {
//       fmt.Fprintf(os.Stderr, "Input:%s Error:%s\n", line, err)
//     } else {
//       fmt.Println(s)
//     }
//   }
//   if err := sc.Err() ; err != nil {
//     panic(err)
//   }
// }
func generateMainPipe(mainProc string) *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: ast.NewIdent("main"),
		Type: &ast.FuncType{},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						ast.NewIdent("sc"),
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("bufio"),
								Sel: ast.NewIdent("NewScanner"),
							},
							Args: []ast.Expr{
								&ast.SelectorExpr{
									X:   ast.NewIdent("os"),
									Sel: ast.NewIdent("Stdin"),
								},
							},
						},
					},
				},
				&ast.ForStmt{
					Cond: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("sc"),
							Sel: ast.NewIdent("Scan"),
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{
									ast.NewIdent("line"),
								},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{
									&ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X:   ast.NewIdent("sc"),
											Sel: ast.NewIdent("Text"),
										},
									},
								},
							},
							&ast.AssignStmt{
								Lhs: []ast.Expr{
									ast.NewIdent("s"),
									ast.NewIdent("err"),
								},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{
									&ast.CallExpr{
										Fun: ast.NewIdent(mainProc),
										Args: []ast.Expr{
											ast.NewIdent("line"),
										},
									},
								},
							},
							&ast.IfStmt{
								Cond: &ast.BinaryExpr{
									X:  ast.NewIdent("err"),
									Op: token.NEQ,
									Y:  ast.NewIdent("nil"),
								},
								Body: &ast.BlockStmt{
									List: []ast.Stmt{
										&ast.ExprStmt{
											X: &ast.CallExpr{
												Fun: &ast.SelectorExpr{
													X:   ast.NewIdent("fmt"),
													Sel: ast.NewIdent("Fprintf"),
												},
												Args: []ast.Expr{
													&ast.SelectorExpr{
														X:   ast.NewIdent("os"),
														Sel: ast.NewIdent("Stderr"),
													},
													&ast.BasicLit{
														Kind:  token.STRING,
														Value: strconv.Quote("Input:%s Error:%s\n"),
													},
													ast.NewIdent("line"),
													ast.NewIdent("err"),
												},
											},
										},
									},
								},
								Else: &ast.BlockStmt{
									List: []ast.Stmt{
										&ast.ExprStmt{
											X: &ast.CallExpr{
												Fun: &ast.SelectorExpr{
													X:   ast.NewIdent("fmt"),
													Sel: ast.NewIdent("Println"),
												},
												Args: []ast.Expr{
													ast.NewIdent("s"),
												},
											},
										},
									},
								},
							},
						},
					},
				},
				&ast.IfStmt{
					Init: &ast.AssignStmt{
						Lhs: []ast.Expr{
							ast.NewIdent("err"),
						},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   ast.NewIdent("sc"),
									Sel: ast.NewIdent("Err"),
								},
							},
						},
					},
					Cond: &ast.BinaryExpr{
						X:  ast.NewIdent("err"),
						Op: token.NEQ,
						Y:  ast.NewIdent("nil"),
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ExprStmt{
								X: &ast.CallExpr{
									Fun: ast.NewIdent("panic"),
									Args: []ast.Expr{
										ast.NewIdent("err"),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// generateFilePipeFromFuncLit make it executable.
//
// signature:
// func(string) (string, error)
// translation:
// package main
//
// var Main = FUNC_LITERAL
//
// func main() {
//   sc := bufio.NewScanner(os.Stdin)
//   for sc.Scan() {
//     line := sc.Text()
//     s, err := Main(line)
//     if err != nil {
//       fmt.Fprintf(os.Stderr, "Input:%s Error:%s\n", line, err)
//     } else {
//       fmt.Println(s)
//     }
//   }
//   if err := sc.Err() ; err != nil {
//     panic(err)
//   }
// }
func generateFilePipeFromFuncLit(f *ast.FuncLit) *ast.File {
	return &ast.File{
		Name: ast.NewIdent("main"),
		Decls: []ast.Decl{
			&ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{
							ast.NewIdent("Main"),
						},
						Values: []ast.Expr{
							f,
						},
					},
				},
			},
			generateMainPipe("Main"),
		},
	}
}
