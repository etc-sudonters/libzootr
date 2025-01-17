package bootstrap

import (
	"io/fs"
	"path/filepath"
	"sudonters/zootler/cmd/magicbean/z16"
	"sudonters/zootler/internal"
	"sudonters/zootler/magicbean"
	"sudonters/zootler/mido/ast"
	"sudonters/zootler/zecs"

	"github.com/etc-sudonters/substrate/slipup"
)

var namef = magicbean.NameF

type name = magicbean.Name

type FilePath string
type DirPath string

type LoadPaths struct {
	Tokens, Placements, Scripts FilePath
	Relations                   DirPath
}

func (this LoadPaths) readscripts() map[string]string {
	scripts, err := internal.ReadJsonFileStringMap(string(this.Scripts))
	panicWhenErr(err)
	return scripts
}

func (this LoadPaths) readtokens() []token {
	tokens, err := internal.ReadJsonFileAs[[]token](string(this.Tokens))
	panicWhenErr(err)
	return tokens
}

func (this LoadPaths) readplacements() []placement {
	regions, err := internal.ReadJsonFileAs[[]placement](string(this.Placements))
	panicWhenErr(err)
	return regions
}

func readrelations(path string) []relations {
	relations, err := internal.ReadJsonFileAs[[]relations](path)
	panicWhenErr(err)
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
			panicWhenErr(storeErr)
		}
		return nil
	})
}

func storeScripts(ocm *zecs.Ocm, paths LoadPaths) error {
	eng := ocm.Engine()
	for decl, source := range paths.readscripts() {
		eng.InsertRow(magicbean.ScriptDecl(decl), magicbean.ScriptSource(source), name(ast.FastScriptNameFromDecl(decl)))
	}
	return nil
}

func storeTokens(tokens z16.Tokens, paths LoadPaths) error {
	for _, raw := range paths.readtokens() {
		var attachments zecs.Attaching
		token := tokens.Named(name(raw.Name))

		if raw.Advancement {
			attachments.Add(magicbean.PriorityAdvancement)
		} else if raw.Priority {
			attachments.Add(magicbean.PriorityMajor)
		} else if raw.Special != nil {
			if _, exists := raw.Special["junk"]; exists {
				attachments.Add(magicbean.PriorityJunk)
			}
		}

		switch raw.Type {
		case "BossKey", "bosskey":
			attachments.Add(magicbean.BossKey{})
			break
		case "Compass", "compass":
			attachments.Add(magicbean.Compass{})
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
			attachments.Add(magicbean.GanonBossKey{})
			break
		case "HideoutSmallKey", "hideoutsmallkey":
			attachments.Add(magicbean.HideoutSmallKey{})
			break
		case "HideoutSmallKeyRing", "hideoutsmallkeyring":
			attachments.Add(magicbean.HideoutSmallKeyRing{})
			break
		case "Item", "item":
			attachments.Add(magicbean.Item{})
			break
		case "Map", "map":
			attachments.Add(magicbean.Map{})
			break
		case "Refill", "refill":
			attachments.Add(magicbean.Refill{})
			break
		case "Shop", "shop":
			attachments.Add(magicbean.Shop{})
			break
		case "SilverRupee", "silverrupee":
			attachments.Add(magicbean.SilverRupee{})
			break
		case "SmallKey", "smallkey":
			attachments.Add(magicbean.SmallKey{})
			break
		case "SmallKeyRing", "smallkeyring":
			attachments.Add(magicbean.SmallKeyRing{})
			break
		case "Song", "song":
			attachments.Add(magicbean.Song{})
			break
		case "TCGSmallKey", "tcgsmallkey":
			attachments.Add(magicbean.TCGSmallKey{})
			break
		case "TCGSmallKeyRing", "tcgsmallkeyring":
			attachments.Add(magicbean.TCGSmallKeyRing{})
			break
		case "GoldSkulltulaToken", "goldskulltulatoken":
			attachments.Add(magicbean.GoldSkulltulaToken{})
			break
		}

		if raw.Special != nil {
			for name, special := range raw.Special {
				// TODO turn this into more components
				_, _ = name, special
			}
		}

		panicWhenErr(token.AttachFrom(attachments))
	}
	return nil
}

func storeplacements(nodes z16.Nodes, tokens z16.Tokens, paths LoadPaths) error {
	for _, raw := range paths.readplacements() {
		place := nodes.Placement(name(raw.Name))
		if raw.Default != "" {
			place.DefaultToken(tokens.Named(name(raw.Default)))
		}
	}

	return nil
}

func storeRelations(nodes z16.Nodes, tokens z16.Tokens, paths LoadPaths) error {
	return paths.readrelationsdir(func(raw relations) error {
		region := nodes.Region(name(raw.RegionName))

		for exit, rule := range raw.Exits {
			transit := region.Connects(nodes.Region(name(exit)))
			transit.Edge.Attach(magicbean.RuleSource(rule), magicbean.EdgeTransit)
		}

		for location, rule := range raw.Locations {
			placement := nodes.Placement(name(location))
			edge := region.Has(placement)
			edge.Attach(magicbean.RuleSource(rule))
		}

		for event, rule := range raw.Events {
			token := tokens.Named(name(event))
			token.Attach(magicbean.Event{})
			placement := nodes.Placement(namef("%s %s", raw.RegionName, event))
			placement.Owns(token)
			edge := region.Has(placement)
			edge.Attach(magicbean.RuleSource(rule))
		}

		var attachments zecs.Attaching

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

		return nil
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
