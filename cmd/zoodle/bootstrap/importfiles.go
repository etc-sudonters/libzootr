package bootstrap

import (
	"context"
	"fmt"
	"io/fs"
	"iter"
	"log/slog"
	"path/filepath"
	"strings"
	"sudonters/libzootr/importers"
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/magicbean/tracking"
	"sudonters/libzootr/mido/optimizer"
	"sudonters/libzootr/zecs"

	"github.com/etc-sudonters/substrate/slipup"
)

var namef = magicbean.NameF

type name = magicbean.Name

type FilePath = string
type DirPath = string

type LoadPaths struct {
	Tokens, Placements, Scripts FilePath
	Relations                   DirPath
}

func (this LoadPaths) readscripts(ctx context.Context, fs fs.FS) iter.Seq2[importers.DumpedScript, error] {
	return func(yield func(importers.DumpedScript, error) bool) {
		fh, fhErr := fs.Open(this.Scripts)
		defer func() {
			if fh != nil {
				fh.Close()
			}
		}()

		if fhErr != nil {
			yield(importers.DumpedScript{}, fhErr)
			return
		}

		for script, err := range importers.DumpScripts.ImportFrom(ctx, fh) {
			if !yield(script, err) || err != nil {
				return
			}
		}
	}
}

func (this LoadPaths) readtokens(ctx context.Context, fs fs.FS) iter.Seq2[importers.DumpedItem, error] {
	return func(yield func(importers.DumpedItem, error) bool) {
		fh, fhErr := fs.Open(string(this.Tokens))
		defer func() {
			if fh != nil {
				fh.Close()
			}
		}()
		if fhErr != nil {
			yield(importers.DumpedItem{}, fhErr)
			return
		}

		for item, err := range importers.DumpItems.ImportFrom(ctx, fh) {
			if !yield(item, err) || err != nil {
				return
			}
		}
	}
}

func (this LoadPaths) readplacements(ctx context.Context, fs fs.FS) iter.Seq2[importers.DumpedLocation, error] {
	return func(yield func(importers.DumpedLocation, error) bool) {
		fh, fhErr := fs.Open(string(this.Placements))
		defer func() {
			if fh != nil {
				fh.Close()
			}
		}()
		if fhErr != nil {
			yield(importers.DumpedLocation{}, fhErr)
			return
		}

		for location, err := range importers.DumpLocations.ImportFrom(ctx, fh) {
			if !yield(location, err) || err != nil {
				return
			}
		}
	}
}

func (this LoadPaths) readrelationsdir(ctx context.Context, fsys fs.FS, store func(importers.DumpedRelation) error) error {
	return filepath.WalkDir(string(this.Relations), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return slipup.Describe(err, "logic directory walk called with err")
		}

		info, err := entry.Info()
		if err != nil || info.Mode() != (^fs.ModeType)&info.Mode() {
			// either we couldn't get the info, which doesn't bode well
			// or it's some kind of not file thing which we also don't want
			return nil
		}

		if ext := filepath.Ext(path); ext != ".json" {
			return nil
		}

		slog.WarnContext(ctx, "loading logic file", "path", path)
		fh, fhErr := fsys.Open(path)
		defer func() {
			if fh != nil {
				fh.Close()
			}
		}()
		if fhErr != nil {
			return fhErr
		}
		for relation, dumpErr := range importers.DumpRelations.ImportFrom(ctx, fh) {
			if dumpErr != nil {
				return dumpErr
			}
			if storeErr := store(relation); storeErr != nil {
				return storeErr
			}
		}

		return nil
	})
}

func storeScripts(ctx context.Context, fs fs.FS, ocm *zecs.Ocm, paths LoadPaths) error {
	eng := ocm.Engine()
	for script, err := range paths.readscripts(ctx, fs) {
		if err != nil {
			return err
		}
		eng.InsertRow(
			magicbean.ScriptDecl(script.Decl),
			magicbean.ScriptSource(script.Src),
			name(optimizer.FastScriptNameFromDecl(script.Decl)),
		)
	}
	return nil
}

