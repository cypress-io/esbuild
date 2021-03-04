package api

import (
	"fmt"
	"os"
	"strings"

	"github.com/evanw/esbuild/internal/config"
	"github.com/evanw/esbuild/internal/js_ast"
	"github.com/evanw/esbuild/internal/js_parser"
	"github.com/evanw/esbuild/internal/logger"
	"github.com/evanw/esbuild/internal/snap_printer"
)

func ErrorToWarningLog(log *logger.Log) logger.Log {
	forgivingLog := logger.Log{
		AddMsg: func(msg logger.Msg) {
			if msg.Kind == logger.Error {
				msg.Data.Text = fmt.Sprintf("[SNAPSHOT_REWRITE_FAILURE] %s", msg.Data.Text)
				msg.Kind = logger.Warning
			}
			log.AddMsg(msg)
		},
		HasErrors: func() bool {
			return log.HasErrors()
		},
		Done: func() []logger.Msg {
			return log.Done()
		},
	}
	return forgivingLog
}

func verifyPrint(result *snap_printer.PrintResult, log *logger.Log, filePath string, shouldPanic bool) {
	// Cannot use printer logger since that would add any issues as error messages which causes the
	// entire process to fail. What we want instead is to provide an indicator of what error
	// occurred in which file so that the caller can process it.
	vlog := ErrorToWarningLog(log)
	path := logger.Path{Text: filePath, Namespace: "file"}
	source := logger.Source{
		Index:          0,
		KeyPath:        path,
		PrettyPath:     filePath,
		Contents:       string(result.JS),
		IdentifierName: filePath,
	}
	js_parser.Parse(vlog, source, js_parser.OptionsFromConfig(&config.Options{}))
}

func reportWarning(
	result *snap_printer.PrintResult,
	log *logger.Log,
	filePath string,
	error string,
	errorStart int32,
	shouldPanic bool) {
	loc := logger.Loc{Start: errorStart}
	path := logger.Path{Text: filePath, Namespace: "file"}
	source := logger.Source{
		Index:          0,
		KeyPath:        path,
		PrettyPath:     filePath,
		Contents:       string(result.JS),
		IdentifierName: filePath,
	}

	s := fmt.Sprintf("Encountered a problem inside '%s'\n  %s", filePath, error)
	log.AddWarning(&source, loc, s)

	if shouldPanic {
		panic(s)
	} else {
		fmt.Fprintln(os.Stderr, s)
	}
}

// Tries to find the needle in the code and normalizes the result to `0` if not found
func tryFindLocInside(js *[]byte, needle string, skip int) int32 {
	// Here we do a cheap search in the code to guess where the use of the needle occurred
	loc := 0
	needleLen := len(needle)
	offset := 0
	s := string(*js)
	for n := 0; n <= skip; n++ {
		loc := strings.Index(s[offset:], needle)
		if loc < 0 {
			return 0
		}
		offset = offset + loc + needleLen
	}
	return int32(offset + loc - needleLen)
}

func RejectDirnameAccess(tree *js_ast.AST, js *[]byte) (string, int32, bool) {
	if tree.UsesDirnameRef {
		loc := tryFindLocInside(js, "__dirname", 1)
		return "Forbidden use of __dirname", loc, true
	}
	return "", 0, false
}

func RejectFilenameAccess(tree *js_ast.AST, js *[]byte) (string, int32, bool) {
	if tree.UsesFilenameRef {
		loc := tryFindLocInside(js, "__filename", 1)
		return "Forbidden use of __filename", loc, true
	}
	return "", 0, false
}
