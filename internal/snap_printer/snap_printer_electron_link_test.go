package snap_printer

import "testing"

// test('simple require')
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

// test('conditional requires')
func TestElinkConditionalRequires(t *testing.T) {
	expectPrinted(t, `
let a, b;
if (condition) {
  a = require('a')
  b = require('b')
} else {
  a = require('c')
  b = require('d')
}

function main () {
  return a + b
}
    `, `
let __get_a__;
let a, b;
if (condition) {
  
__get_a__ = function() {
  return a = a || require("a")
};
  b = require("b");
} else {
  
__get_a__ = function() {
  return a = a || require("c")
};
  b = require("d");
}
function main() {
  return __get_a__() + b;
}
`,
		func(mod string) bool { return mod == "a" || mod == "c" })
}

// TODO: not yet wrapping access to d  (line 76)
// test('top-level variables assignments that depend on previous requires')
func _TestElinkVarAssignmentsDependingOnPreviousRequires(t *testing.T) {
	debugPrinted(t, `
const a = require('a')
const b = require('b')
const c = require('c').foo.bar
const d = c.X | c.Y | c.Z
var e
e = c.e
const f = b.f
function main () {
  c.qux()
  console.log(d)
  e()
} `,
		func(mod string) bool { return mod == "a" || mod == "c" })

}

//
// Function Closures
//

// First three following are parts of the related electron-link example which is
// tested in one piece in the forth test
// test('requires that appear in a closure wrapper defined in the top-level scope (e.g. CoffeeScript)')
func TestElinkTopLevelClosureWrapperCall(t *testing.T) {
	expectPrinted(t, `
(function () {
	const a = require('a')
	const b = require('b')
	function main () {
		return a + b
	}
}).call(this)
`, `
(function() {

let a;
function __get_a__() {
  return a = a || require("a")
}

let b;
function __get_b__() {
  return b = b || require("b")
}
  function main() {
    return __get_a__() + __get_b__();
  }
}).call(this);
`, ReplaceAll)
}

func TestElinkTopLevelClosureWrapperSelfExecuteFiltered(t *testing.T) {
	expectPrinted(t, `
(function () {
  const a = require('a')
  const b = require('b')
  function main () {
    return a + b
  }
})()
`, `
(function() {

let a;
function __get_a__() {
  return a = a || require("a")
}
  const b = require("b");
  function main() {
    return __get_a__() + b;
  }
})();
`,
		func(mod string) bool { return mod == "a" },
	)
}

// NOTE: electron-link does not rewrite anything here, however this may be a mistake as
// `foo` might invoke the callback synchronously when it runs and thus execute the `require`s
func TestElinkTopLevelFunctionInvokingCallback(t *testing.T) {
	expectPrinted(t, `
foo(function () {
  const b = require('b')
  const c = require('c')
  function main () {
    return b + c
  }
})
`, `
foo(function() {

let b;
function __get_b__() {
  return b = b || require("b")
}

let c;
function __get_c__() {
  return c = c || require("c")
}
  function main() {
    return __get_b__() + __get_c__();
  }
});
`,
		ReplaceAll,
	)
}
func TestElinkTopLevelClosureCompleteFiltered(t *testing.T) {
	expectPrinted(t, `
(function () {
  const a = require('a')
  const b = require('b')
  function main () {
    return a + b
  }
}).call(this)

(function () {
  const a = require('a')
  const b = require('b')
  function main () {
    return a + b
  }
})()

foo(function () {
  const b = require('b')
  const c = require('c')
  function main () {
    return b + c
  }
})
`, `
(function() {

let a;
function __get_a__() {
  return a = a || require("a")
}
  const b = require("b");
  function main() {
    return __get_a__() + b;
  }
}).call(this)(function() {

let a;
function __get_a__() {
  return a = a || require("a")
}
  const b = require("b");
  function main() {
    return __get_a__() + b;
  }
})();
foo(function() {
  const b = require("b");

let c;
function __get_c__() {
  return c = c || require("c")
}
  function main() {
    return b + __get_c__();
  }
});
`,
		func(mod string) bool { return mod == "a" || mod == "c" })
}

