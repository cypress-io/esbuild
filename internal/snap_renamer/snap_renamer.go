package snap_renamer

import "github.com/evanw/esbuild/internal/js_ast"

type SnapRenamer struct {
	symbols             js_ast.SymbolMap
	deferredIdentifiers map[js_ast.Ref]string
}

func NewSnapRenamer(symbols js_ast.SymbolMap) SnapRenamer {
	return SnapRenamer{
		symbols:             symbols,
		deferredIdentifiers: make(map[js_ast.Ref]string),
	}
}

func (r *SnapRenamer) resolveRefFromSymbols(ref js_ast.Ref) js_ast.Ref {
	return js_ast.FollowSymbols(r.symbols, ref)
}

func (r *SnapRenamer) NameForSymbol(ref js_ast.Ref) string {
	ref = r.resolveRefFromSymbols(ref)
	deferredIdentifier, ok := r.deferredIdentifiers[ref]
	if ok {
		return deferredIdentifier
	}
	return r.symbols.Get(ref).OriginalName
}

func (r *SnapRenamer) Replace(ref js_ast.Ref, replacement string) {
	ref = r.resolveRefFromSymbols(ref)
	r.deferredIdentifiers[ref] = replacement
}
