package snap_printer

import "github.com/evanw/esbuild/internal/js_ast"


// TODO: this is a somewhat hacky approach as we're mucking with parts of the JS string that were already printed.
// However we don't know what declaration we need to insert where until we see the assigned require statement.

// In this example

// let a
// function foo() {
// 	 a = require('a')
// }

// We don't know that we need to insert the `let __get_a__` declaration in the same scope as `let a` until we
// encounter the `require` inside `foo`. But at this point the first 2 lines have already been printed.

// The only less hacky approach would be to do two passes on over the AST and note extra print statements that need
// to be done at second pass, i.e. we could say _print x after the declaration including the ref for a was complete_.
// One thing to consider there is if the refs will be the same on second pass.


// Keeping track of all printed declarations is necessary in order to insert a declaration
// for the replacement function at the same scope.

type refDecl struct {
	declEnd              int
	replacementDeclAdded bool
}

func (p *printer) trackRefDeclEnd(ref js_ast.Ref, declEnd int) {
	p.refDecls[ref] = &refDecl{declEnd, false}
}

func (p *printer) offsetRefs(outerIndex uint32, startingAt int, offset int) {
	for ref, refDecl := range p.refDecls {
		if ref.OuterIndex != outerIndex || refDecl.declEnd < startingAt  { continue	 }
		refDecl.declEnd += offset
	}
}

func (p *printer) spliceAfter(loc int, content []byte) {
	// TODO: this slicing approach is broken, we need to actually copy the original slice and construct a new one
	// 	this could also be expensive and is another argument to use a less hacky approach
	// https://golang.org/ref/spec#Appending_and_copying_slices
	js := append(p.js[:loc + 1], content...)
	p.js = append(js, p.js[loc + 1:]...)
}

func (p *printer) spliceAfterDeclEnd(ref js_ast.Ref, content string) {
	refDecl, ok := p.refDecls[ref]
	if !ok {
		panic("Unable to find declaration end for ref")
	}
	// Ensure we add a declaration for the first time the identifier is assigned only
	if refDecl.replacementDeclAdded {
		return
	}

	bytes := []byte(content)
	p.spliceAfter(refDecl.declEnd, bytes)
	p.offsetRefs(ref.OuterIndex, refDecl.declEnd, len(bytes))
	refDecl.replacementDeclAdded = true
}
