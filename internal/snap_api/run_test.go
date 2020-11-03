package snap_api

import (
	"github.com/evanw/esbuild/pkg/api"
	"testing"
)

var entryPoints = []string{"../../examples/express-app/snap.js"}

// var entryPoints =  []string{"input.js"}

func TestRunJS(t *testing.T) {
	result := api.Build(api.BuildOptions{
		EntryPoints: entryPoints,
		Outfile:     "output_js.js",

		Platform: api.PlatformNode,
		Bundle:   true,
		Write:    true,
		LogLevel: api.LogLevelInfo,
	})

	if len(result.Errors) > 0 {
		t.FailNow()
	}
}

func TestRunSnap(t *testing.T) {
	result := api.Build(api.BuildOptions{
		// https://esbuild.github.io/api/#log-level
		LogLevel: api.LogLevelInfo,

		// https://esbuild.github.io/api/#target
		Target: api.ES2020,

		// inline any imported dependencies into the file itself
		// https://esbuild.github.io/api/#bundle
		Bundle: true,

		// write out a JSON file with metadata about the build
		// https://esbuild.github.io/api/#metafile
		Metafile: "meta_snap.json",

		// Applies when multiple entry points are used.
		// https://esbuild.github.io/api/#outdir
		Outdir: "",
		// Applies when one entry point is used.
		// https://esbuild.github.io/api/#outfile
		Outfile:     "output_snap.js",
		EntryPoints: entryPoints,

		// https://esbuild.github.io/getting-started/#bundling-for-node
		// https://esbuild.github.io/api/#platform
		//
		// Setting to Node results in:
		// - the default output format is set to cjs
		// - built-in node modules such as fs are automatically marked as external
		// - disables the interpretation of the browser field in package.json
		Platform: api.PlatformNode,
		Engines: []api.Engine{
			{api.EngineNode, "12.4"},
		},

		// https://esbuild.github.io/api/#format
		// three possible values: iife, cjs, and esm
		Format: api.FormatCommonJS,

		// the import will be preserved and will be evaluated at run time instead
		// https://esbuild.github.io/api/#external
		External: []string{"inherits"},

		//
		// Combination of the below two might be a better way to replace globals
		// while taking the snapshot
		// We'd copy the code for each from the electron blueprint and add it to
		// a module which we use to inject.
		//

		// replace a global variable with an import from another file.
		// https://esbuild.github.io/api/#inject
		// i.e. Inject:      []string{"./process-shim.js"},
		Inject: nil,

		// replace global identifiers with constant expressions
		// https://esbuild.github.io/api/#define
		// i.e.: Define: map[string]string{"DEBUG": "true"},
		Define: nil,

		// When `false` a buffer is returned instead
		// https://esbuild.github.io/api/#write
		Write: true,

		Snapshot: true,

		//
		// Unused
		//

		// only matters when the format setting is iife
		GlobalName: "",

		Sourcemap: 0,

		// Only works with ESM modules
		// https://esbuild.github.io/api/#splitting
		Splitting: false,

		MinifyWhitespace:  false,
		MinifyIdentifiers: false,
		MinifySyntax:      false,

		JSXFactory:  "",
		JSXFragment: "",

		// Temporal Dead Zone related perf tweak (var vs let)
		// https://esbuild.github.io/api/#avoid-tdz
		AvoidTDZ: false,

		// https://esbuild.github.io/api/#charset
		Charset: 0,

		// https://esbuild.github.io/api/#color
		Color: 0,

		// https://esbuild.github.io/api/#error-limit
		ErrorLimit: 0,


		// additional package.json fields to try when resolving a package
		// https://esbuild.github.io/api/#main-fields
		MainFields: nil,

		// https://esbuild.github.io/api/#out-extension
		OutExtensions: nil,

		// useful in combination with the external file loader
		// https://esbuild.github.io/api/#public-path
		PublicPath: "",

		// /* #__PURE__ */ before a new or call expression means that that
		// expression can be removed
		// https://esbuild.github.io/api/#pure
		Pure: nil,

		// Tweak resolution algorithm used by node via implicit file extensions
		// https://esbuild.github.io/api/#resolve-extensions
		ResolveExtensions: nil,
		Loader:            nil,

		// Use stdin as input instead of a file
		// https://esbuild.github.io/api/#stdin
		Stdin: nil,

		Tsconfig: "",
	})

	if len(result.Errors) > 0 {
		t.FailNow()
	}
}

/*

--- Encountered Issues ---

# depd: deprecate package

// ../../examples/express-app/node_modules/depd/index.js

## Process Reference

	var basePath = process.cwd()

Even though `process` reference is rewritten, it is resolved at module level.
However `basePath` is used inside functions, thus a solution would be to rewrite
references to results obtained from globals like we do with requires.

# inherits: wrapper

// ../../examples/express-app/node_modules/inherits/inherits.js

Resolves `util.inherits` at module level to know what to export.

## Require Rewrites

Making 'inherits' an external via `External:    []string{"inherits"},`
caused it to not be included in the bundle which possibly is the only solution here.

# http-errors

// ../../examples/express-app/node_modules/http-errors/index.js

## Require Rewrites

	var deprecate = require('depd')('http-errors')
->
	var deprecate = require_depd()("http-errors");
->
   snap_printer doesn't recognize it as `require` call and doesn't rewrite it.
   others that don't include an immediate call are rewritten correctly
*/
