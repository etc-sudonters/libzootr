package bootstrap

import (
	"context"
	"fmt"
	"io/fs"
	"iter"
	"path/filepath"
	"strings"
	"sudonters/libzootr/importers"
	"sudonters/libzootr/internal"
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

func (this LoadPaths) readscripts() map[string]string {
	scripts, err := internal.ReadJsonFileStringMap(string(this.Scripts))
	PanicWhenErr(err)
	return scripts
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

func (this LoadPaths) readplacements() []placement {
	regions, err := internal.ReadJsonFileAs[[]placement](string(this.Placements))
	PanicWhenErr(err)
	return regions
}

func readrelations(path string) []relations {
	relations, err := internal.ReadJsonFileAs[[]relations](path)
	PanicWhenErr(err)
	return relations
}

func (this LoadPaths) readrelationsdir(store func(relations) error) error {
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

		for _, relations := range readrelations(path) {
			storeErr := store(relations)
			PanicWhenErr(storeErr)
		}
		return nil
	})
}

func storeScripts(ocm *zecs.Ocm, paths LoadPaths) error {
	eng := ocm.Engine()
	for decl, source := range paths.readscripts() {
		eng.InsertRow(magicbean.ScriptDecl(decl), magicbean.ScriptSource(source), name(optimizer.FastScriptNameFromDecl(decl)))
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
			break
		case "Compass", "compass":
			attachments.Add(magicbean.Compass{}, magicbean.ParseDungeonGroup(item.Name))
			break
		case "Drop", "drop":
			attachments.Add(magicbean.Drop{})
			break
		case "DungeonReward", "dungeonreward":
			attachments.Add(magicbean.DungeonReward{})
			break
		case "Event", "event":
			attachments.Add(magicbean.Event{})
			break
		case "GanonBossKey", "ganonbosskey":
			attachments.Add(magicbean.BossKey{}, magicbean.DUNGEON_GANON_CASTLE)
			break
		case "Item", "item":
			attachments.Add(magicbean.Item{})
			break
		case "Map", "map":
			attachments.Add(magicbean.Map{}, magicbean.ParseDungeonGroup(item.Name))
			break
		case "Refill", "refill":
			attachments.Add(magicbean.Refill{})
			break
		case "Shop", "shop":
			attachments.Add(magicbean.Shop{})
			break
		case "SilverRupee", "silverrupee":
			attachments.Add(magicbean.ParseSilverRupeePuzzle(item.Name))

			if strings.Contains(item.Name, "Pouch") {
				attachments.Add(magicbean.SilverRupeePouch{})
			} else {
				attachments.Add(magicbean.SilverRupee{})
			}
			break
		case "SmallKey", "smallkey",
			"HideoutSmallKey", "hideoutsmallkey",
			"TCGSmallKey", "tcgsmallkey":
			attachments.Add(magicbean.SmallKey{}, magicbean.ParseDungeonGroup(item.Name))
			break
		case "SmallKeyRing", "smallkeyring",
			"HideoutSmallKeyRing", "hideoutsmallkeyring",
			"TCGSmallKeyRing", "tcgsmallkeyring":
			attachments.Add(magicbean.DungeonKeyRing{}, magicbean.ParseDungeonGroup(item.Name))
			break
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
			break
		case "GoldSkulltulaToken", "goldskulltulatoken":
			attachments.Add(magicbean.GoldSkulltulaToken{})
			break
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

func storePlacements(nodes tracking.Nodes, tokens tracking.Tokens, paths LoadPaths) error {
	for _, raw := range paths.readplacements() {
		place := nodes.Placement(name(raw.Name))
		if raw.Default != "" {
			place.DefaultToken(tokens.Named(name(raw.Default)))
		}
	}

	return nil
}

func storeRelations(nodes tracking.Nodes, tokens tracking.Tokens, paths LoadPaths) error {
	return paths.readrelationsdir(func(raw relations) error {
		region := nodes.Region(name(raw.RegionName))

		for exit, rule := range raw.Exits {
			transit := region.ConnectsTo(nodes.Region(name(exit)))
			transit.Proxy.Attach(magicbean.RuleSource(rule), magicbean.EdgeTransit)
		}

		for location, rule := range raw.Locations {
			placename := namef("%s %s", raw.RegionName, location)
			placement := nodes.Placement(placename)
			edge := region.Has(placement)
			edge.Attach(magicbean.RuleSource(rule))
		}

		for event, rule := range raw.Events {
			token := tokens.Named(name(event))
			token.Attach(magicbean.Event{})
			placement := nodes.Placement(namef("%s %s", raw.RegionName, event))
			placement.Fixed(token)
			edge := region.Has(placement)
			edge.Attach(magicbean.RuleSource(rule))
		}

		var attachments zecs.Attaching

		if raw.RegionName == "Root" {
			attachments.Add(magicbean.WorldGraphRoot{})
		}

		if raw.Hint != "" {
			attachments.Add(magicbean.HintRegion(raw.Hint))
		}

		if raw.AltHint != "" {
			attachments.Add(magicbean.AltHintRegion(raw.AltHint))
		}

		if raw.Dungeon != "" {
			attachments.Add(magicbean.DungeonName(raw.Dungeon))
		}

		if raw.IsBossRoom {
			attachments.Add(magicbean.IsBossRoom{})
		}

		if raw.Savewarp != "" {
			attachments.Add(magicbean.Savewarp(raw.Savewarp))
		}

		if raw.Scene != "" {
			attachments.Add(magicbean.Scene(raw.Scene))
		}

		if raw.TimePasses {
			attachments.Add(magicbean.TimePassess{})
		}

		return region.AttachFrom(attachments)
	})
}

type placement struct {
	Categories []string `json:"categories"`
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Default    string   `json:"vanilla"`
}

type relations struct {
	Events     map[string]string `json:"events"`
	Exits      map[string]string `json:"exits"`
	Locations  map[string]string `json:"locations"`
	RegionName string            `json:"region_name"`
	AltHint    string            `json:"alt_hint"`
	Hint       string            `json:"hint"`
	Dungeon    string            `json:"dungeon"`
	IsBossRoom bool              `json:"is_boss_room"`
	Savewarp   string            `json:"savewarp"`
	Scene      string            `json:"scene"`
	TimePasses bool              `json:"time_passes"`
}

type token struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Advancement bool                   `json:"advancement"`
	Priority    bool                   `json:"priority"`
	Special     map[string]interface{} `json:"special"`
}
