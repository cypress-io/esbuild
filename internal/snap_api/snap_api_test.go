package snap_api

import (
	"testing"
)

var snapApiSuite = suite{
	name: "Snap API",
}

func TestEntryRequiringLocalModule(t *testing.T) {
	snapApiSuite.expectBuild(t, built{
		files: map[string]string{
			"/entry.js": `
				const { oneTwoThree } = require('./foo')
                module.exports = function () {
				  console.log(oneTwoThree)
			    }
			`,
			"/foo.js": `exports.oneTwoThree = 123`,
		},
		entryPoints: []string{"/entry.js"},
	},

		buildResult{
			files: map[string]string{
				`/entry.js`: `
var require_entry = __commonJS((exports, module) => {
let oneTwoThree;
function __get_oneTwoThree__() {
  return oneTwoThree = oneTwoThree || require_foo().oneTwoThree
}
  module.exports = function() {
    get_console().log(__get_oneTwoThree__());
  };
});`,
				`/foo.js`: `
var require_foo = __commonJS((exports2) => {
  exports2.oneTwoThree = 123;
});`,
			},
		},
	)
}

func TestEntryImportingLocalModule(t *testing.T) {
	snapApiSuite.expectBuild(t, built{
		files: map[string]string{
			"/entry.js": `
				import { oneTwoThree } from'./foo'
                module.exports = function () {
				  console.log(oneTwoThree)
			    }
			`,
			"/foo.js": `exports.oneTwoThree = 123`,
		},
		entryPoints: []string{"/entry.js"},
	},
		buildResult{
			files: map[string]string{
				`/foo.js`: `
var require_foo = __commonJS((exports2) => {
  exports2.oneTwoThree = 123;
});`,
				`/entry.js`: `
var require_entry = __commonJS((exports, module) => {
let foo;
function __get_foo__() {
  return foo = foo || __toModule(require_foo())
}
  module.exports = function() {
    get_console().log(__get_foo__().oneTwoThree);
  };
});`,
			},
		},
	)
}
func TestCallingResultOfRequiringModule(t *testing.T) {
	snapApiSuite.expectBuild(t, built{
		files: map[string]string{
			"/entry.js": `
var deprecate = require('./depd')('http-errors')
module.exports = function () { deprecate() }
`,
			"/depd.js": "module.exports = function (s) {}",
		},
		entryPoints: []string{"/entry.js"},
	},

		buildResult{
			files: map[string]string{
				`/entry.js`: `
var require_entry = __commonJS((exports, module) => {
let deprecate;
function __get_deprecate__() {
  return deprecate = deprecate || require_depd()("http-errors")
}
  module.exports = function() {
    __get_deprecate__()();
  };
});`,
			},
		},
	)
}

func TestNotWrappingExports(t *testing.T) {
	snapApiSuite.expectBuild(t,
		built{
			files: map[string]string{
				"/entry.js":
				`require('./body-parser')`,
				"/body-parser.js":
				`exports = module.exports = foo()`,
			},
			entryPoints: []string{"/entry.js"},
		},
		buildResult{
			files: map[string]string{
				"/body-parser.js": `
var require_body_parser = __commonJS((exports, module2) => {
  exports = module2.exports = foo();
});`,
				"/entry.js": `
var require_entry = __commonJS(() => {
  require_body_parser();
});`,
			},
		},
	)
}
