package snap_api

import (
	"testing"

	"github.com/evanw/esbuild/internal/snap_printer"
)

var snapApiSuite = suite{
	name: "Snap API",
}

func TestEntryRequiringLocalModule(t *testing.T) {
	snapApiSuite.expectBuild(t, built{
		files: map[string]string{
			ProjectBaseDir + "/entry.js": `
				const { oneTwoThree } = require('./foo')
                module.exports = function () {
				  console.log(oneTwoThree)
			    }
			`,
			ProjectBaseDir + "/foo.js": `exports.oneTwoThree = 123`,
		},
		entryPoints: []string{ProjectBaseDir + "/entry.js"},
	},
		buildResult{
			files: map[string]string{
				ProjectBaseDir + "/entry.js": `
__commonJS["./entry.js"] = function(exports, module2, __filename, __dirname, require) {
let oneTwoThree;
function __get_oneTwoThree__() {
  return oneTwoThree = oneTwoThree || (require("./foo", "./foo.js", (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname)).oneTwoThree)
}
  module2.exports = function() {
    get_console().log((__get_oneTwoThree__()));
  };
};`,
				ProjectBaseDir + `/foo.js`: `
__commonJS["./foo.js"] = function(exports, module2, __filename, __dirname, require) {
  exports.oneTwoThree = 123;
};`,
			},
		},
	)
}

func TestTypeScriptImportingDefaultModule(t *testing.T) {
	snapApiSuite.expectBuild(t, built{
		files: map[string]string{
			ProjectBaseDir + "/entry.js": `
				import Debug from './debug'
				const debug = Debug('foo')
			`,
			ProjectBaseDir + "/debug.js": `module.exports = function () {}`,
		},
		entryPoints: []string{ProjectBaseDir + "/entry.js"},
	},
		buildResult{
			files: map[string]string{
				ProjectBaseDir + "/debug.js": `
__commonJS["./debug.js"] = function(exports, module2, __filename, __dirname, require) {
  module2.exports = function() {
  };
};`,
				ProjectBaseDir + "/entry.js": `
__commonJS["./entry.js"] = function(exports, module2, __filename, __dirname, require) {
let import_debug;
function __get_import_debug__() {
  return import_debug = import_debug || (__toModule(require("./debug", "./debug.js", (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname))))
}
let debug;
function __get_debug__() {
  return debug = debug || ((0, (__get_import_debug__()).default)("foo"))
}
};`,
			},
		},
	)
}

func TestJavaScriptRequiringDefaultModule(t *testing.T) {
	snapApiSuite.expectBuild(t, built{
		files: map[string]string{
			ProjectBaseDir + "/entry.js": `
				const Debug = require('./debug')
				const debug = Debug('foo')
			`,
			ProjectBaseDir + "/debug.js": `module.exports = function () {}`,
		},
		entryPoints: []string{ProjectBaseDir + "/entry.js"},
	},
		buildResult{
			files: map[string]string{
				ProjectBaseDir + "/debug.js": `
__commonJS["./debug.js"] = function(exports, module2, __filename, __dirname, require) {
  module2.exports = function() {
  };
};`,
				ProjectBaseDir + "/entry.js": `
__commonJS["./entry.js"] = function(exports, module2, __filename, __dirname, require) {
let Debug;
function __get_Debug__() {
  return Debug = Debug || (require("./debug", "./debug.js", (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname)))
}
let debug;
function __get_debug__() {
  return debug = debug || ((__get_Debug__())("foo"))
}
};`,
			},
		},
	)
}

