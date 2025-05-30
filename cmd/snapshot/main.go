package main

import (
	"github.com/evanw/esbuild/internal/snap_api"
	"github.com/evanw/esbuild/pkg/api"
)

func main() {
	snap_api.SnapCmd(nodeJavaScript)
}

func nodeJavaScript(args *snap_api.SnapCmdArgs) api.BuildResult {
	platform := api.PlatformNode
	var external = []string{
		// should always be excluded
		"electron",
		// Causes numerous problems including FATAL:v8_context_snapshot_impl.cc(229)] Unknown WrapperTypeInfo
		// when running mksnapshot
		"bluebird",
	}

	shouldReplaceRequire := func(mdl string) bool {
		if args.Deferred == nil {
			return false
		}
		for _, m := range args.Deferred {
			if m == mdl {
				return true
			}
		}
		return false
	}

	// TODO(rebase): still needed?
	// HACK: this is needed to make esbuild include the metafile with the out files in the
	// result. I'm not sure how that works with the `{ write: false }` JS API.
	// Additionally in that case the `Outdir` needs to be set as well.
	// Note however that despite all this nothing is ever written and all paths are changed
	// to `<stdout>` when writing output files to JSON (see `snap_api/snap_cmd_helpers.go`)
	outdir := ""
	if !args.Write {
		outdir = "/"
	}

	sourcemap := api.SourceMapNone
	if args.Sourcemap != "" {
		sourcemap = api.SourceMapExternal
	}

	shouldRewriteModule := snap_api.CreateShouldRewriteModule(args)

	return api.Build(api.BuildOptions{
		// https://esbuild.github.io/api/#log-level
		LogLevel: api.LogLevelInfo,

		// https://esbuild.github.io/api/#target
		Target: api.ES2020,

		// inline any imported dependencies into the file itself
		// https://esbuild.github.io/api/#bundle
		Bundle: true,

		// https://esbuild.github.io/api/#outdir
		Outdir: outdir,

		// include JSON file with metadata about the build with the result
		// https://esbuild.github.io/api/#metafile
		Metafile: args.Metafile,

		// Applies when one entry point is used.
		// https://esbuild.github.io/api/#outfile
		Outfile:     args.Outfile,
		EntryPoints: []string{args.Entryfile},

		// https://esbuild.github.io/getting-started/#bundling-for-node
		// https://esbuild.github.io/api/#platform
		//
		// Setting to Node results in:
		// - the default output format is set to cjs
		// - built-in node modules such as fs are automatically marked as external
		// - disables the interpretation of the browser field in package.json
		Platform: platform,
		Engines: []api.Engine{
			{Name: api.EngineNode, Version: "12.4"},
		},

		// https://esbuild.github.io/api/#format
		// three possible values: iife, cjs, and esm
		Format: api.FormatCommonJS,

		// the import will be preserved and will be evaluated at run time instead
		// https://esbuild.github.io/api/#external
		External: external,

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

		// When `false` a buffer is returned instead.
		// The default for the snapshot version is `false`
		// https://esbuild.github.io/api/#write
		Write: args.Write,

		Snapshot: &api.SnapshotOptions{
			CreateSnapshot:       true,
			ShouldReplaceRequire: snap_api.CreateShouldReplaceRequire(platform, external, shouldReplaceRequire, shouldRewriteModule),
			ShouldRewriteModule:  shouldRewriteModule,
			AbsBasedir:           args.Basedir,
			Doctor:               args.Doctor,
			VerifyPrint:          true,
			PanicOnError:         false,
		},

		//
		// Unused
		//

		// only matters when the format setting is iife
		GlobalName: "",

		Sourcemap: sourcemap,

		// Only works with ESM modules
		// https://esbuild.github.io/api/#splitting
		Splitting: false,

		MinifyWhitespace:  false,
		MinifyIdentifiers: false,
		MinifySyntax:      false,

		JSXFactory:  "",
		JSXFragment: "",

		// https://esbuild.github.io/api/#charset
		Charset: 0,

		// https://esbuild.github.io/api/#color
		Color: 0,

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
}
