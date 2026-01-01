package importers

import (
	"io"
	"iter"
	"maps"
	"sudonters/libzootr/internal/json"

	"github.com/etc-sudonters/substrate/slipup"
)

type DumpedRelation struct {
	Events            map[string]string
	Exits             map[string]string
	Relations         map[string]string
	RegionName        string
	AltHint           string
	Hint              string
	Dungeon           string
	IsBossRoom        bool
	Savewarp          string
	Scene             string
	TimePasses        bool
	ProvidesTimeOfDay string
}

var emptyRelation = DumpedRelation{}

// imports OOTR logic relation files
type DumpedRelations struct{}

var DumpRelations = DumpedRelations{}

// yields relations from the reader until the reader is completely consumed or
// an error occurs. If an error occurs then an empty Relation is yielded
// and the error will not be nil. Iteration may be canceled from the provided
// context
//
//	func storeRelation(importer.Relation)
//
//	for relation, err := importers.DumpRelations.ImportFrom(ctx, relationReader) {
//	    if err != nil {
//	        return err
//	    }
//	    storeRelation(relation)
//	}
func (this *DumpedRelations) ImportFrom(ctx ctx, r io.Reader) iter.Seq2[DumpedRelation, error] {
	return func(yield func(DumpedRelation, error) bool) {
		parser := json.ParserFrom(r)
		relations, notRelationArray := parser.ReadArray()
		if notRelationArray != nil {
			yield(emptyRelation, notRelationArray)
			return
		}
		for relations.More() {
			select {
			case <-ctx.Done():
				yield(emptyRelation, ctx.Err())
				return
			default:
				obj, notRelationErr := relations.ReadObject()
				if notRelationErr != nil {
					yield(emptyRelation, notRelationErr)
					return
				}

				relation, dumpErr := dumpOneRelation(ctx, obj)
				if !yield(relation, dumpErr) || dumpErr != nil {
					return
				}
			}
		}
	}
}

func dumpOneRelation(ctx ctx, obj *json.ObjectParser) (DumpedRelation, error) {
	var relation DumpedRelation
	var property string

	defer func() {
		if recovered := recover(); recovered != nil {
			var panicwith error
			switch recovered := recovered.(type) {
			case error:
				panicwith = slipup.Describef(recovered, "while handling relation %#v last property %q", relation, property)
			default:
				panicwith = slipup.Createf("while handling relation: %#v last property %q: %v", relation, property, recovered)
			}
			panic(panicwith)
		}
	}()

	for obj.More() {
		select {
		case <-ctx.Done():
			return relation, ctx.Err()
		default:
			var propertyErr error
			property, propertyErr = obj.ReadPropertyName()
			if propertyErr != nil {
				return relation, propertyErr
			}

			var readErr error
			switch property {
			case "events":
				relation.Events = maps.Collect(json.ReadNullableStringObject(obj, &readErr))
			case "exits":
				relation.Exits = maps.Collect(json.ReadNullableStringObject(obj, &readErr))
			case "locations":
				relation.Relations = maps.Collect(json.ReadNullableStringObject(obj, &readErr))
			case "region_name":
				relation.RegionName, readErr = json.ReadNullableString(obj)
			case "alt_hint":
				relation.AltHint, readErr = json.ReadNullableString(obj)
			case "hint":
				relation.Hint, readErr = json.ReadNullableString(obj)
			case "dungeon":
				relation.Dungeon, readErr = json.ReadNullableString(obj)
			case "is_boss_room":
				relation.IsBossRoom, readErr = json.ReadNullableBool(obj)
			case "savewarp":
				relation.Savewarp, readErr = json.ReadNullableString(obj)
			case "scene":
				relation.Scene, readErr = json.ReadNullableString(obj)
			case "time_passes":
				relation.TimePasses, readErr = json.ReadNullableBool(obj)
			case "provides_time":
				relation.ProvidesTimeOfDay, readErr = json.ReadNullableString(obj)
			default:
				readErr = slipup.Createf("unknown property %s", property)
			}

			if readErr != nil {
				return relation, slipup.Describef(
					readErr, "failed while reading relation %#v", relation,
				)
			}
		}
	}

	return relation, obj.ReadEnd()
}
