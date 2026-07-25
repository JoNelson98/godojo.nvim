package grader

import (
	"go/ast"
	"go/parser"
	"go/token"
)

// VerifyASTRules parses the Go source code and executes each specified AST verification check rule.
func VerifyASTRules(src string, rules []string) ([]ASTCheck, error) {
	fs := token.NewFileSet()
	fileNode, err := parser.ParseFile(fs, "submission.go", src, 0)
	if err != nil {
		return nil, err
	}

	var checks []ASTCheck

	for _, rule := range rules {
		passed := false

		switch rule {
		case "uses_json_new_decoder":
			passed = checkUsesJSONNewDecoder(fileNode)
		case "decoder_reads_request_body":
			passed = checkDecoderReadsRequestBody(fileNode)
		case "decode_receives_address":
			passed = checkDecodeReceivesAddress(fileNode)
		case "uses_json_new_encoder":
			passed = checkUsesJSONNewEncoder(fileNode)
		case "pointer_receiver":
			passed = checkPointerReceiver(fileNode)
		case "uses_httptest_new_request":
			passed = checkUsesHttptestNewRequest(fileNode)
		case "uses_defer_body_close":
			passed = checkUsesDeferBodyClose(fileNode)
		case "declares_handler_tests":
			passed = checkDeclaresHandlerTests(fileNode)
		default:
			// Unknown rules default to passed to avoid blocking learners
			passed = true
		}

		checks = append(checks, ASTCheck{
			Name:   rule,
			Passed: passed,
		})
	}

	return checks, nil
}

func isSelector(expr ast.Expr, xName, selName string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == xName && sel.Sel.Name == selName
}

func checkUsesJSONNewDecoder(file *ast.File) bool {
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if isSelector(call.Fun, "json", "NewDecoder") {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

func checkDecoderReadsRequestBody(file *ast.File) bool {
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if isSelector(call.Fun, "json", "NewDecoder") {
				if len(call.Args) > 0 {
					// Check if argument is a selector expression ending with .Body (like r.Body)
					if sel, ok := call.Args[0].(*ast.SelectorExpr); ok {
						if sel.Sel.Name == "Body" {
							found = true
							return false
						}
					}
				}
			}
		}
		return true
	})
	return found
}

func checkDecodeReceivesAddress(file *ast.File) bool {
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "Decode" {
					if len(call.Args) > 0 {
						// Arg should be unary address operator (&input)
						if op, ok := call.Args[0].(*ast.UnaryExpr); ok {
							if op.Op == token.AND { // &
								found = true
								return false
							}
						}
					}
				}
			}
		}
		return true
	})
	return found
}

func checkUsesJSONNewEncoder(file *ast.File) bool {
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if isSelector(call.Fun, "json", "NewEncoder") {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

func checkPointerReceiver(file *ast.File) bool {
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok && fd.Recv != nil {
			if len(fd.Recv.List) > 0 {
				// Receivers list (usually just one item)
				recvType := fd.Recv.List[0].Type
				if _, ok := recvType.(*ast.StarExpr); ok {
					found = true
					return false
				}
			}
		}
		return true
	})
	return found
}

func checkUsesHttptestNewRequest(file *ast.File) bool {
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if isSelector(call.Fun, "httptest", "NewRequest") {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

func checkUsesDeferBodyClose(file *ast.File) bool {
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		if def, ok := n.(*ast.DeferStmt); ok {
			call := def.Call
			if call != nil {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if sel.Sel.Name == "Close" {
						found = true
						return false
					}
				}
			}
		}
		return true
	})
	return found
}

func checkDeclaresHandlerTests(file *ast.File) bool {
	hasValid := false
	hasInvalid := false
	ast.Inspect(file, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok {
			if fd.Name.Name == "TestHandleCreateCompany_Valid" {
				hasValid = true
			}
			if fd.Name.Name == "TestHandleCreateCompany_Invalid" {
				hasInvalid = true
			}
		}
		return true
	})
	return hasValid && hasInvalid
}