func TestEntryImportingLocalModule(t *testing.T) {
	snapApiSuite.expectBuild(t, built{
		files: map[string]string{
			ProjectBaseDir + "/entry.js": `
				import { oneTwoThree } from'./foo'
                module.exports = function () {
				  console.log(oneTwoThree)
			    }
			`,
			ProjectBaseDir + "/foo.js": `exports.oneTwoThree = 123`,
		},
		entryPoints: []string{ProjectBaseDir + "/entry.js"},
	},
		buildResult{
			files: map[string]string{
				ProjectBaseDir + `/foo.js`: `
__commonJS["./foo.js"] = function(exports, module2, __filename, __dirname, require) {
  exports.oneTwoThree = 123;
};`,
				ProjectBaseDir + `/entry.js`: `
__commonJS["./entry.js"] = function(exports, module2, __filename, __dirname, require) {
let import_foo;
function __get_import_foo__() {
  return import_foo = import_foo || (__toModule(require("./foo", "./foo.js", (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname))))
}
  module2.exports = function() {
    get_console().log((__get_import_foo__()).oneTwoThree);
  };
};`,
			},
		},
	)
}
func TestCallingResultOfRequiringModule(t *testing.T) {
	snapApiSuite.expectBuild(t, built{
		files: map[string]string{
			ProjectBaseDir + "/entry.js": `
var deprecate = require('./depd')('http-errors')
module.exports = function () { deprecate() }
`,
			ProjectBaseDir + "/depd.js": "module.exports = function (s) {}",
		},
		entryPoints: []string{ProjectBaseDir + "/entry.js"},
	},

		buildResult{
			files: map[string]string{
				ProjectBaseDir + `/entry.js`: `
__commonJS["./entry.js"] = function(exports, module2, __filename, __dirname, require) {
let deprecate;
function __get_deprecate__() {
  return deprecate = deprecate || (require("./depd", "./depd.js", (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname))("http-errors"))
}
  module2.exports = function() {
    (__get_deprecate__())();
  };
};`,
			},
		},
	)
}

func TestNotWrappingExports(t *testing.T) {
	snapApiSuite.expectBuild(t,
		built{
			files: map[string]string{
				ProjectBaseDir + "/entry.js":       `require('./body-parser')`,
				ProjectBaseDir + "/body-parser.js": `exports = module.exports = foo()`,
			},
			entryPoints: []string{ProjectBaseDir + "/entry.js"},
		},
		buildResult{
			files: map[string]string{
				ProjectBaseDir + "/body-parser.js": `
__commonJS["./body-parser.js"] = function(exports, module2, __filename, __dirname, require) {
  exports = module2.exports = foo();
};`,
				ProjectBaseDir + "/entry.js": `
__commonJS["./entry.js"] = function(exports, module2, __filename, __dirname, require) {
  require("./body-parser", "./body-parser.js", (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname));
};`,
			},
		},
	)
}

func TestDeclarationsInsertedAfterUseStrict(t *testing.T) {
	snapApiSuite.expectBuild(t, built{
		files: map[string]string{
			ProjectBaseDir + "/entry.js": `
"use strict";
var old;
old = Promise;
`,
		},
		entryPoints: []string{ProjectBaseDir + "/entry.js"},
	},
		buildResult{
			files: map[string]string{
				ProjectBaseDir + `/entry.js`: `
__commonJS["./entry.js"] = function(exports, module2, __filename, __dirname, require) {
  "use strict";
let __get_old__;  var old;
  
__get_old__ = function() {
  return old = old || (Promise)
};
};`,
			},
		},
	)
}

func TestMissingFileRequiredOnlyWarns(t *testing.T) {
	snapApiSuite.expectBuild(t, built{
		files: map[string]string{
			ProjectBaseDir + "/entry.js": `
require('non-existent')
`,
		},
		entryPoints: []string{ProjectBaseDir + "/entry.js"},
	},
		buildResult{
			files: map[string]string{
				ProjectBaseDir + `/entry.js`: `
__commonJS["./entry.js"] = function(exports, module2, __filename, __dirname, require) {
  require("non-existent", "non-existent", (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname));
};`,
			},
		})
}

