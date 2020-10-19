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
  get_console().log(__get_foo__().bar);
}
`, ReplaceAll)
}

func TestSingleLateAssignment(t *testing.T) {
	expectPrinted(t, `
let a;
a = require('a')
`, `
let __get_a__;
let a;

__get_a__ = function() {
  return a = a || require("a")
};
`, ReplaceAll)
}

func TestDoubleLateAssignment(t *testing.T) {
	expectPrinted(t, `
let a, b;
a = require('a')
b = require('b')
`, `
let __get_a__, __get_b__;
let a, b;

__get_a__ = function() {
  return a = a || require("a")
};

__get_b__ = function() {
  return b = b || require("b")
};
`, ReplaceAll)
}

func TestSingleLateAssignmentWithReference(t *testing.T) {
	expectPrinted(t, `
let a;
a = require('a')
`, `
let __get_a__;
let a;

__get_a__ = function() {
  return a = a || require("a")
};
`, ReplaceAll)
}

func TestDoubleLateAssignmentReplaceFilter(t *testing.T) {
	expectPrinted(t, `
let a, b;
a = require('a')
b = require('b')
`, `
let __get_a__;
let a, b;

__get_a__ = function() {
  return a = a || require("a")
};
b = require("b");
`, func(mod string) bool { return mod == "a" })
}

func TestConsoleReplacment(t *testing.T) {
	expectPrinted(
		t,
		`console.log('hello')`,
		`get_console().log("hello");`,
		ReplaceAll)
}

func TestProcessReplacement(t *testing.T) {
	expectPrinted(
		t,
		`process.a = 1`,
		`get_process().a = 1;`,
		ReplaceAll)
}

func TestReferencingGlobalProcessAndConstOfSameNamet(t *testing.T) {
	expectPrinted(
		t,
		`
{
  process.a = 1
}
{
  const process = {}
  process.b = 1
}
`,`
{
  get_process().a = 1;
}
{
  const process = {};
  process.b = 1;
}
`,
		ReplaceAll)
}