func storeTokens(ctx context.Context, fs fs.FS, tokens tracking.Tokens, paths LoadPaths) error {
	for item, decodeErr := range paths.readtokens(ctx, fs) {
		if decodeErr != nil {
			return decodeErr
		}
		var attachments zecs.Attaching
		token := tokens.Named(name(item.Name))

		if item.Advancement {
			attachments.Add(magicbean.PriorityAdvancement)
		} else if item.Priority {
			attachments.Add(magicbean.PriorityMajor)
		} else if item.Special != nil {
			if _, exists := item.Special["junk"]; exists {
				attachments.Add(magicbean.PriorityJunk)
			}
		}

		switch item.Type {
		case "BossKey", "bosskey":
			attachments.Add(magicbean.BossKey{}, magicbean.ParseDungeonGroup(item.Name))
		case "Compass", "compass":
			attachments.Add(magicbean.Compass{}, magicbean.ParseDungeonGroup(item.Name))
		case "Drop", "drop":
			attachments.Add(magicbean.Drop{})
		case "DungeonReward", "dungeonreward":
			attachments.Add(magicbean.DungeonReward{})
		case "Event", "event":
			attachments.Add(magicbean.Event{})
		case "GanonBossKey", "ganonbosskey":
			attachments.Add(magicbean.BossKey{}, magicbean.DUNGEON_GANON_CASTLE)
		case "Item", "item":
			attachments.Add(magicbean.Item{})
		case "Map", "map":
			attachments.Add(magicbean.Map{}, magicbean.ParseDungeonGroup(item.Name))
		case "Refill", "refill":
			attachments.Add(magicbean.Refill{})
		case "Shop", "shop":
			attachments.Add(magicbean.Shop{})
		case "SilverRupee", "silverrupee":
			attachments.Add(magicbean.ParseSilverRupeePuzzle(item.Name))

			if strings.Contains(item.Name, "Pouch") {
				attachments.Add(magicbean.SilverRupeePouch{})
			} else {
				attachments.Add(magicbean.SilverRupee{})
			}
		case "SmallKey", "smallkey",
			"HideoutSmallKey", "hideoutsmallkey",
			"TCGSmallKey", "tcgsmallkey":
			attachments.Add(magicbean.SmallKey{}, magicbean.ParseDungeonGroup(item.Name))
		case "SmallKeyRing", "smallkeyring",
			"HideoutSmallKeyRing", "hideoutsmallkeyring",
			"TCGSmallKeyRing", "tcgsmallkeyring":
			attachments.Add(magicbean.DungeonKeyRing{}, magicbean.ParseDungeonGroup(item.Name))
		case "Song", "song":
			switch item.Name {
			case "Prelude of Light":
				attachments.Add(magicbean.SONG_PRELUDE, magicbean.SongNotes("^>^><^"))
			case "Bolero of Fire":
				attachments.Add(magicbean.SONG_BOLERO, magicbean.SongNotes("vAvA>v>v"))
			case "Minuet of Forest":
				attachments.Add(magicbean.SONG_MINUET, magicbean.SongNotes("A^<><>"))
			case "Serenade of Water":
				attachments.Add(magicbean.SONG_SERENADE, magicbean.SongNotes("Av>><"))
			case "Requiem of Spirit":
				attachments.Add(magicbean.SONG_REQUIEM, magicbean.SongNotes("AvA>vA"))
			case "Nocturne of Shadow":
				attachments.Add(magicbean.SONG_NOCTURNE, magicbean.SongNotes("<>>A<>v"))
			case "Sarias Song":
				attachments.Add(magicbean.SONG_SARIA, magicbean.SongNotes("v><v><"))
			case "Eponas Song":
				attachments.Add(magicbean.SONG_EPONA, magicbean.SongNotes("^<>^<>"))
			case "Zeldas Lullaby":
				attachments.Add(magicbean.SONG_LULLABY, magicbean.SongNotes("<^><^>"))
			case "Suns Song":
				attachments.Add(magicbean.SONG_SUN, magicbean.SongNotes(">v^>v^"))
			case "Song of Time":
				attachments.Add(magicbean.SONG_TIME, magicbean.SongNotes(">Av>Av"))
			case "Song of Storms":
				attachments.Add(magicbean.SONG_STORMS, magicbean.SongNotes("Av^Av^"))
			default:
				panic(fmt.Errorf("unknown song %q", item.Name))
			}
		case "GoldSkulltulaToken", "goldskulltulatoken":
			attachments.Add(magicbean.GoldSkulltulaToken{})
		}

		if item.Special != nil {
			for name, special := range item.Special {
				// TODO turn this into more components
				_, _ = name, special
			}
		}

		if err := token.AttachFrom(attachments); err != nil {
			return slipup.Describef(err, "failed to create token %s", item.Name)
		}
	}
	return nil
}

func storePlacements(ctx context.Context, fs fs.FS, nodes tracking.Nodes, tokens tracking.Tokens, paths LoadPaths) error {
	for location, decodeErr := range paths.readplacements(ctx, fs) {
		if decodeErr != nil {
			return decodeErr
		}
		place := nodes.Placement(name(location.Name))
		if location.Default != "" {
			place.DefaultToken(tokens.Named(name(location.Default)))
		}
	}

	return nil
}

func storeRelations(ctx context.Context, fs fs.FS, nodes tracking.Nodes, tokens tracking.Tokens, paths LoadPaths) error {
	return paths.readrelationsdir(ctx, fs, func(relation importers.DumpedRelation) error {
		region := nodes.Region(name(relation.RegionName))

		for exit, rule := range relation.Exits {
			transit := region.ConnectsTo(nodes.Region(name(exit)))
			transit.Proxy.Attach(magicbean.RuleSource(rule), magicbean.EdgeTransit)
		}

		for location, rule := range relation.Relations {
			placename := namef("%s %s", relation.RegionName, location)
			placement := nodes.Placement(placename)
			edge := region.Has(placement)
			edge.Attach(magicbean.RuleSource(rule))
		}

		for event, rule := range relation.Events {
			token := tokens.Named(name(event))
			token.Attach(magicbean.Event{})
			placement := nodes.Placement(namef("%s %s", relation.RegionName, event))
			placement.Fixed(token)
			edge := region.Has(placement)
			edge.Attach(magicbean.RuleSource(rule))
		}

		var attachments zecs.Attaching

		if relation.RegionName == "Root" {
			attachments.Add(magicbean.WorldGraphRoot{})
		}

		if relation.Hint != "" {
			attachments.Add(magicbean.HintRegion(relation.Hint))
		}

		if relation.AltHint != "" {
			attachments.Add(magicbean.AltHintRegion(relation.AltHint))
		}

		if relation.Dungeon != "" {
			attachments.Add(magicbean.DungeonName(relation.Dungeon))
		}

		if relation.IsBossRoom {
			attachments.Add(magicbean.IsBossRoom{})
		}

		if relation.Savewarp != "" {
			attachments.Add(magicbean.Savewarp(relation.Savewarp))
		}

		if relation.Scene != "" {
			attachments.Add(magicbean.Scene(relation.Scene))
		}

		if relation.TimePasses {
			attachments.Add(magicbean.TimePassess{})
		}

		return region.AttachFrom(attachments)
	})
}
