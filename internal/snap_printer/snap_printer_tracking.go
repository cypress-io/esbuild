package snap_printer

import "github.com/evanw/esbuild/internal/js_ast"

// Keeping track of all printed declarations is necessary in order to insert a declaration
// for the replacement function at the same scope.
func (p *printer) trackRefDeclEnd(ref js_ast.Ref, declEnd int) {
	p.declRefLocs[ref] = declEnd
}

func (p *printer) spliceAfter(loc int, content string) {
	js := append(p.js[:loc + 1], content...)
	p.js = append(js, p.js[loc + 1:]...)
}

func (p *printer) spliceAfterDeclEnd(ref js_ast.Ref, content string) {
	declEnd, ok := p.declRefLocs[ref]
	if !ok {
		panic("Unable to find declaration end for ref")
	}
	p.spliceAfter(declEnd, content)
}
