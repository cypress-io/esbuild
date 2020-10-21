package snap_printer

import (
	"fmt"
	"github.com/evanw/esbuild/internal/js_ast"
)

func (p *printer) printRequireReplacementFunctionAssign(
	require *RequireExpr,
	bindingId string,
	isDestructuring bool,
	fnName string) {

	fnHeader := fmt.Sprintf("%s = function() {", fnName)
	fnBodyStart := fmt.Sprintf("  return %s = %s || ", bindingId, bindingId)
	fnClose := "}"

	p.printNewline()
	p.print(fnHeader)
	p.printNewline()
	p.print(fnBodyStart)
	p.printRequireBody(require)
	if isDestructuring {
		p.print(".")
		p.print(bindingId)
	}
	p.printNewline()
	p.print(fnClose)
}

// similar to slocal but assigning to an already declared variable
// x = require('x')
func (p *printer) handleEBinary(e *js_ast.EBinary) (handled bool) {
	if e.Op != js_ast.BinOpAssign {
		return false
	}

	require, isRequire := p.extractRequireExpression(e.Right, 0)
	if !isRequire {
		return false
	}

	identifiers, ok := p.extractIdentifiers(e.Left.Data)
	if !ok {
		return false
	}

	for _, b := range identifiers {
		var fnName string
		var fnCall string
		var id string

		// Ensure that we don't register a replacement for a ref for which we did this already
		// Additionally the `identifierName` will not be the original one in this case so we need
		// to obtain it and then derive the dependent ids from it.
		if p.renamer.HasBeenReplaced(b.identifier) {
			id = p.renamer.GetOriginalId(b.identifier)
			fnName = functionNameForId(id)
			fnCall = functionCallForId(id)
		} else {
			id = b.identifierName
			fnName = functionNameForId(id)
			fnCall = functionCallForId(id)
			p.renamer.Replace(b.identifier, fnCall)
			p.trackTopLevelVar(fnName)
		}
		p.printRequireReplacementFunctionAssign(require, id, b.isDestructuring, fnName)
	}

	return true
}
