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
};`, ReplaceAll)
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
`, `
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

func TestRequireDeclPropertyChain(t *testing.T) {
	expectPrinted(t, `
const bar = require('foo').bar
`, `
let bar;
function __get_bar__() {
  return bar = bar || require("foo").bar
}
`, ReplaceAll)

	expectPrinted(t, `
const baz = require('foo').bar.baz
`, `
let baz;
function __get_baz__() {
  return baz = baz || require("foo").bar.baz
}
`, ReplaceAll)
}

func TestRequireLateAssignmentPropertyChain(t *testing.T) {
	expectPrinted(t, `
let bar
bar = require('foo').bar
`, `
let __get_bar__;
let bar;

__get_bar__ = function() {
  return bar = bar || require("foo").bar
};
`, ReplaceAll)

	expectPrinted(t, `
let baz
baz = require('foo').bar.baz
`, `
let __get_baz__;
let baz;

__get_baz__ = function() {
  return baz = baz || require("foo").bar.baz
};
`, ReplaceAll)
}

func TestDestructuringDeclarationReferenced(t *testing.T) {
	expectPrinted(t, `
const { foo, bar } = require('foo-bar')
function id() {
  foo.id = 'hello'
}
`, `
let foo;
function __get_foo__() {
  return foo = foo || require("foo-bar").foo
}

let bar;
function __get_bar__() {
  return bar = bar || require("foo-bar").bar
}
function id() {
  __get_foo__().id = "hello";
}
`, ReplaceAll)
}

func TestDestructuringLateAssignmentReferenced(t *testing.T) {
	expectPrinted(t, `
let foo, bar;
({ foo, bar } = require('foo-bar'))
function id() {
  foo.id = 'hello'
}
`, `
let __get_foo__, __get_bar__;
let foo, bar;

__get_foo__ = function() {
  return foo = foo || require("foo-bar").foo
}
__get_bar__ = function() {
  return bar = bar || require("foo-bar").bar
};
function id() {
  __get_foo__().id = "hello";
}
`, ReplaceAll)
}

func TestAssignToSameVarConditionallyAndReferenceIt(t *testing.T) {
	expectPrinted(t, `
let a
if (condition) {
  a = require('a')
} else { 
  a = require('b')
}
function foo() {
  a.b = 'c'
}
`, `
let __get_a__;
let a;
if (condition) {
  
__get_a__ = function() {
  return a = a || require("a")
};
} else {
  
__get_a__ = function() {
  return a = a || require("b")
};
}
function foo() {
  __get_a__().b = "c";
}
`, ReplaceAll)
}

func TestVarAssignedToRequiredVarAndReferenced(t *testing.T) {
	expectPrinted(t, `
const a = require('a')
const b = a
function main() {
  b.c = 1
}
`, `
let a;
function __get_a__() {
  return a = a || require("a")
}

let b;
function __get_b__() {
  return b = b || __get_a__()
}
function main() {
  __get_b__().c = 1;
}
`, ReplaceAll)

}

func TestVarAssignedToPropertyOfRequiredVarAndReferenced(t *testing.T) {
	expectPrinted(t, `
const a = require('a')
const b = a.foo
function main() {
  b.c = 1
}
`, `
let a;
function __get_a__() {
  return a = a || require("a")
}

let b;
function __get_b__() {
  return b = b || __get_a__().foo
}
function main() {
  __get_b__().c = 1;
}
`, ReplaceAll)

}
func TestDestructuredVarsAssignedToPropertyOfRequiredVarAndReferenced(t *testing.T) {
	expectPrinted(t, `
const a = require('a')
const { foo, bar } = a
function main() {
  return foo + bar 
}
`, `
let a;
function __get_a__() {
  return a = a || require("a")
}

let foo;
function __get_foo__() {
  return foo = foo || __get_a__().foo
}

let bar;
function __get_bar__() {
  return bar = bar || __get_a__().bar
}
function main() {
  return __get_foo__() + __get_bar__();
}
`, ReplaceAll)
}

func TestVarsInSingleDeclarationReferencingEachOtherReferenced(t *testing.T) {
	expectPrinted(t, `
let a = require('a'), b  = a.c
function main() {
  return a + b 
}
`, `
let a;
function __get_a__() {
  return a = a || require("a")
}

