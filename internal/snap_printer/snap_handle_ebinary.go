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


// similar to slocal but assigning to an already declared variable
// x = require('x')
func (p *printer) handleEBinary(e *js_ast.EBinary) (handled bool) {
	if e.Op != js_ast.BinOpAssign {
		return false
	}

	require, isRequire := p.extractRequireExpression(e.Right)
	if !isRequire { return false }

	_, id, isId := p.extractIdentifier(&e.Left)
	if !isId { return false }

	// TODO: handle destructured assignment

	fnCall := functionNameForId(id)
	fnHeader := fmt.Sprintf("function %s {", fnCall)
	fnBodyStart := fmt.Sprintf("  return %s = %s || ", id, id)
	fnClose := "}"

	p.printNewline()
	p.print(fnHeader)
	p.printNewline()
	p.print(fnBodyStart)
	p.printExpr(require.requireCall, js_ast.LLowest, 0)
	p.printNewline()
	p.print(fnClose)
	p.printNewline()

	return true
}
