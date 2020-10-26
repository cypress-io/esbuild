package snap_printer

import (
	"github.com/evanw/esbuild/internal/js_ast"
)

type RequireExpr struct {
	requireCall js_ast.Expr
	requireArg  string
	propChain   []string
}

type RequireReference struct {
	assignedValue *js_ast.Expr
	bindings      []RequireBinding
}

type RequireBinding struct {
	identifier        js_ast.Ref
	identifierName    string
	fnCallReplacement string
	isDestructuring   bool
}

type RequireDecl struct {
	RequireExpr
	bindings []RequireBinding
}

func (e *RequireExpr) toRequireDecl(bindings []RequireBinding) RequireDecl {
	return RequireDecl{*e, bindings}
}

func (d *RequireDecl) getRequireExpr() *RequireExpr {
	return &RequireExpr{requireCall: d.requireCall, requireArg: d.requireArg, propChain: d.propChain}
}

type OriginalDecl struct {
	kind js_ast.LocalKind
	decl js_ast.Decl
}

type MaybeRequireDecl struct {
	isRequire          bool
	require            RequireDecl // use if this is a require
	isRequireReference bool
	requireReference   RequireReference // use if this is a reference to a required var
	originalDecl       OriginalDecl     // use if this is not a require nor a reference
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

func (p *printer) extractBinding(b js_ast.B, isDestructuring bool) RequireBinding {
	switch b := b.(type) {
	case *js_ast.BIdentifier:
		identierName := p.nameForSymbol(b.Ref)
		return RequireBinding{
			identifier:        b.Ref,
			identifierName:    identierName,
			fnCallReplacement: functionCallForId(identierName),
			isDestructuring:   isDestructuring,
		}
	default:
		panic("Expected a BIdentifier")
	}
}

func (p *printer) extractBindings(binding js_ast.Binding) ([]RequireBinding, bool) {
	switch b := binding.Data.(type) {
	case *js_ast.BIdentifier:
		// const a = ...
		binding := p.extractBinding(b, false)
		return []RequireBinding{binding}, true
	case *js_ast.BObject:
		// const { a, b } = ...
		bindings := make([]RequireBinding, len(b.Properties))
		for i, prop := range b.Properties {
			bindings[i] = p.extractBinding(prop.Value.Data, true)
		}
		return bindings, true
	}
	return []RequireBinding{}, false
}

func (p *printer) extractIdentifier(b js_ast.E, isDestructuring bool) RequireBinding {
	// NOTE: this duplication (extractBinding) is necessary since there is no common
	// base for both types of `b`
	switch b := b.(type) {
	case *js_ast.EIdentifier:
		return RequireBinding{
			identifier:      b.Ref,
			identifierName:  p.nameForSymbol(b.Ref),
			isDestructuring: isDestructuring,
		}
	default:
		panic("Expected a EIdentifier")
	}
}

func (p *printer) extractIdentifiers(expr js_ast.E) ([]RequireBinding, bool) {
	switch b := expr.(type) {
	case *js_ast.EIdentifier:
		// a = ...
		binding := p.extractIdentifier(b, false)
		return []RequireBinding{binding}, true
	case *js_ast.EObject:
		// ({ a, b } = ...)
		bindings := make([]RequireBinding, len(b.Properties))
		for i, prop := range b.Properties {
			bindings[i] = p.extractIdentifier(prop.Value.Data, true)
		}
		return bindings, true
	}
	return []RequireBinding{}, false
}

func (p *printer) expressionHasRequireReference(expr *js_ast.Expr) bool {
	if expr == nil {
		return false
	}

	switch x := expr.Data.(type) {
	case *js_ast.EIdentifier:
		return p.renamer.HasBeenReplaced(x.Ref)
	case *js_ast.ECall:
		for _, arg := range x.Args {
			if p.expressionHasRequireReference(&arg) {
				return true
			}
		}
		return false
	case *js_ast.EDot:
		return p.expressionHasRequireReference(&x.Target)
	case *js_ast.EBinary:
		return p.expressionHasRequireReference(&x.Left) || p.expressionHasRequireReference(&x.Right)
	}

	return false
}

//
// Predicates
//
func (p *printer) haveUnboundIdentifier(bindings []RequireBinding) bool {
	for _, b := range bindings {
		if p.renamer.IsUnbound(b.identifier) {
			return true
		}
	}
	return false
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
