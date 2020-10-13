package snap_printer

import (
	"fmt"
	"github.com/evanw/esbuild/internal/js_ast"
)

type RequireDecl struct {
	requireCall    js_ast.Expr
	requireArg     string
	identifier     js_ast.Ref
	identifierName string
}

type NonRequireDecl struct {
	kind js_ast.LocalKind
	decl js_ast.Decl
}

type MaybeRequireDecl struct {
	isRequire  bool
	require    RequireDecl    // use if this is a require
	nonRequire NonRequireDecl // use if this is not a require
}

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

func (p *printer) extractRequireExpression(expr js_ast.Expr) (RequireDecl, bool) {
	switch x := expr.Data.(type) {
	case *js_ast.ECall:
		target := x.Target
		args := x.Args
		// require('foo') has exactly one arg
		if len(args) == 1 {
			switch x := target.Data.(type) {
			case *js_ast.EIdentifier:
				name := p.nameForSymbol(x.Ref)
				if name == "require" {
					arg := args[0]
					var argString string
					switch x := arg.Data.(type) {
					case *js_ast.EString:
						argString = stringifyEString(x)
					}
					if p.shouldReplaceRequire(argString) {
						return RequireDecl{
							requireCall: expr,
							requireArg:  argString,
						}, true
					}
				}
			}
		}
	}
	return RequireDecl{}, false
}

func (p *printer) extractBinding(binding js_ast.Binding) (js_ast.Ref, string, bool) {
	switch b := binding.Data.(type) {
	case *js_ast.BIdentifier:
		return b.Ref, p.nameForSymbol(b.Ref), true
	}
	return js_ast.Ref{}, "", false
}

func (p *printer) extractRequireDeclaration(decl js_ast.Decl) (RequireDecl, bool) {
	if decl.Value != nil {
		// First verify that this is a statement that assigns the result of a
		// `require` call.
		require, isRequire := p.extractRequireExpression(*decl.Value)
		if !isRequire {
			return RequireDecl{}, false
		}
		// Dealing with a require we need to figure out what the result of it is
		// assigned to
		identifier, identifierName, ok := p.extractBinding(decl.Binding)
		// If it is not assigned we cannot handle it at this point
		if ok {
			return RequireDecl{
				requireCall:    require.requireCall,
				requireArg:     require.requireArg,
				identifier:     identifier,
				identifierName: identifierName}, true
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

func (p *printer) printRequireReplacement(require RequireDecl, fnCall string) {
	id := require.identifierName

	idDeclaration := fmt.Sprintf("let %s;", id)
	fnHeader := fmt.Sprintf("function %s {", fnCall)
	fnBodyStart := fmt.Sprintf("  return %s = %s || ", id, id)
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

func (p *printer) handleSLocal(local *js_ast.SLocal) (handled bool) {
	maybeRequires := p.extractRequireDeclarations(local)
	if !hasRequire(maybeRequires) {
		return false
	}

	for _, maybeRequire := range maybeRequires {
		if maybeRequire.isRequire {
			require := maybeRequire.require

			fnCall := fmt.Sprintf("__get_%s__()", require.identifierName)
			p.renamer.Replace(require.identifier, fnCall)
			p.printRequireReplacement(require, fnCall)
		} else {
			p.printNonRequire(maybeRequire.nonRequire)
		}
	}
	return true
}
