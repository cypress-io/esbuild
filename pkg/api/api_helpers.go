package api

import (
	"github.com/evanw/esbuild/internal/bundler"
	"github.com/evanw/esbuild/internal/config"
	"github.com/evanw/esbuild/internal/js_ast"
	"github.com/evanw/esbuild/internal/js_printer"
	"github.com/evanw/esbuild/internal/renamer"
	"github.com/evanw/esbuild/internal/snap_printer"
	"github.com/evanw/esbuild/internal/snap_renamer"
)

func emptyPrintResult() js_printer.PrintResult {
	return js_printer.PrintResult{
		JS:                       make([]byte, 0),
		ExtractedComments:        make(map[string]bool),
		FirstDeclByteOffset:      0,
		FirstDeclSourceMapOffset: 0,
	}

}

func replaceNone(string) bool { return false }
func rewriteAll(string) bool  { return true }

func createPrintAST(snapshot *SnapshotOptions) bundler.PrintAST {
	if snapshot.CreateSnapshot {
		shouldReplaceRequire := snapshot.ShouldReplaceRequire
		if shouldReplaceRequire == nil {
			shouldReplaceRequire = replaceNone
		}
		shouldRewriteModule := snapshot.ShouldRewriteModule
		if shouldRewriteModule == nil {
			shouldRewriteModule = rewriteAll
		}

		return func(
			tree js_ast.AST,
			symbols js_ast.SymbolMap,
			jsRenamer renamer.Renamer,
			options js_printer.Options) js_printer.PrintResult {
			r := snap_renamer.WrapRenamer(&jsRenamer, symbols)
			if options.IsRuntime {
				return js_printer.Print(tree, symbols, &r, options)
			} else {
				if snapshot.ShouldRejectAst != nil {
					// if we can see from the AST that this file cannot be included in a snapshot then we
					// don't parse it, but report the error instead and return early
					err, reject := snapshot.ShouldRejectAst(&tree)
					if reject {
						reportError(options.FilePath, err, snapshot.PanicOnError)
						return emptyPrintResult()
					}
				}
				result := snap_printer.Print(
					tree,
					symbols,
					&r,
					options,
					true,
					shouldReplaceRequire,
					shouldRewriteModule(options.FilePath))
				if snapshot.VerifyPrint {
					verifyPrint(&result, options.FilePath, snapshot.PanicOnError)
				}
				return result
			}
		}
	} else {
		return js_printer.Print
	}
}

func addSnapshotOpts(buildOpts *BuildOptions, configOpts *config.Options) {
	if buildOpts.Snapshot == nil || !buildOpts.Snapshot.CreateSnapshot {
		return
	}
	if buildOpts.Snapshot.AbsBasedir == "" {
		panic("Build configOpts need to have 'Snapshot.AbsBasedir' set when creating a snapshot")
	}
	configOpts.CreateSnapshot = true
	configOpts.SnapshotAbsBaseDir = buildOpts.Snapshot.AbsBasedir
}
