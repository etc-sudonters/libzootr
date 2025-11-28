package boot

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"sudonters/libzootr/components"
	"sudonters/libzootr/internal"
	"sudonters/libzootr/magicbean/tracking"
	"sudonters/libzootr/mido/optimizer"
	"sudonters/libzootr/zecs"

	"github.com/etc-sudonters/substrate/slipup"
)

var namef = components.NameF

type name = components.Name

type FilePath = string
type DirPath = string

type LoadPaths struct {
	Spoiler, Tokens, Placements, Scripts FilePath
	Relations                            DirPath
}

func (this LoadPaths) readscripts() (map[string]string, error) {
	return internal.ReadJsonFileStringMap(string(this.Scripts))
}

func (this LoadPaths) readtokens() ([]token, error) {
	return internal.ReadJsonFileAs[[]token](string(this.Tokens))
}

func (this LoadPaths) readplacements() ([]placement, error) {
	return internal.ReadJsonFileAs[[]placement](string(this.Placements))
}

func readrelations(path string) ([]relations, error) {
	return internal.ReadJsonFileAs[[]relations](path)
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

		if strings.Contains(path, "mq") {
			return nil
		}

		relations, relationsErr := readrelations(path)
		if relationsErr != nil {
			slipup.Describef(err, "while reading %s", path)
		}
		for _, relations := range relations {

			if storeErr := store(relations); storeErr != nil {
				return slipup.Describef(storeErr, "while storing relationships from %q", path)
			}
		}
		return nil
	})
}

func storeScripts(ocm *zecs.Ocm, paths LoadPaths) error {
	eng := ocm.Engine()
	scripts, readErr := paths.readscripts()
	if readErr != nil {
		return readErr
	}
	for decl, source := range scripts {
		_, err := eng.InsertRow(components.ScriptDecl(decl), components.ScriptSource(source), name(optimizer.FastScriptNameFromDecl(decl)))
		if err != nil {
			return slipup.Describef(err, "while storing script %q", decl)
		}
	}
	return nil
}

