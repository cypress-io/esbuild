package api

import (
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/evanw/esbuild/internal/config"
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
	source := logger.Source{
		Index:          0,
		KeyPath:        logger.Path{Text: filePath},
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

func reportError(log *logger.Log, filePath string, error string, shouldPanic bool) {
	loc := logger.Loc{Start: 0}
	path := logger.Path{Text: filePath, Namespace: "file"}
	source := logger.Source{
		Index:          0,
		KeyPath:        path,
		PrettyPath:     filePath,
		IdentifierName: filePath, Contents: ""}

	s := fmt.Sprintf("Encountered an error inside '%s'\n  %s", filePath, error)
	log.AddError(&source, loc, s)

	if shouldPanic {
		panic(s)
	}
}
