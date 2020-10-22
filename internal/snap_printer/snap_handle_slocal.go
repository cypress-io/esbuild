package snap_printer

import (
	"fmt"
	"github.com/evanw/esbuild/internal/js_ast"
)

//
// Utils
//
func hasRequire(maybeRequires *[]MaybeRequireDecl) bool {
	for _, x := range *maybeRequires {
		if x.isRequire {
			return true
		}
	}
	return false
}

func hasRequireReference(maybeRequires *[]MaybeRequireDecl) bool {
	for _, x := range *maybeRequires {
		if x.isRequireReference {
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
		requireExpr, isRequire := p.extractRequireExpression(*decl.Value, 0)
		if !isRequire {
			return RequireDecl{}, false
		}
		// Dealing with a require we need to figure out what the result of it is
		// assigned to
		bindings, ok := p.extractBindings(decl.Binding)
		// If it is not assigned we cannot handle it at this point
		if ok {
			return requireExpr.toRequireDecl(bindings), true
		}
	}

	return RequireDecl{}, false
}
func (p *printer) extractRequireReferenceDeclaration(decl js_ast.Decl) (RequireReference, bool) {
	// TODO: may need the refs of requires we just extracted as part of multiple declarations
	// var hasRequireReference bool
	if decl.Value == nil {
		return RequireReference{}, false
	}

	switch x := decl.Value.Data.(type) {
	case *js_ast.EIdentifier:
		if p.renamer.HasBeenReplaced(x.Ref) {
			bindings, ok := p.extractBindings(decl.Binding)
			if ok {
				return RequireReference{
					assignedValue: decl.Value,
					bindings:      bindings,
				}, true
			}
		}
	}

	return RequireReference{}, false
}

func (p *printer) extractDeclarations(local *js_ast.SLocal) []MaybeRequireDecl {
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
					continue
				}
				reference, hasReference := p.extractRequireReferenceDeclaration(decl)
				if hasReference {
					if reference.assignedValue == nil {
						panic("requireReference should have assigned value set")
					}
					maybeRequires = append(maybeRequires, MaybeRequireDecl{
						isRequireReference: true,
						requireReference:   reference})
					continue
				}
				maybeRequires = append(maybeRequires, MaybeRequireDecl{
					isRequire:  false,
					nonRequire: NonRequireDecl{kind: local.Kind, decl: decl},
				})
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

func (p *printer) printRequireReplacementFunctionDeclaration(
	require *RequireExpr,
	bindingId string,
	isDestructuring bool,
	fnCall string) {

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
	p.printRequireBody(require)
	if isDestructuring {
		// Rewriting `const { a, b } = require()` to `let a; a = require().a`, thus adding `.a` here
		p.print(".")
		p.print(bindingId)
	}
	p.printNewline()
	p.print(fnClose)
	p.printNewline()
}

func (p *printer) printRequireReferenceReplacementFunctionDeclaration(
	reference *RequireReference,
	bindingId string,
	isDestructuring bool,
	fnCall string) {

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
	if isDestructuring {
		// Rewriting `const { a, b } = require()` to `let a; a = require().a`, thus adding `.a` here
		p.print(".")
		p.print(bindingId)
	}
	// TODO: not sure where I'd get a level + flags from in this case
	p.printExpr(*reference.assignedValue, js_ast.LLowest, 0)
	p.printNewline()
	p.print(fnClose)
	p.printNewline()
}

// const|let|var x = require('x')
func (p *printer) handleSLocal(local *js_ast.SLocal) (handled bool) {
	maybeRequires := p.extractDeclarations(local)
	if !hasRequire(&maybeRequires) && !hasRequireReference(&maybeRequires) {
		return false
	}

	for _, maybeRequire := range maybeRequires {
		if maybeRequire.isRequire {
			require := maybeRequire.require
			for _, b := range require.bindings {
				id := b.identifierName
				fnCall := functionCallForId(id)
				p.printRequireReplacementFunctionDeclaration(require.getRequireExpr(), id, b.isDestructuring, fnCall)
				p.renamer.Replace(b.identifier, fnCall)
			}
			continue
		}
		if maybeRequire.isRequireReference {
			reference := &maybeRequire.requireReference
			for _, b := range reference.bindings {
				id := b.identifierName
				fnCall := functionCallForId(id)
				p.printRequireReferenceReplacementFunctionDeclaration(reference, id, b.isDestructuring, fnCall)
				p.renamer.Replace(b.identifier, fnCall)
			}
			continue
		}

		p.printNonRequire(maybeRequire.nonRequire)
	}
	return true
}
