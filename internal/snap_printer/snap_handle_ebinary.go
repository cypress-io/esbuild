package snap_printer

import (
	"fmt"
	"github.com/evanw/esbuild/internal/js_ast"
)

func (p *printer) extractIdentifier(expr *js_ast.Expr) (js_ast.Ref, string, bool) {
	switch eid := expr.Data.(type) {
	case *js_ast.EIdentifier:
		return eid.Ref, p.nameForSymbol(eid.Ref), true
	}

	return js_ast.Ref{}, "", false
}

func (p *printer) printRequireReplacementFunctionAssign(require RequireExpr, bindingId string, fnName string) {
	fnHeader := fmt.Sprintf("%s = function() {", fnName)
	fnBodyStart := fmt.Sprintf("  return %s = %s || ", bindingId, bindingId)
	fnClose := "}"

	p.printNewline()
	p.print(fnHeader)
	p.printNewline()
	p.print(fnBodyStart)
	p.printExpr(require.requireCall, js_ast.LLowest, 0)
	p.printNewline()
	p.print(fnClose)
	p.printNewline()
}



// similar to slocal but assigning to an already declared variable
// x = require('x')
func (p *printer) handleEBinary(e *js_ast.EBinary) (handled bool) {
	if e.Op != js_ast.BinOpAssign {
		return false
	}

	require, isRequire := p.extractRequireExpression(e.Right)
	if !isRequire { return false }

	idRef, bindingId, isId := p.extractIdentifier(&e.Left)
	if !isId { return false }

	// TODO: handle destructured assignment

	fnName := functionNameForId(bindingId)
	// Splicing on same line as declaration end to hopefully prevent messing source maps up too much
	// TODO: verify how much this affects sourcemaps and if we need valid ones add code to fix them after
	p.spliceAfterDeclEnd(idRef, fmt.Sprintf("let %s;", fnName))
	p.printRequireReplacementFunctionAssign(require, bindingId, fnName)

	p.renamer.Replace(idRef, fnName)

	return true
}
