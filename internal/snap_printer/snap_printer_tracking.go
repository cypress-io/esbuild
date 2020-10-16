package snap_printer

// Tracks `let` statements that need to be inserted at the top level scope and
// the top of the file.
// This is the simplest way to ensure that the replacement functions are declared
// before they are used and accessible where needed.
// The fact that they're not declared at the exact same scope as the original identifier
// should not matter esp. since their names are unique and thus won't be shadowed.
// Example:
// ```
// let a
// a = require('a')
// ```
// becomes
// ```
// let __get_a__;
// let a;
// __get_a__ = function() {
// 	 return a = a || require("a")
// };
// ```

func (p *printer) trackTopLevelVar(decl string) {
	p.topLevelVars = append(p.topLevelVars, decl)
}

func prepend(p *printer, s string) {
	data := []byte(s)
	p.js = append(data, p.js...)

}

func (p *printer) prependTopLevelDecls() {
	if len(p.topLevelVars) == 0 { return }
	decl := "let "
	for i, v := range p.topLevelVars {
		if i > 0 { decl += ", " }
		decl += v
	}
	// TODO: consider not adding a newline here to avoid affecting source-mapped lines
	decl += ";\n"
	prepend(p, decl)
}
