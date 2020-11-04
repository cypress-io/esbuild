package snap_printer

import (
	"github.com/evanw/esbuild/internal/ast"
	"github.com/evanw/esbuild/internal/js_ast"
)

type RequireExpr struct {
	requireCall js_ast.Expr
	requireArg  string
	propChain   []string
	callChain   [][]js_ast.Expr
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
	return &RequireExpr{
		requireCall: d.requireCall,
		requireArg:  d.requireArg,
		propChain:   d.propChain,
		callChain:   d.callChain,
	}
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
func (p *printer) extractRequireExpression(expr js_ast.Expr, propDepth int, callDepth int) (*RequireExpr, bool) {
	switch data := expr.Data.(type) {
	case *js_ast.ERequire:
		// @see snap_printer.go `printRequireOrImportExpr`
		record := &p.importRecords[data.ImportRecordIndex]
		// Make sure this is a require we want to handle, for now `import` statements are not
		if record.Kind == ast.ImportDynamic {
			break
		}
		return &RequireExpr{
			requireCall: expr,
			requireArg:  record.Path.Text,
			propChain:   make([]string, propDepth),
			callChain:   make([][]js_ast.Expr, callDepth),
		}, true

	case *js_ast.ECall:
		target := data.Target
		args := data.Args
		switch targetData := target.Data.(type) {
		case *js_ast.EIdentifier:
			name := p.nameForSymbol(targetData.Ref)
			// require('foo') has exactly one arg
			if name == "require" && len(args) == 1 {
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
						propChain:   make([]string, propDepth),
						callChain:   make([][]js_ast.Expr, callDepth),
					}, true
				}
			}
		// require('debug')('express:view')
		case *js_ast.ERequire, *js_ast.ECall:
			require, ok := p.extractRequireExpression(target, propDepth, callDepth+1)
			if !ok {
				return require, false
			}
			// add calls in the order they need to be written
			idx := len(require.callChain) - 1 - callDepth
			require.callChain[idx] = data.Args
			return require, true
		}

	case *js_ast.EDot:
		// const b = require('data').a.b
		// we see .b then .a then the require (ECall) when we recursively call this function
		require, ok := p.extractRequireExpression(data.Target, propDepth+1, callDepth)
		if !ok {
			return require, false
		}
		// add properties in the order they need to be written
		idx := len(require.propChain) - 1 - propDepth
		require.propChain[idx] = data.Name
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

func (p *printer) expressionHasRequireOrGlobalReference(expr *js_ast.Expr) bool {
	if expr == nil {
		return false
	}

	switch x := expr.Data.(type) {
	case *js_ast.EIdentifier:
		if p.renamer.HasBeenReplaced(x.Ref) {
			return true
		}
		// Globals except 'require'
		if p.renamer.IsUnboundNonRequire(x.Ref) {
			return true
		}
		return false
	case *js_ast.ECall:
		for _, arg := range x.Args {
			if p.expressionHasRequireOrGlobalReference(&arg) {
				return true
			}
		}
		return p.expressionHasRequireOrGlobalReference(&x.Target)
	case *js_ast.EDot:
		return p.expressionHasRequireOrGlobalReference(&x.Target)
	case *js_ast.EBinary:
		return p.expressionHasRequireOrGlobalReference(&x.Left) || p.expressionHasRequireOrGlobalReference(&x.Right)
	}

	return false
}

//
// Predicates
//
func (p *printer) haveUnwrappableIdentifier(bindings []RequireBinding) bool {
	for _, b := range bindings {
		if p.renamer.IsUnwrappable(b.identifier) {
			return true
		}
	}
	return false
}

func isDirectFunctionInvocation(e *js_ast.ECall) bool {
	if e == nil || e.Target.Data == nil {
		return false
	}
	switch dot := e.Target.Data.(type) {
	// Invocations via .call and .apply
	case *js_ast.EDot:
		if dot.Target.Data == nil {
			return false
		}
		switch dot.Target.Data.(type) {
		case *js_ast.EFunction:
			if dot.Name == "call" || dot.Name == "apply" {
				return true
			}
		}
	// Direct invocations, i.e. (function () {})()
	case *js_ast.EFunction:
		return true
	}

	return false
}

//
// Printers
//
func (p *printer) printRequireBody(require *RequireExpr) {
	p.printExpr(require.requireCall, js_ast.LLowest, 0)
	for _, args := range require.callChain {
		p.print("(")
		for _, arg := range args {
			p.printExpr(arg, js_ast.LLowest, 0)
		}
		p.print(")")
	}
	for _, prop := range require.propChain {
		p.print(".")
		p.print(prop)
	}
}
