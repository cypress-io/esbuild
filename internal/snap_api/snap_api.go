package snap_api

import (
	"github.com/evanw/esbuild/internal/resolver"
	"github.com/evanw/esbuild/pkg/api"
)

func IsExternalModule(platform api.Platform, external []string) func(mdl string) bool {
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