func storeTokens(tokens tracking.Tokens, paths LoadPaths) error {
	fileTokens, readErr := paths.readtokens()
	if readErr != nil {
		return readErr
	}
	for _, raw := range fileTokens {
		var attachments zecs.Attaching
		token := tokens.Named(name(raw.Name))

		if raw.Advancement {
			attachments.Add(components.PriorityAdvancement)
		} else if raw.Priority {
			attachments.Add(components.PriorityMajor)
		} else if raw.Special != nil {
			if _, exists := raw.Special["junk"]; exists {
				attachments.Add(components.PriorityJunk)
			}
		}

		switch raw.Type {
		case "BossKey", "bosskey":
			attachments.Add(components.BossKey{}, components.ParseDungeonGroup(raw.Name))
			break
		case "Compass", "compass":
			attachments.Add(components.Compass{}, components.ParseDungeonGroup(raw.Name))
			break
		case "Drop", "drop":
			attachments.Add(components.Drop{})
			break
		case "DungeonReward", "dungeonreward":
			attachments.Add(components.DungeonReward{})
			break
		case "Event", "event":
			attachments.Add(components.Event{})
			break
		case "GanonBossKey", "ganonbosskey":
			attachments.Add(components.BossKey{}, components.DUNGEON_GANON_CASTLE)
			break
		case "Item", "item":
			attachments.Add(components.Item{})
			break
		case "Map", "map":
			attachments.Add(components.Map{}, components.ParseDungeonGroup(raw.Name))
			break
		case "Refill", "refill":
			attachments.Add(components.Refill{})
			break
		case "Shop", "shop":
			attachments.Add(components.Shop{})
			break
		case "SilverRupee", "silverrupee":
			attachments.Add(components.ParseSilverRupeePuzzle(raw.Name))

			if strings.Contains(raw.Name, "Pouch") {
				attachments.Add(components.SilverRupeePouch{})
			} else {
				attachments.Add(components.SilverRupee{})
			}
			break
		case "SmallKey", "smallkey",
			"HideoutSmallKey", "hideoutsmallkey",
			"TCGSmallKey", "tcgsmallkey":
			attachments.Add(components.SmallKey{}, components.ParseDungeonGroup(raw.Name))
			break
		case "SmallKeyRing", "smallkeyring",
			"HideoutSmallKeyRing", "hideoutsmallkeyring",
			"TCGSmallKeyRing", "tcgsmallkeyring":
			attachments.Add(components.DungeonKeyRing{}, components.ParseDungeonGroup(raw.Name))
			break
		case "Song", "song":
			switch raw.Name {
			case "Prelude of Light":
				attachments.Add(components.SONG_PRELUDE, components.SongNotes("^>^><^"))
			case "Bolero of Fire":
				attachments.Add(components.SONG_BOLERO, components.SongNotes("vAvA>v>v"))
			case "Minuet of Forest":
				attachments.Add(components.SONG_MINUET, components.SongNotes("A^<><>"))
			case "Serenade of Water":
				attachments.Add(components.SONG_SERENADE, components.SongNotes("Av>><"))
			case "Requiem of Spirit":
				attachments.Add(components.SONG_REQUIEM, components.SongNotes("AvA>vA"))
			case "Nocturne of Shadow":
				attachments.Add(components.SONG_NOCTURNE, components.SongNotes("<>>A<>v"))
			case "Sarias Song":
				attachments.Add(components.SONG_SARIA, components.SongNotes("v><v><"))
			case "Eponas Song":
				attachments.Add(components.SONG_EPONA, components.SongNotes("^<>^<>"))
			case "Zeldas Lullaby":
				attachments.Add(components.SONG_LULLABY, components.SongNotes("<^><^>"))
			case "Suns Song":
				attachments.Add(components.SONG_SUN, components.SongNotes(">v^>v^"))
			case "Song of Time":
				attachments.Add(components.SONG_TIME, components.SongNotes(">Av>Av"))
			case "Song of Storms":
				attachments.Add(components.SONG_STORMS, components.SongNotes("Av^Av^"))
			default:
				panic(fmt.Errorf("unknown song %q", raw.Name))
			}
			break
		case "GoldSkulltulaToken", "goldskulltulatoken":
			attachments.Add(components.GoldSkulltulaToken{})
			break
		}

		if raw.Special != nil {
			for name, special := range raw.Special {
				// TODO turn this into more components
				_, _ = name, special
			}
		}

		if err := (token.AttachFrom(attachments)); err != nil { // attachment issues
			return slipup.Describef(err, "while attaching components to %q", raw.Name)
		}
	}
	return nil
}

func storePlacements(nodes tracking.Nodes, tokens tracking.Tokens, paths LoadPaths) error {
	placements, readErr := paths.readplacements()
	if readErr != nil {
		return readErr
	}
	for _, raw := range placements {
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
			transit.Proxy.Attach(components.RuleSource(rule), components.EdgeTransit)
		}

		for location, rule := range raw.Locations {
			placename := namef("%s", location)
			placement := nodes.Placement(placename)
			edge := region.Has(placement)
			edge.Attach(components.RuleSource(rule))
		}

		for event, rule := range raw.Events {
			token := tokens.Named(name(event))
			token.Attach(components.Event{})
			placement := nodes.Placement(namef("%s %s", raw.RegionName, event))
			placement.Fixed(token)
			edge := region.Has(placement)
			edge.Attach(components.RuleSource(rule))
		}

		var attachments zecs.Attaching

		if raw.RegionName == "Root" {
			attachments.Add(components.WorldGraphRoot{})
		}

		if raw.Hint != "" {
			attachments.Add(components.HintRegion(raw.Hint))
		}

		if raw.AltHint != "" {
			attachments.Add(components.AltHintRegion(raw.AltHint))
		}

		if raw.Dungeon != "" {
			attachments.Add(components.DungeonName(raw.Dungeon))
		}

		if raw.IsBossRoom {
			attachments.Add(components.IsBossRoom{})
		}

		if raw.Savewarp != "" {
			attachments.Add(components.Savewarp(raw.Savewarp))
		}

		if raw.Scene != "" {
			attachments.Add(components.Scene(raw.Scene))
		}

		if raw.TimePasses {
			attachments.Add(components.TimePassess{})
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

var smallKeyGroup = regexp.MustCompile(`Small Key( Ring)? \((.*)\)`)
var silverRupeeGroup = regexp.MustCompile(`Silver Rupee( Pouch)? \((.*)\)`)
