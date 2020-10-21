package snap_renamer

import (
	"github.com/evanw/esbuild/internal/js_ast"
)

type replacement struct {
	original string
	replaced string
}

type SnapRenamer struct {
	symbols             js_ast.SymbolMap
	deferredIdentifiers map[js_ast.Ref]replacement
}

func NewSnapRenamer(symbols js_ast.SymbolMap) SnapRenamer {
	return SnapRenamer{
		symbols:             symbols,
		deferredIdentifiers: make(map[js_ast.Ref]replacement),
	}
}

func (r *SnapRenamer) resolveRefFromSymbols(ref js_ast.Ref) js_ast.Ref {
	return js_ast.FollowSymbols(r.symbols, ref)
}

func (r *SnapRenamer) NameForSymbol(ref js_ast.Ref) string {
	ref = r.resolveRefFromSymbols(ref)
	deferredIdentifier, ok := r.deferredIdentifiers[ref]
	if ok {
		return deferredIdentifier.replaced
	}
	name := r.symbols.Get(ref).OriginalName
	return name
}

// Stores a replacement string for accesses to the given ref that is used when
// @see NameForSymbol is called later.
// The replacement is a function call, i.e. `__get_a__()` which will be printed
// in place of the original var, i.e. `a`.
func (r *SnapRenamer) Replace(ref js_ast.Ref, replaceWith string) {
	ref = r.resolveRefFromSymbols(ref)
	original := r.NameForSymbol(ref)
	r.deferredIdentifiers[ref] = replacement{
		original: original,
		replaced: replaceWith,
	}
}

// Returns `true` if a replacement was registered for the given ref
func (r *SnapRenamer) HasBeenReplaced(ref js_ast.Ref) bool {
	ref = r.resolveRefFromSymbols(ref)
	_, ok := r.deferredIdentifiers[ref]
	return ok
}

// Returns the original id of the ref whose id has been replaced before.
// This function panics if no replacement is found for this ref.
func (r *SnapRenamer) GetOriginalId(ref js_ast.Ref) string {
	ref = r.resolveRefFromSymbols(ref)
	replacement, ok := r.deferredIdentifiers[ref]
	if !ok {
		panic("Should only ask for original ids for the ones that were replaced")
	}
	return replacement.original
}
