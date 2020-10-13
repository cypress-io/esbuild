package snap_printer

import (
	"fmt"
	"github.com/evanw/esbuild/internal/config"
	"github.com/evanw/esbuild/internal/js_ast"
	"github.com/evanw/esbuild/internal/js_parser"
	"github.com/evanw/esbuild/internal/logger"
	"github.com/evanw/esbuild/internal/snap_renamer"
	"github.com/evanw/esbuild/internal/test"
	"strings"
	"testing"
)

func RunOnly(
	contents string,
) {
	options := PrintOptions{}
	log := logger.NewDeferLog()

	tree, ok := js_parser.Parse(log, test.SourceForTest(contents), config.Options{
		UnsupportedJSFeatures: options.UnsupportedFeatures,
	})
	msgs := log.Done()
	text := ""
	for _, msg := range msgs {
		text += msg.String(logger.StderrOptions{}, logger.TerminalInfo{})
	}
	if len(text) > 0 {
		fmt.Printf(text)
		panic("Parse error")
	}
	if !ok {
		panic("Parse error")
	}
	symbols := js_ast.NewSymbolMap(1)
	symbols.Outer[0] = tree.Symbols
	r := snap_renamer.NewSnapRenamer(symbols)
	var js []byte
	js = Print(tree, symbols, r, options).JS
	fmt.Printf(strings.TrimSpace(string(js)))
}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	t.Helper()
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}

func expectPrintedCommon(
	t *testing.T,
	name string,
	contents string,
	expected string,
	options PrintOptions,
	shouldReplaceRequire func(string) bool,
) {
	// TODO: pass + use shouldReplaceRequire
	t.Helper()
	t.Run(name, func(t *testing.T) {
		t.Helper()
		log := logger.NewDeferLog()
		tree, ok := js_parser.Parse(log, test.SourceForTest(contents), config.Options{
			UnsupportedJSFeatures: options.UnsupportedFeatures,
		})
		msgs := log.Done()
		text := ""
		for _, msg := range msgs {
			text += msg.String(logger.StderrOptions{}, logger.TerminalInfo{})
		}
		assertEqual(t, text, "")
		if !ok {
			t.Fatal("Parse error")
		}
		symbols := js_ast.NewSymbolMap(1)
		symbols.Outer[0] = tree.Symbols
		r := snap_renamer.NewSnapRenamer(symbols)
		js := Print(tree, symbols, r, options).JS
		assertEqual(t, strings.TrimSpace(string(js)), strings.TrimSpace(expected))
	})
}

func expectPrinted(t *testing.T, contents string, expected string, shouldReplaceRequire func(string) bool) {
	t.Helper()
	expectPrintedCommon(
		t,
		contents,
		contents,
		expected,
		PrintOptions{},
		shouldReplaceRequire)
}

func ReplaceAll(string) bool { return true }
