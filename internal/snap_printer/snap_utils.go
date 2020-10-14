package snap_printer

import (
	"fmt"
	"github.com/evanw/esbuild/internal/js_ast"
)

func stringifyEString(estring *js_ast.EString) string {
	s := ""
	for _, char := range estring.Value {
		s += fmt.Sprintf("%c", char)
	}
	return s
}

func functionNameForId(id string) string {
	return fmt.Sprintf("__get_%s__()", id)
}
