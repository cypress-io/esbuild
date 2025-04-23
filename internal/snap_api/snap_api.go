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
		return isExternal(mdl) || IsNative(mdl) || replaceRequire(mdl) || !rewriteModule(mdl)
	}
}

func trimPrefix(mdl string, prefix string) string {
	if strings.HasPrefix(mdl, prefix) {
		return mdl[len(prefix):]
	}
	return mdl
}

func CreateShouldRewriteModule(
	args *SnapCmdArgs,
) api.ShouldRewriteModulePredicate {
	return func(mdl string) bool {
		if len(mdl) == 0 {
			return true
		}
		mdl = trimPrefix(mdl, "./")

		if args.Norewrite != nil {
			for _, m := range args.Norewrite {
				// The force no rewrite file follows a convention where we try
				// and match all possible node_modules paths if the force no
				// rewrite entry starts with "*". If it does not
				// start with "*/" then it is an exact match.
				if strings.HasPrefix(m, "*") {
					m = trimPrefix(m, "*/")
					if strings.HasSuffix(mdl, m) {
						return false
					}
				} else if m == mdl {
					return false
				}
			}
		}
		return true
	}
}
