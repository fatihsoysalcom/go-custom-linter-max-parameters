package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
)

// maxParams defines our custom linter rule: functions should not have more than this many parameters.
const maxParams = 3

// linter is a custom AST visitor that collects issues.
type linter struct {
	fset   *token.FileSet
	issues []string
}

// Visit is called for each node in the AST.
// It implements the ast.Visitor interface.
func (l *linter) Visit(node ast.Node) ast.Visitor {
	// Check if the node is a function declaration.
	if fn, ok := node.(*ast.FuncDecl); ok {
		// If the function has parameters, count them.
		if fn.Type.Params != nil {
			paramCount := len(fn.Type.Params.List)
			// If the parameter count exceeds our limit, record an issue.
			if paramCount > maxParams {
				pos := l.fset.Position(fn.Pos()) // Get the file position of the function.
				l.issues = append(l.issues, fmt.Sprintf("%s:%d:%d: Function '%s' has %d parameters, exceeding the limit of %d.",
					pos.Filename, pos.Line, pos.Column, fn.Name.Name, paramCount, maxParams))
			}
		}
	}
	// Continue traversing the AST.
	return l
}

func main() {
	// Get the current file's path dynamically to lint itself.
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Println("Could not determine current file path.")
		os.Exit(1)
	}
	filename := filepath.Base(currentFile) // Just the filename, not the full path

	fset := token.NewFileSet() // A FileSet is a collection of files.
	// Parse the current file into an AST.
	// parser.ParseComments ensures comments are included in the AST, though not used in this linter.
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing file %s: %v\n", filename, err)
		os.Exit(1)
	}

	l := &linter{fset: fset} // Initialize our linter visitor.
	ast.Walk(l, node)        // Traverse the AST using our linter.

	// Report any issues found.
	if len(l.issues) > 0 {
		fmt.Println("Linter found issues:")
		for _, issue := range l.issues {
			fmt.Println(issue)
		}
		os.Exit(1) // Indicate that issues were found.
	} else {
		fmt.Printf("No issues found in %s.\n", filename)
	}
}

// --- Functions below this line are for demonstrating the linter's capabilities ---

// This function is designed to pass the linter check (2 parameters).
func exampleFunctionPass(a int, b string) int {
	return a + len(b)
}

// This function is also designed to pass the linter check (3 parameters).
func anotherExamplePass(a int, b string, c float64) bool {
	return a > 0 && len(b) > 0 && c != 0
}

// This function is designed to FAIL the linter check (4 parameters).
// It demonstrates a function with too many parameters according to our rule.
func exampleFunctionFail(a int, b string, c float64, d bool) string {
	if d {
		return fmt.Sprintf("%d %s %.2f", a, b, c)
	}
	return ""
}

// This function is also designed to FAIL the linter check (5 parameters).
func anotherExampleFail(p1, p2, p3, p4, p5 int) int {
	return p1 + p2 + p3 + p4 + p5
}

// This function passes as it has no parameters.
func noParametersFunc() {
	fmt.Println("This function has no parameters.")
}
