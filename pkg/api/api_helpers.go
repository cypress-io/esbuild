package api

import (
	"github.com/evanw/esbuild/internal/bundler"
	"github.com/evanw/esbuild/internal/js_ast"
	"github.com/evanw/esbuild/internal/js_printer"
	"github.com/evanw/esbuild/internal/renamer"
	"github.com/evanw/esbuild/internal/snap_printer"
	"github.com/evanw/esbuild/internal/snap_renamer"
)

func replaceNone(string) bool { return false }

func createPrintAST(snapshot *SnapshotOptions) bundler.PrintAST {
	if snapshot.CreateSnapshot {
		shouldReplaceRequire := snapshot.ShouldReplaceRequire
		if shouldReplaceRequire == nil {
			shouldReplaceRequire = replaceNone
		}

		return func(
			tree js_ast.AST,
			symbols js_ast.SymbolMap,
			jsRenamer renamer.Renamer,
			options js_printer.PrintOptions) js_printer.PrintResult {
			r := snap_renamer.WrapRenamer(&jsRenamer, symbols)
			return snap_printer.Print(tree, symbols, &r, options, shouldReplaceRequire)
		}
	} else {
		return js_printer.Print
	}
}
