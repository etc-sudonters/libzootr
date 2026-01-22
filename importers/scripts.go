package importers

import (
	"io"
	"iter"
	"sudonters/libzootr/internal/json"

	"github.com/etc-sudonters/substrate/slipup"
)

type DumpedScript struct {
	Decl, Src string
}

// imports OOTR logic script files
type DumpedScripts struct{}

var DumpScripts = DumpedScripts{}

// yields scripts from the reader until the reader is completely consumed or
// an error occurs. If an error occurs then an empty Script is yielded
// and the error will not be nil. Iteration may be canceled from the provided
// context
//
//	func storeScript(importer.Script)
//
//	for script, err := importers.DumpScripts.ImportFrom(ctx, scriptReader) {
//	    if err != nil {
//	        return err
//	    }
//	    storeScript(script)
//	}
func (this *DumpedScripts) ImportFrom(ctx ctx, r io.Reader) iter.Seq2[DumpedScript, error] {
	return func(yield func(DumpedScript, error) bool) {
		var emptyScript DumpedScript
		parser := json.ParserFrom(r)
		scripts, notScriptObject := parser.ReadObject()
		if notScriptObject != nil {
			yield(emptyScript, notScriptObject)
			return
		}
		for scripts.More() {
			select {
			case <-ctx.Done():
				yield(emptyScript, ctx.Err())
				return
			default:
				decl, declErr := scripts.ReadPropertyName()
				if declErr != nil {
					yield(emptyScript, declErr)
					return
				}
				src, srcErr := scripts.ReadString()
				if srcErr != nil {
					yield(emptyScript, slipup.Describef(srcErr, "while loading script %q", decl))
					return
				}

				if !yield(DumpedScript{decl, src}, nil) {
					return
				}
			}
		}
	}
}