let b;
function __get_b__() {
  return b = b || __get_a__().c
}
function main() {
  return __get_a__() + __get_b__();
}
`, ReplaceAll)
}

func TestLateAssignmentToRequireReference(t *testing.T) {
	expectPrinted(t, `
const a = require('a')
let b
b = a.c
function main() {
  return a + b 
}
`, `
let __get_b__;

let a;
function __get_a__() {
  return a = a || require("a")
}
let b;

__get_b__ = function() {
  return b = b || __get_a__().c
};
function main() {
  return __get_a__() + __get_b__();
}
`, ReplaceAll)
}

func TestIndirectReferencesToRequireInSameDeclaration(t *testing.T) {
	expectPrinted(t, `
let d = require("d"), e = d.e, f = e.f;
`, `
let d;
function __get_d__() {
  return d = d || require("d")
}

let e;
function __get_e__() {
  return e = e || __get_d__().e
}

let f;
function __get_f__() {
  return f = f || __get_e__().f
}
`, ReplaceAll)
}

func TestIndirectReferencesToRequireLateAssign(t *testing.T) {
	expectPrinted(t, `
let d, e, f;
d = require("d");
e = d.e;
f = e.f;
`, `
let __get_d__, __get_e__, __get_f__;
let d, e, f;

__get_d__ = function() {
  return d = d || require("d")
};

__get_e__ = function() {
  return e = e || __get_d__().e
};

__get_f__ = function() {
  return f = f || __get_e__().f
}; `, ReplaceAll)
}

func TestDeclarationToCallResultWithRequireReferenceArgReferenced(t *testing.T) {
	expectPrinted(t, `
var pack = require('pack')
const x = someCall(pack);
function main() {
  return x + 1
}
`, `
let pack;
function __get_pack__() {
  return pack = pack || require("pack")
}

let x;
function __get_x__() {
  return x = x || someCall(__get_pack__())
}
function main() {
  return __get_x__() + 1;
}
`, ReplaceAll)
}

func TestDeclarationWithEBinaryReferencingRequire(t *testing.T) {
	expectPrinted(t, `
const c = require('c').foo.bar
const d = c.X | c.Y | c.Z
`, `
let c;
function __get_c__() {
  return c = c || require("c").foo.bar
}

let d;
function __get_d__() {
  return d = d || __get_c__().X | __get_c__().Y | __get_c__().Z
}
`, ReplaceAll)
}

func TestTopLevelVsNestedRequiresAndReferences(t *testing.T) {
	expectPrinted(t, `
function nested() {
  const a = require('a')
}
const b = require('b')
const c = b.foo
`, `
function nested() {
  const a = require("a");
}

let b;
function __get_b__() {
  return b = b || require("b")
}

let c;
function __get_c__() {
  return c = c || __get_b__().foo
}
`, ReplaceAll)
}

func TestLateAssignedTopLevelVsNestedRequiresAndReferences(t *testing.T) {
	expectPrinted(t, `
function nested() {
  let a
  a = require('a')
}
let b, c
b = require('b')
c = b.foo
`, `
let __get_b__, __get_c__;
function nested() {
  let a;
  a = require("a");
}
let b, c;

__get_b__ = function() {
  return b = b || require("b")
};

__get_c__ = function() {
  return c = c || __get_b__().foo
};
`, ReplaceAll)
}

func TestRequireReferencesInsideBlock(t *testing.T) {
	expectPrinted(t, `
{
  const a = require('a')
  const c = a.bar
}
`, `
{

let a;
function __get_a__() {
  return a = a || require("a")
}

let c;
function __get_c__() {
  return c = c || __get_a__().bar
}
}
`, ReplaceAll)
}

func TestRequireWithCallchain(t *testing.T) {
	expectPrinted(t, `
 var debug = require('debug')('express:view')
`, `
let debug;
function __get_debug__() {
  return debug = debug || require("debug")("express:view")
}
`, ReplaceAll)

	expectPrinted(t, `
 var chain = require('chainer')('hello')('world')(foo())(1)
`, `
let chain;
function __get_chain__() {
  return chain = chain || require("chainer")("hello")("world")(foo())(1)
}
`, ReplaceAll)
}

func TestRequireWithCallchainAndPropChain(t *testing.T) {
	expectPrinted(t, `
 var chain = require('chainer')('hello').foo.bar
`, `
let chain;
function __get_chain__() {
  return chain = chain || require("chainer")("hello").foo.bar
}
`, ReplaceAll)
}

