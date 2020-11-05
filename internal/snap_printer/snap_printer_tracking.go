package snap_printer

import (
	"github.com/evanw/esbuild/internal/js_ast"
	"regexp"
)

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

// var require_express2 = __commonJS((exports, module2) => {
var wrapperRx = regexp.MustCompile(`^var require_.+ = __commonJS\([^{]+{(\r\n|\r|\n)`)
func prepend(p *printer, s string) {
	data := []byte(s)
	// We need to ensure that we add our declarations inside the wrapper function when we're dealing
	// with a bundle and the module code is wrapped.
	// Therefore some copying is necessary even though it most likely affects performance.

	idxs := wrapperRx.FindIndex(p.js)
	if idxs == nil {
		p.js = append(data, p.js...)
	} else {
		end := idxs[1]
		jsLen := len(p.js)
		dataLen := len(data)
		completeJs := make([]byte, jsLen + dataLen)
		// Copy the wrapper open code that we matched
		for i := 0; i < end; i++ {
			completeJs[i] = p.js[i]
		}
		// Insert our declaration code
		for i := 0; i < dataLen; i++ {
			completeJs[i + end] = data[i]
		}
		// Copy the module body and wrapper close code after our declarations
		for i := end; i < jsLen; i++ {
			completeJs[i + dataLen] = p.js[i]
		}
		p.js = completeJs
	}
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
						name := functionCallForGlobal(global)
						p.symbols.Outer[outerIdx][innerIdx].OriginalName = name
						continue
					}
				}
			}
		}
	}
}