// @see https://github.com/evanw/esbuild/commit/918d44e7e2912fa23f9ba409e1d6623275f7b83f
func TestNestedScopeVarsAreNotRelocated(t *testing.T) {
	snapApiSuite.expectBuild(t, built{
		files: map[string]string{
			ProjectBaseDir + "/entry.js": `
{ var obj = Array.from({}) }
`,
		},
		entryPoints: []string{ProjectBaseDir + "/entry.js"},
	},
		buildResult{
			files: map[string]string{
				ProjectBaseDir + `/entry.js`: `
__commonJS["./entry.js"] = function(exports, module2, __filename, __dirname, require) {
  {
let obj;
function __get_obj__() {
  return obj = obj || (Array.from({}))
}
  }
};`,
			},
		},
	)
}

func TestShouldRewriteModule(t *testing.T) {
	snapApiSuite.expectBuild(t, built{
		shouldRewriteModule: func(filePath string) bool {
			return filePath != ProjectBaseDir[1:]+"/foo.js"
		},
		files: map[string]string{
			ProjectBaseDir + "/foo.js": `var fs = require('fs')`,
			ProjectBaseDir + "/bar.js": `var path = require('path')`,
			ProjectBaseDir + "/entry.js": `
exports.foo = require('./foo')
exports.bar = require('./bar')
`,
		},
		entryPoints: []string{ProjectBaseDir + "/entry.js"},
	},
		buildResult{
			files: map[string]string{
				`dev/foo.js`: `
__commonJS["./foo.js"] = function(exports, module, __filename, __dirname, require) {
  var fs = require("fs", "fs", (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname));
};`,
				`dev/bar.js`: `
__commonJS["./bar.js"] = function(exports, module2, __filename, __dirname, require) {
let path;
function __get_path__() {
  return path = path || (require("path", "path", (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname)))
}
};`,
				`dev/entry.js`: `
__commonJS["./entry.js"] = function(exports, module2, __filename, __dirname, require) {
  Object.defineProperty(exports, "foo", { get: () => require("./foo", "./foo.js", (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname)) });
  Object.defineProperty(exports, "bar", { get: () => require("./bar", "./bar.js", (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname)) });
};`,
			},
		},
	)
}

func TestPreventResolutionOfNativeModules(t *testing.T) {
	snapApiSuite.expectBuild(t, built{
		shouldRewriteModule: func(filePath string) bool {
			return false
		},
		files: map[string]string{
			ProjectBaseDir + "/node_modules/fsevents/fsevents.js": `
const Native = require('./fsevents.node');
const events = Native.constants;
`,
			ProjectBaseDir + "/entry.js": `
exports.fsevents = require('` + ProjectBaseDir + `/node_modules/fsevents/fsevents.js')
`,
		},
		entryPoints: []string{ProjectBaseDir + "/entry.js"},
	},
		buildResult{
			files: map[string]string{
				`dev/node_modules/fsevents/fsevents.js`: `
__commonJS["./node_modules/fsevents/fsevents.js"] = function(exports, module, __filename, __dirname, require) {
  var Native = require("./node_modules/fsevents/fsevents.node", "./node_modules/fsevents/fsevents.node", (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname));
  var events = Native.constants;
};`,
				`dev/entry.js`: `
__commonJS["./entry.js"] = function(exports, module, __filename, __dirname, require) {
  exports.fsevents = require("/dev/node_modules/fsevents/fsevents.js", "./node_modules/fsevents/fsevents.js", (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname));
};`,
			},
		},
	)
}

// NOTE: this test documents that the need to defer fsevents isn't detected here
// Instead while determining that the `__resolve_path` function needs to throw so
// that the snapshot verifier ends up deferring fsevents.
func TestWrapsDirnameAccessOnInitAndDoesNotDeferModule(t *testing.T) {
	snapApiSuite.expectBuild(t, built{
		shouldReplaceRequire: snap_printer.ReplaceNone,
		files: map[string]string{
			ProjectBaseDir + "/node_modules/fsevents/fsevents.js": `
module.exports = __dirname
`,
			ProjectBaseDir + "/entry.js": `
exports.fsevents = require('` + ProjectBaseDir + `/node_modules/fsevents/fsevents.js')
`,
		},
		entryPoints: []string{ProjectBaseDir + "/entry.js"},
	},
		buildResult{
			files: map[string]string{
				`dev/node_modules/fsevents/fsevents.js`: `
__commonJS["./node_modules/fsevents/fsevents.js"] = function(exports, module2, __filename, __dirname, require) {
  module2.exports = __resolve_path(typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname);
};`,
				`dev/entry.js`: `
__commonJS["./entry.js"] = function(exports, module2, __filename, __dirname, require) {
  exports.fsevents = require("/dev/node_modules/fsevents/fsevents.js", "./node_modules/fsevents/fsevents.js", (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname));
};`,
			},
		},
	)
}

