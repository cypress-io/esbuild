package snap_printer

import "testing"

func TestElinkSimpleRequire(t *testing.T) {
	expectPrinted(t, `
const a = require('a')
const b = require('b')
function main () {
  const c = {a: b, b: a}
  return a + b
}
    `, `
let a;
function __get_a__() {
  return a = a || require("a")
}
const b = require("b");
function main() {
  const c = {a: b, b: __get_a__()};
  return __get_a__() + b;
}
`,
		func(mod string) bool { return mod == "a" })
}
