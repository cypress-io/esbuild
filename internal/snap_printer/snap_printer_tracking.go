package snap_printer

import "github.com/evanw/esbuild/internal/js_ast"

// Tracks `let` statements that need to be inserted at the top level scope and
// the top of the file.
// This is the simplest way to ensure that the replacement functions are declared
// before they are used and accessible where needed.
// The fact that they're not declared at the exact same scope as the original identifier
// should not matter esp. since their names are unique and thus won't be shadowed.
// Example:
// ```
// let a
// a = require('a')
// ```
// becomes
// ```
// let __get_a__;
// let a;
// __get_a__ = function() {
// 	 return a = a || require("a")
// };
// ```

func (p *printer) trackTopLevelVar(decl string) {
	p.topLevelVars = append(p.topLevelVars, decl)
}

func prepend(p *printer, s string) {
	data := []byte(s)
	p.js = append(data, p.js...)

}

func (p *printer) prependTopLevelDecls() {
	if len(p.topLevelVars) == 0 {
		return
	}
	decl := "let "
	for i, v := range p.topLevelVars {
		if i > 0 {
			decl += ", "
		}
		decl += v
	}
	// TODO: consider not adding a newline here to avoid affecting source-mapped lines
	decl += ";\n"
	prepend(p, decl)
}

//
// Rewrite globals
//

// globals derived from electron-link blueprint declarations
// See: https://github.com/atom/electron-link/blob/abeb97d8633c06ac6a762ac427b272adebd32c4f/src/blueprint.js#L6
// Also related to: internal/resolver/resolver.go :1246 (BuiltInNodeModules)
var snapGlobals = []string{"process", "document", "global", "window", "console"}

func (p *printer) rewriteGlobals() {
	for outerIdx, outer := range p.symbols.Outer {
		for innerIdx, ref := range outer {
			// Globals aren't declared anywhere and thus are unbound
			if ref.Kind == js_ast.SymbolUnbound {
				for _, global := range snapGlobals {
					if ref.OriginalName == global {
						p.symbols.Outer[outerIdx][innerIdx].OriginalName = functionCallForGlobal(global)
						continue
					}
				}
			}
		}
	}
}
