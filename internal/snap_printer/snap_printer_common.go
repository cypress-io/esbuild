package snap_printer

import (
	"github.com/evanw/esbuild/internal/js_ast"
)

type RequireExpr struct {
	requireCall js_ast.Expr
	requireArg  string
	propChain   []string
}

type RequireDecl struct {
	RequireExpr
	identifier     js_ast.Ref
	identifierName string
}

func (e *RequireExpr) toRequireDecl(identifier js_ast.Ref, identifierName string) RequireDecl {
	return RequireDecl{
		*e,
		identifier,
		identifierName}
}

func (d *RequireDecl) getRequireExpr() *RequireExpr {
	return &RequireExpr{requireCall: d.requireCall, requireArg: d.requireArg, propChain: d.propChain}
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
// Extractors
//

// Extracts the require call expression including information about the argument to the require call.
// NOTE: that this does not include any information about the identifier to which the require call
// result was bound to.
func (p *printer) extractRequireExpression(expr js_ast.Expr, depth int) (*RequireExpr, bool) {
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
						return &RequireExpr{
							requireCall: expr,
							requireArg:  argString,
							propChain:   make([]string, depth),
						}, true
					}
				}
			}
		}

	case *js_ast.EDot:
		// const b = require('x').a.b
		// we see .b then .a then the require (ECall) when we recursively call this function
		require, ok := p.extractRequireExpression(x.Target, depth+1)
		if !ok {
			return require, false
		}
		// add properties in the order they need to be written
		idx := len(require.propChain) - 1 - depth
		require.propChain[idx] = x.Name
		return require, true
	}
	return &RequireExpr{}, false
}

func (p *printer) extractBinding(binding js_ast.Binding) (js_ast.Ref, string, bool) {
	switch b := binding.Data.(type) {
	case *js_ast.BIdentifier:
		return b.Ref, p.nameForSymbol(b.Ref), true
	}
	return js_ast.Ref{}, "", false
}

//
// Printers
//
func (p *printer) printRequireBody(require *RequireExpr) {
	p.printExpr(require.requireCall, js_ast.LLowest, 0)
	for _, prop := range require.propChain {
		p.print(".")
		p.print(prop)
	}
}