func TestWrapsFilenameDelayedAccessAndDoesNotDeferModule(t *testing.T) {
	snapApiSuite.expectBuild(t, built{
		shouldReplaceRequire: snap_printer.ReplaceNone,
		files: map[string]string{
			ProjectBaseDir + "/node_modules/file-url.js": `
      module.exports = function foo() {
return  'file://' + __filename 
}
`,
			ProjectBaseDir + "/entry.js": `
exports.fileUrl = require('` + ProjectBaseDir + `/node_modules/file-url.js')
`,
		},
		entryPoints: []string{ProjectBaseDir + "/entry.js"},
	},
		buildResult{
			files: map[string]string{
				`dev/node_modules/file-url.js`: `
__commonJS["./node_modules/file-url.js"] = function(exports, module2, __filename, __dirname, require) {
  module2.exports = function foo() {
    return "file://" + __resolve_path(typeof __filename2 !== 'undefined' ? __filename2 : __filename);
  };
};`,
				`dev/entry.js`: `
__commonJS["./entry.js"] = function(exports, module2, __filename, __dirname, require) {
  exports.fileUrl = require("/dev/node_modules/file-url.js", "./node_modules/file-url.js", (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname));
};`,
			},
		},
	)
}

func TestReassignCoupledWithUseOfConsole(t *testing.T) {
	snapApiSuite.expectBuild(t, built{
		files: map[string]string{
			ProjectBaseDir + "/fine.js": `console.log('fine')`,
			ProjectBaseDir + "/reassigns-console.js": `
			console = function () {}
			console.log('reassigned')
	`,
			ProjectBaseDir + "/entry.js": `
module.exports = function () {
  require('./fine')
  require('./reassigns-console')
}
`},
		entryPoints: []string{ProjectBaseDir + "/entry.js"},
	},
		buildResult{
			files: map[string]string{
				`dev/fine.js`: `
__commonJS["./fine.js"] = function(exports, module2, __filename, __dirname, require) {
  get_console().log("fine");
};`,
				`dev/reassigns-console.js`: `
__commonJS["./reassigns-console.js"] = function(exports, module2, __filename, __dirname, require) {
  console = function() {
  };
  get_console().log("reassigned");
};`,
				`dev/entry.js`: `
__commonJS["./entry.js"] = function(exports, module2, __filename, __dirname, require) {
  module2.exports = function() {
    require("./fine", "./fine.js", (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname));
    require("./reassigns-console", "./reassigns-console.js", (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname));
  };
};`,
			},
		},
	)
}

func TestReportsNoRewriteValidationErrorsAsWarnings(t *testing.T) {
	snapApiSuite.expectWarnings(t, built{
		files: map[string]string{
			ProjectBaseDir + "/entry.js": `
function override() {} 
process.emitWarning = override 
`,
		},
		entryPoints: []string{ProjectBaseDir + "/entry.js"},
	}, []string{
		"[SNAPSHOT_REWRITE_FAILURE] Cannot override 'process.emitWarning'",
	},
	)
}

