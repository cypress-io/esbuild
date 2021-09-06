package snap_api

import (
	"strings"

	"github.com/evanw/esbuild/internal/resolver"
	"github.com/evanw/esbuild/pkg/api"
)

func IsExternalModule(platform api.Platform, external []string) api.ShouldReplaceRequirePredicate {
	return func(mdl string) bool {
		if platform == api.PlatformNode {
			if _, ok := resolver.BuiltInNodeModules[mdl]; ok {
				return true
			}
		}
		for _, ext := range external {
			if ext == mdl {
				return true
			}
		}
		return false
	}
}

func IsNative(mdl string) bool {
	return strings.HasSuffix(mdl, ".node")
}

func CreateShouldReplaceRequire(
	platform api.Platform,
	external []string,
	replaceRequire api.ShouldReplaceRequirePredicate,
	rewriteModule api.ShouldRewriteModulePredicate,
) api.ShouldReplaceRequirePredicate {
	isExternal := IsExternalModule(platform, external)
	return func(mdl string) bool {
		// NOTE: normalizing cache/require keys to always use forward slashes
		// We could already store them as such which would be more efficient, but after measuring
		// I found zero perf overhead on OSX. Nothing gets replaced in this case so the only affected
		// OS would be windows, but should be neglible there as well.
		rewriteMdl := strings.ReplaceAll(mdl, "\\", "/")
		return isExternal(mdl) || IsNative(mdl) || replaceRequire(mdl) || !rewriteModule(rewriteMdl)
	}
}
