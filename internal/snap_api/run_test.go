package snap_api

import (
	"testing"

	"github.com/evanw/esbuild/pkg/api"
)

var entryPoints =  []string{"../../examples/express-app/snap.js"}
// var entryPoints =  []string{"input.js"}

func TestRunJS(t *testing.T) {
	result := api.Build(api.BuildOptions{
		EntryPoints: entryPoints,
		Outfile:     "output_js.js",
		Platform:    api.PlatformNode,
		Bundle:      true,
		Write:       true,
		LogLevel:    api.LogLevelInfo,
	})

	if len(result.Errors) > 0 {
		t.FailNow()
	}
}

func TestRunSnap(t *testing.T) {
	result := api.Build(api.BuildOptions{
		EntryPoints: entryPoints,
		Outfile:     "output_snap.js",
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