func TestRequireResolveRewrite(t *testing.T) {
	snapApiSuite.expectBuild(t, built{
		files: map[string]string{
			ProjectBaseDir + "/fixtures/sync-deps.js": `module.exports = 1`,
			ProjectBaseDir + "/foo.js":                `module.exports = 1`,
			ProjectBaseDir + "/entry.js": `
const fooPath = require.resolve('./foo')
require.resolve('./foo')
delete require.cache[require.resolve('./fixtures/sync-deps.js')]
function toBeResolved(prefix) {
  return prefix + 'foo'
}
require.resolve(toBeResolved('./'))
`,
		},
		entryPoints: []string{ProjectBaseDir + "/entry.js"},
	},
		buildResult{
			files: map[string]string{
				`dev/entry.js`: `
__commonJS["./entry.js"] = function(exports, module2, __filename, __dirname, require) {
  var fooPath = require.resolve("./foo", (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname));
  require.resolve("./foo", (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname));
  delete require.cache[require.resolve("./fixtures/sync-deps.js", (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname))];
  function toBeResolved(prefix) {
    return prefix + "foo";
  }
  require.resolve(toBeResolved("./"), (typeof __filename2 !== 'undefined' ? __filename2 : __filename), (typeof __dirname2 !== 'undefined' ? __dirname2 : __dirname));
};`,
			},
		},
	)
}

func TestExportDeferredLocalVar(t *testing.T) {
	snapApiSuite.expectBuild(t, built{
		files: map[string]string{
			ProjectBaseDir + "/entry.js": `
export const cwd = process.cwd()
`,
		},
		entryPoints: []string{ProjectBaseDir + "/entry.js"},
	},
		buildResult{
			files: map[string]string{
				`dev/entry.js`: `
__commonJS["./entry.js"] = function(exports, module, __filename, __dirname, require) {
  __markAsModule(exports);
  __export(exports, {
    cwd: () => cwd
  });
  var cwd = get_process().cwd();
};`,
			},
		},
	)
}

func TestDebug(t *testing.T) {
	snapApiSuite.debugBuild(t, built{
		files: map[string]string{
			ProjectBaseDir + "/entry.js": `
	const a = __dirname
	module.exports = a
`,
		},
		entryPoints: []string{ProjectBaseDir + "/entry.js"},
	},
	)
}

func TestCreateShouldRewriteModule(t *testing.T) {
	tests := []struct {
		name     string
		args     *SnapCmdArgs
		module   string
		expected bool
	}{
		{
			name:     "empty module path",
			args:     &SnapCmdArgs{},
			module:   "",
			expected: true,
		},
		{
			name: "exact match with norewrite",
			args: &SnapCmdArgs{
				Norewrite: []string{"react"},
			},
			module:   "react",
			expected: false,
		},
		{
			name: "node_modules prefix no node_modules",
			args: &SnapCmdArgs{
				Norewrite: []string{"*/node_modules/react/dist/index.js"},
			},
			module:   "node_modules/react/dist/index.js",
			expected: false,
		},
		{
			name: "* prefix nested node_modules",
			args: &SnapCmdArgs{
				Norewrite: []string{"*/node_modules/react/dist/index.js"},
			},
			module:   "packages/server/node_modules/react/dist/index.js",
			expected: false,
		},
		{
			name: "no match with norewrite",
			args: &SnapCmdArgs{
				Norewrite: []string{"react"},
			},
			module:   "vue",
			expected: true,
		},
		{
			name: "multiple norewrite entries matching nested dependencies",
			args: &SnapCmdArgs{
				Norewrite: []string{"react", "*/node_modules/vue/dist/index.js"},
			},
			module:   "packages/app/node_modules/vue/dist/index.js",
			expected: false,
		},
		{
			name: "multiple norewrite entries with no match",
			args: &SnapCmdArgs{
				Norewrite: []string{"react", "*/node_modules/vue/dist/file.js"},
			},
			module:   "packages/app/node_modules/vue/dist/index.js",
			expected: true,
		},
	}

	for _, tt := range tests {
		for _, prefix := range []string{"", "./"} {
			t.Run(tt.name, func(t *testing.T) {
				predicate := CreateShouldRewriteModule(tt.args)
				result := predicate(prefix + tt.module)
				if result != tt.expected {
					t.Errorf("CreateShouldRewriteModule() = %v, want %v for module %q", result, tt.expected, tt.module)
				}
			})
		}
	}
}
