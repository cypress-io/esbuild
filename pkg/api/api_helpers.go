package api

import (
	"github.com/evanw/esbuild/internal/bundler"
	"github.com/evanw/esbuild/internal/js_ast"
	"github.com/evanw/esbuild/internal/js_printer"
	"github.com/evanw/esbuild/internal/renamer"
	"github.com/evanw/esbuild/internal/snap_printer"
	"github.com/evanw/esbuild/internal/snap_renamer"
)

func replaceAll(string) bool { return true }

func createPrintAST(snapshot bool) bundler.PrintAST {
	if snapshot {
		// TODO: we need more snapshot related config here
		return func(
			tree js_ast.AST,
			symbols js_ast.SymbolMap,
			_ renamer.Renamer,
			options js_printer.PrintOptions) js_printer.PrintResult {
			r := snap_renamer.NewSnapRenamer(symbols)
			return snap_printer.Print(tree, symbols, &r, options, replaceAll)
		}
	} else {
		return js_printer.Print
	}
}
