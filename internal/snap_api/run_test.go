package snap_api

import (
	"testing"

	"github.com/evanw/esbuild/pkg/api"
)

func TestRun(t *testing.T) {
	result := api.Build(api.BuildOptions{
		EntryPoints: []string{"input.js"},
		Outfile:     "output.js",
		Platform:    api.PlatformNode,
		Bundle:      true,
		Write:       true,
		LogLevel:    api.LogLevelInfo,
		Snapshot:    true,
	})

	if len(result.Errors) > 0 {
		t.FailNow()
	}
}
