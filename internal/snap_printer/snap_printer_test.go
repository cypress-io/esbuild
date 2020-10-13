package snap_printer

import "testing"

func TestIsolatedRequireRewrites(t *testing.T) {
	expectPrinted(t, "const foo = require('./foo')", `
let foo;
function __get_foo__() {
  return foo = foo || require("./foo")
}
`, ReplaceAll)

	expectPrinted(t, `
 const foo = require('./foo'),
   bar = require('./bar')
 `, `
let foo;
function __get_foo__() {
  return foo = foo || require("./foo")
}

let bar;
function __get_bar__() {
  return bar = bar || require("./bar")
}
`, ReplaceAll)
}

func TestIntegratedRequireRewrites(t *testing.T) {
	expectPrinted(t, `
const a = 1
const foo = require('./foo')
const b = 'hello world'
`, `
const a = 1;

let foo;
function __get_foo__() {
  return foo = foo || require("./foo")
}
const b = "hello world";
`, ReplaceAll)

	expectPrinted(t, `
const foo = require('./foo'),
  a = 1,
  bar = require('./bar'),
  b = 'hello world'
`, `
let foo;
function __get_foo__() {
  return foo = foo || require("./foo")
}
const a = 1
let bar;
function __get_bar__() {
  return bar = bar || require("./bar")
}
const b = "hello world"`,
		ReplaceAll)
}

func TestRequireReferences(t *testing.T) {
	expectPrinted(t, `
const foo = require('./foo')
function logFoo() {
  console.log(foo.bar)
}
`, `
let foo;
function __get_foo__() {
  return foo = foo || require("./foo")
}
function logFoo() {
  console.log(__get_foo__().bar);
}
`, ReplaceAll)
}
