package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fwp-linter <path to go file>")
		return
	}

	path := os.Args[1]
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Failed to parse file: %v\n", err)
		return
	}

	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			if hasFwPAnnotation(genDecl.Doc) {
				for _, method := range node.Decls {
					funcDecl, ok := method.(*ast.FuncDecl)
					if !ok {
						continue
					}

					if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
						switch expr := funcDecl.Recv.List[0].Type.(type) {
						case *ast.Ident:
							ident := expr
							if ident.Name == typeSpec.Name.Name {
								doc := generateDocumentation(funcDecl, structType)

								// Insert the generated documentation above the function declaration
								fmt.Println(doc)
							}
						case *ast.StarExpr:
							if ident, ok := expr.X.(*ast.Ident); ok && ident.Name == typeSpec.Name.Name {
								doc := generateDocumentation(funcDecl, structType)

								// Insert the generated documentation above the function declaration
								fmt.Println(doc)
							}
						default:
						}
					}
				}
			}
		}
	}

	// Write the original file content with the additional documentation to a new file

	printer.Fprint(os.Stdout, fset, node)
}

func hasFwPAnnotation(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}

	for _, comment := range doc.List {
		if strings.Contains(comment.Text, "@fwp.linter") {
			return true
		}
	}

	return false
}

func generateDocumentation(funcDecl *ast.FuncDecl, structType *ast.StructType) string {
	fieldOperations := make(map[string]map[string]struct{})

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		switch x := n.(type) {

		case *ast.AssignStmt:
			for _, lhs := range x.Lhs {
				if sel, ok := lhs.(*ast.SelectorExpr); ok {
					if _, ok := sel.X.(*ast.Ident); ok {
						if fieldOperations[sel.Sel.Name] == nil {
							fieldOperations[sel.Sel.Name] = make(map[string]struct{}, 2)
						}
						fieldOperations[sel.Sel.Name]["write"] = struct{}{}
					}
				}
			}
			for _, rhs := range x.Rhs {
				if sel, ok := rhs.(*ast.SelectorExpr); ok {
					if _, ok := sel.X.(*ast.Ident); ok {
						if fieldOperations[sel.Sel.Name] == nil {
							fieldOperations[sel.Sel.Name] = make(map[string]struct{}, 2)
						}
						fieldOperations[sel.Sel.Name]["read"] = struct{}{}
					}
				}
			}
		case *ast.BinaryExpr:
			if sel, ok := x.X.(*ast.SelectorExpr); ok {
				if _, ok := sel.X.(*ast.Ident); ok {
					if fieldOperations[sel.Sel.Name] == nil {
						fieldOperations[sel.Sel.Name] = make(map[string]struct{}, 2)
					}
					fieldOperations[sel.Sel.Name]["read"] = struct{}{}
				}
			}
			if sel, ok := x.Y.(*ast.SelectorExpr); ok {
				if _, ok := sel.X.(*ast.Ident); ok {
					if fieldOperations[sel.Sel.Name] == nil {
						fieldOperations[sel.Sel.Name] = make(map[string]struct{}, 2)
					}
					fieldOperations[sel.Sel.Name]["read"] = struct{}{}
				}
			}
		}
		return true
	})

	var doc strings.Builder
	doc.WriteString(fmt.Sprintf("/* START FwP Linter Generated Documentation for method %s\n", funcDecl.Name.Name))
	for field, operations := range fieldOperations {
		switch len(operations) {
		case 0:
			doc.WriteString(fmt.Sprintf("%s have operation %s\n", field, operations))
		case 1:
			var op string
			for operation := range operations {
				op = operation
			}
			doc.WriteString(fmt.Sprintf("%s have operation %s\n", field, op))
		case 2:
			doc.WriteString(fmt.Sprintf("%s have operation both read and write\n", field))
		default:
			panic("not expected more than 2 operations")
		}
	}
	doc.WriteString("END FwP Linter Generated Documentation */\n")

	return doc.String()
}
