package bundler

import (
	"fmt"
	"github.com/evanw/esbuild/internal/renamer"
)

func fileInfoJSON(f *file, r renamer.Renamer) string {
	replacementFunc := r.NameForSymbol(f.repr.(*reprJS).ast.WrapperRef)

	return fmt.Sprintf(`{
             "identifierName": "%s",
             "fullPath": "%s",
             "isEntryPoint": %t,
             "replacementFunction": "%s"
           }`,
		f.source.IdentifierName,
		f.source.KeyPath.Text,
		f.isEntryPoint,
		replacementFunc)
}