// test('references to shadowed variables')
func TestElinkReferencesToShadowedVars(t *testing.T) {
	expectPrinted(t, `
const a = require('a')
function outer () {
  console.log(a)
  function inner () {
    console.log(a)
  }
  let a = []
}

function other () {
  console.log(a)
  function inner () {
    let a = []
    console.log(a)
  }
}
`, `
let a;
function __get_a__() {
  return a = a || require("a")
}
function outer() {
  get_console().log(a);
  function inner() {
    get_console().log(a);
  }
  let a = [];
}
function other() {
  get_console().log(__get_a__());
  function inner() {
    let a = [];
    get_console().log(a);
  }
}
`,
		func(mod string) bool { return mod == "a" })
}

// test('references to globals')
func TestElinkReferencesToGlobals(t *testing.T) {
	expectPrinted(t, `
global.a = 1
process.b = 2
window.c = 3
document.d = 4

function inner () {
  const window = {}
  global.e = 4
  process.f = 5
  window.g = 6
  document.h = 7
}
`, `
get_global().a = 1;
get_process().b = 2;
get_window().c = 3;
get_document().d = 4;
function inner() {
  const window = {};
  get_global().e = 4;
  get_process().f = 5;
  window.g = 6;
  get_document().h = 7;
}
`, ReplaceAll)
}

// test('multiple assignments separated by commas referencing deferred modules')
// TODO: need to wrap access to `e` by taking declarations into account that just happened before
//   and haven't been written yet
func _TestElinkMultipleAssignmentsByCommaReferencingDeferredModules(t *testing.T) {
	debugPrinted(t, `
let a, b, c, d, e, f;
a = 1, b = 2, c = 3;
d = require("d"), e = d.e, f = e.f;
`, ReplaceAll)
}

// test('require with destructuring assignment')
func TestElinkRequireWithDestructuringAssignment(t *testing.T) {
	expectPrinted(t, `
const {a, b, c} = require('module').foo

function main() {
  a.bar()
}
`, `
let a;
function __get_a__() {
  return a = a || require("module").foo.a
}

let b;
function __get_b__() {
  return b = b || require("module").foo.b
}

let c;
function __get_c__() {
  return c = c || require("module").foo.c
}
function main() {
  __get_a__().bar();
}
`, ReplaceAll)
}

// TODO: this needs to be done at another level as it is not about rewriting JS, but
//  about converting a JSON file to a JS file which exports the JSON as an object
// test('JSON source') line 322

// test('Object spread properties')
// - merely assuring that we handle it, no rewrite
func TestElinkObjectSpreadProperties(t *testing.T) {
	expectPrinted(t, `
let {a, b, ...rest} = {a: 1, b: 2, c: 3}
`, `
let {a, b, ...rest} = {a: 1, b: 2, c: 3};
`, ReplaceAll)
}

// TODO: not strictly about require rewrites, but we need to handle these cases
//   basically this is about rewriting require strings depending on a basedir
// test('path resolution') line 353

// TODO: this is an odd example which is related to vars depending on one that is
//  assigned via a require. However the example resolves that function on top level
//  which seems not entirely correct
// TODO: need to wrap `x` since expression whose result it is assigned to references `pack`
// test('use reference directly') line 417
func _TestElinkUseReferenceDirectly(t *testing.T) {
	debugPrinted(t, `
var pack = require('pack')

const x = console.log(pack);
if (condition) {
  pack
} else {
Object.keys(pack).forEach(function (prop) {
  exports[prop] = pack[prop]
})
}
`, ReplaceAll)
}

// TODO: this broke due to exports being treated like a var with a reference to a require
//  however we shouldn't defer assigning exports. The solution seems to be to disable deferring
//  assigning required references to unbound identifiers.
// test('assign to `module` or `exports`')
func _TestElinkAssignToModuleOrExports(t *testing.T) {
	expectPrinted(t, `
var pack = require('pack')      
if (condition) {
  module.exports.pack = pack
  module.exports = pack
  exports.pack = pack
  exports = pack
}
`, `
let pack;
function __get_pack__() {
  return pack = pack || require("pack")
}
if (condition) {
  module.exports.pack = __get_pack__();
  module.exports = __get_pack__();
  exports.pack = __get_pack__();
  exports = __get_pack__();
}
`, ReplaceAll)
}
