package snap_printer

import (
	"fmt"
	"github.com/evanw/esbuild/internal/js_ast"
)

//
// Utils
//
func hasRequire(maybeRequires []MaybeRequireDecl) bool {
	for _, x := range maybeRequires {
		if x.isRequire {
			return true
		}
	}
	return false
}

//
// Extractors
//
func (p *printer) nameForSymbol(ref js_ast.Ref) string {
	return p.renamer.NameForSymbol(ref)
}

func (p *printer) extractRequireDeclaration(decl js_ast.Decl) (RequireDecl, bool) {
	if decl.Value != nil {
		// First verify that this is a statement that assigns the result of a
		// `require` call.
		requireExpr, isRequire := p.extractRequireExpression(*decl.Value)
		if !isRequire {
			return RequireDecl{}, false
		}
		// Dealing with a require we need to figure out what the result of it is
		// assigned to
		identifier, identifierName, ok := p.extractBinding(decl.Binding)
		// If it is not assigned we cannot handle it at this point
		if ok {
			return requireExpr.toRequireDecl(identifier, identifierName), true
		}
	}

	return RequireDecl{}, false
}

func (p *printer) extractRequireDeclarations(local *js_ast.SLocal) []MaybeRequireDecl {
	var maybeRequires []MaybeRequireDecl

	switch local.Kind {
	case js_ast.LocalConst,
		js_ast.LocalLet,
		js_ast.LocalVar:
		if !local.IsExport {
			for _, decl := range local.Decls {
				require, isRequire := p.extractRequireDeclaration(decl)
				if isRequire {
					maybeRequires = append(maybeRequires, MaybeRequireDecl{
						isRequire: true,
						require:   require})
				} else {
					maybeRequires = append(maybeRequires, MaybeRequireDecl{
						isRequire:  false,
						nonRequire: NonRequireDecl{kind: local.Kind, decl: decl},
					})
				}
			}
		}

	}

	return maybeRequires
}

//
// Printers
//
func (p *printer) printNonRequire(nonRequire NonRequireDecl) {
	var keyword string

	switch nonRequire.kind {
	case js_ast.LocalVar:
		keyword = "var"
	case js_ast.LocalLet:
		keyword = "let"
	case js_ast.LocalConst:
		keyword = "const"
	}

	decl := nonRequire.decl

	p.print(keyword)
	p.printSpace()
	p.printBinding(decl.Binding)

	if decl.Value != nil {
		p.printSpace()
		p.print("=")
		p.printSpace()
		p.printExpr(*decl.Value, js_ast.LComma, forbidIn)
	}
}

func (p *printer) printRequireReplacementFunctionDeclaration(require RequireExpr, bindingId string, fnCall string) {
	idDeclaration := fmt.Sprintf("let %s;", bindingId)
	fnHeader := fmt.Sprintf("function %s {", fnCall)
	fnBodyStart := fmt.Sprintf("  return %s = %s || ", bindingId, bindingId)
	fnClose := "}"

	p.printNewline()
	p.print(idDeclaration)
	p.printNewline()
	p.print(fnHeader)
	p.printNewline()
	p.print(fnBodyStart)
	p.printExpr(require.requireCall, js_ast.LLowest, 0)
	p.printNewline()
	p.print(fnClose)
	p.printNewline()
}

// const|let|var x = require('x')
func (p *printer) handleSLocal(local *js_ast.SLocal) (handled bool) {
	maybeRequires := p.extractRequireDeclarations(local)
	if !hasRequire(maybeRequires) {
		return false
	}

	for _, maybeRequire := range maybeRequires {
		if maybeRequire.isRequire {
			require := maybeRequire.require

			id := require.identifierName
			fnCall := functionCallForId(id)
			p.printRequireReplacementFunctionDeclaration(require.getRequireExpr(), id, fnCall)
			p.renamer.Replace(require.identifier, fnCall)
		} else {
			p.printNonRequire(maybeRequire.nonRequire)
		}
	}
	return true
}
