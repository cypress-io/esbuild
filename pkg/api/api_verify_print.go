package api

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/evanw/esbuild/internal/config"
	"github.com/evanw/esbuild/internal/js_ast"
	"github.com/evanw/esbuild/internal/js_parser"
	"github.com/evanw/esbuild/internal/logger"
	"github.com/evanw/esbuild/internal/snap_printer"
)

func VerifyLog() (logger.Log, *logger.SortableMsgs) {
	var msgs logger.SortableMsgs
	var mutex sync.Mutex
	var hasErrors bool

	log := logger.Log{
		AddMsg: func(msg logger.Msg) {
			mutex.Lock()
			defer mutex.Unlock()
			if msg.Kind == logger.Error {
				hasErrors = true
			}
			msgs = append(msgs, msg)
		},
		HasErrors: func() bool {
			mutex.Lock()
			defer mutex.Unlock()
			return hasErrors
		},
		Done: func() []logger.Msg {
			mutex.Lock()
			defer mutex.Unlock()
			sort.Stable(msgs)
			return msgs
		},
	}
	return log, &msgs
}

func verifyPrint(result *snap_printer.PrintResult, filePath string, shouldPanic bool) {
	log, msgs := VerifyLog()
	path := logger.Path{Text: filePath, Namespace: "file"}
	source := logger.Source{
		Index:          0,
		KeyPath:        path,
		PrettyPath:     filePath,
		Contents:       string(result.JS),
		IdentifierName: filePath,
	}
	js_parser.Parse(log, source, js_parser.OptionsFromConfig(&config.Options{}))
	if !log.HasErrors() {
		return
	}
	s := "\nVerification failed!"
	for _, msg := range *msgs {
		loc := msg.Data.Location
		s += fmt.Sprintf("\n----------------------------\n")
		s += fmt.Sprintf("%s\n\n", msg.Data.Text)
		s += fmt.Sprintf("at %s:%d:%d\n", loc.File, loc.Line, loc.Column)
		s += fmt.Sprintf("%s", loc.LineText)
	}
	if shouldPanic {
		panic(s)
	} else {
		fmt.Fprintln(os.Stderr, s)
	}
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
