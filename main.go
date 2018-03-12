package main

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		panic("skullernator: I need the folder path!")
	}

	dirPath := os.Args[1]
	fset := token.NewFileSet()
	f, blobs, err := ParseDir(fset, dirPath, parser.ParseComments)
	if err != nil {
		panic("skullernator: ParseDir" + err.Error())
	}

	docToken := "skulleton:skullernator"

	for _, astPkg := range f {
		docPkg := doc.New(astPkg, dirPath, 0)
		println("aaaaaaaaaaaaa", docPkg.Imports[0], docPkg.ImportPath)
		for _, t := range docPkg.Types {
			if strings.Contains(t.Doc, docToken) {
				fmt.Println("type:", t.Name)
				fmt.Println("docs:", t.Doc)
				for _, m := range t.Methods {
					if m.Decl.Type.Params.NumFields() != 2 || m.Decl.Type.Results.NumFields() != 2 {
						continue
					}

					if !isContext(m.Decl.Type.Params.List[0].Type) {
						continue
					}

					i, s, ok := isStarExpr(m.Decl.Type.Params.List[1].Type)
					if !ok {
						continue
					}
					fmt.Println(i, s)

					i, s, ok = isStarExpr(m.Decl.Type.Results.List[0].Type)
					if !ok {
						continue
					}
					fmt.Println(i, s)

					astFile := fset.File(m.Decl.Pos())
					fmt.Println(string(blobs[astFile.Name()][int(m.Decl.Name.Pos())-astFile.Base() : int(m.Decl.End())-astFile.Base()]))
				}
			}
		}
	}

}

func isContext(t ast.Expr) bool {
	selectorExpr, isSelectorExpr := t.(*ast.SelectorExpr)
	if !isSelectorExpr {
		return false
	}

	if selectorExpr.Sel == nil {
		return false
	}

	if selectorExpr.Sel.Name != "Context" {
		return false
	}

	return true
}

func isStarExpr(expr ast.Expr) (string, string, bool) {
	starExpr, isStarExpr := expr.(*ast.StarExpr)
	if !isStarExpr {
		return "", "", false
	}

	switch t := starExpr.X.(type) {
	case *ast.Ident:
		return "", t.Name, true
	case *ast.SelectorExpr:
		ident, isIdent := t.X.(*ast.Ident)
		if !isIdent {
			return "", "", false
		}

		return ident.Name, t.Sel.Name, true
	}

	return "", "", false
}

func ParseDir(fset *token.FileSet, path string, mode parser.Mode) (pkgs map[string]*ast.Package, fileBlobs map[string][]byte, first error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer fd.Close()

	list, err := fd.Readdir(-1)
	if err != nil {
		return nil, nil, err
	}

	pkgs = make(map[string]*ast.Package)
	fileBlobs = make(map[string][]byte)
	for _, d := range list {
		if strings.HasSuffix(d.Name(), ".go") {
			filename := filepath.Join(path, d.Name())
			blob, err := ioutil.ReadFile(filename)
			if err != nil {
				return nil, nil, err
			}

			if src, err := parser.ParseFile(fset, filename, blob, mode); err == nil {
				name := src.Name.Name
				pkg, found := pkgs[name]
				if !found {
					pkg = &ast.Package{
						Name:  name,
						Files: make(map[string]*ast.File),
					}
					pkgs[name] = pkg
				}
				pkg.Files[filename] = src
				fileBlobs[filename] = blob
			} else if first == nil {
				first = err
			}
		}
	}

	return
}
