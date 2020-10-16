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
