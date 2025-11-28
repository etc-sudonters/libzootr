package components

import (
	"fmt"
	"regexp"
	"sudonters/libzootr/mido/ast"
	"sudonters/libzootr/mido/compiler"
	"sudonters/libzootr/mido/objects"
	"sudonters/libzootr/zecs"
)

type region = zecs.Entity
type token = zecs.Entity
type placement = zecs.Entity

type Name string
type AliasingName string

func NameF(tpl string, v ...any) Name {
	return Name(fmt.Sprintf(tpl, v...))
}

type Connection struct{ From, To zecs.Entity }
type RegionMarker struct{}
type PlacementLocationMarker struct{}
type TokenMarker struct{}
type CollectableToken token
type DefaultPlacement zecs.Entity
type Fixed struct{}
type Collected struct{}
type Skipped struct{}

type ScriptDecl string
type ScriptSource string
type ScriptParsed struct{ ast.Node }

type RuleSource string
type RuleParsed struct{ ast.Node }
type RuleOptimized struct{ ast.Node }
type RuleCompiled compiler.Bytecode

type HoldsToken zecs.Entity
type HeldAt zecs.Entity
type Empty struct{}
type Generated struct{}
type Ptr objects.Object
type Price int
type Collectable struct{}
type LocationMarker struct{}
type EdgeKind uint8
type HintRegion string
type AltHintRegion string
type DungeonName string
type IsBossRoom struct{}
type Savewarp string
type Scene string
type TimePassess struct{}
type CollectablePriority uint8
type WorldGraphRoot struct{}

const (
	_             EdgeKind = 0
	EdgeTransit   EdgeKind = 0x69
	EdgePlacement EdgeKind = 0xBB

	PriorityJunk        CollectablePriority = 0
	PriorityMajor       CollectablePriority = 0xE0
	PriorityAdvancement CollectablePriority = 0xF0
)

type Compass struct{}
type Drop struct{}
type DungeonReward struct{}
type Event struct{}
type Item struct{}
type Map struct{}
type Refill struct{}
type Shop struct{}
type GoldSkulltulaToken struct{}
type Bottle struct{}
type Medallion struct{}
type Stone struct{}
type OcarinaNote rune
type SongNotes string

type DungeonGroup uint8
type SmallKey struct{}
type BossKey struct{}
type DungeonKeyRing struct{}
type SilverRupeePuzzle uint8
type SilverRupee struct{}
type SilverRupeePouch struct{}

type Song uint8
type WarpSong struct{}
type ScarecrowSong struct{}

const (
	SONG_PRELUDE   Song = 0x32
	SONG_BOLERO    Song = 0x33
	SONG_MINUET    Song = 0x34
	SONG_SERENADE  Song = 0x35
	SONG_REQUIEM   Song = 0x36
	SONG_NOCTURNE  Song = 0x37
	SONG_SARIA     Song = 0x44
	SONG_EPONA     Song = 0x45
	SONG_LULLABY   Song = 0x46
	SONG_SUN       Song = 0x47
	SONG_TIME      Song = 0x48
	SONG_STORMS    Song = 0x49
	SONG_SCARECROW Song = 0x60

	NOTE_A     OcarinaNote = 'A'
	NOTE_UP    OcarinaNote = '^'
	NOTE_LEFT  OcarinaNote = '<'
	NOTE_DOWN  OcarinaNote = 'v'
	NOTE_RIGHT OcarinaNote = '>'

	DUNGEON_GENERIC DungeonGroup = iota
	DUNGEON_DEKU_TREE
	DUNGEON_DODONGOS_CAVERN
	DUNGEON_JABU_JABU
	DUNGEON_FOREST_TEMPLE
	DUNGEON_FIRE_TEMPLE
	DUNGEON_WATER_TEMPLE
	DUNGEON_SPIRIT_TEMPLE
	DUNGEON_SHADOW_TEMPLE
	DUNGEON_BOTTOM_OF_THE_WELL
	DUNGEON_TRAINING_GROUNDS
	DUNGEON_HIDEOUT
	DUNGEON_GANON_CASTLE
	DUNGEON_TREASURE_CHEST_GAME
	DUNGEON_ICE_CAVERN

	_ SilverRupeePuzzle = iota
	SR_DC_STAIRCASE
	SR_ICE_SPINNING_SCYTHE
	SR_ICE_PUSH_BLOCK
	SR_WELL_BASEMENT
	SR_SHADOW_SCYTHE_SHORTCUT
	SR_SHADOW_INVISIBLE_BLADES
	SR_SHADOW_HUGE_PIT
	SR_SHADOW_INVISIBLE_SPIKES
	SR_GTG_SLOPES
	SR_GTG_LAVA
	SR_GTG_WATER
	SR_SPIRIT_CHILD_TORCHES
	SR_SPIRIT_ADULT_BOULDERS
	SR_SPIRIT_LOBBY
	SR_SPIRIT_SUN_BLOCK
	SR_SPIRIT_ADULT_CLIMB
	SR_GC_SPIRIT_TRIAL
	SR_GC_LIGHT_TRIAL
	SR_GC_FIRE_TRIAL
	SR_GC_SHADOW_TRIAL
	SR_GC_WATER_TRIAL
	SR_GC_FOREST_TRIAL
)

func ParseDungeonGroup(name string) DungeonGroup {
	parsed := groupName.FindStringSubmatch(name)

	if len(parsed) != 2 {
		return DUNGEON_GENERIC
	}

	switch parsed[1] {
	case "Deku Tree":
		return DUNGEON_DEKU_TREE
	case "Dodongos Cavern":
		return DUNGEON_DODONGOS_CAVERN
	case "Jabu Jabus Belly":
		return DUNGEON_JABU_JABU
	case "Forest Temple":
		return DUNGEON_FOREST_TEMPLE
	case "Fire Temple":
		return DUNGEON_FIRE_TEMPLE
	case "Water Temple":
		return DUNGEON_WATER_TEMPLE
	case "Spirit Temple":
		return DUNGEON_SPIRIT_TEMPLE
	case "Shadow Temple":
		return DUNGEON_SHADOW_TEMPLE
	case "Bottom of the Well":
		return DUNGEON_BOTTOM_OF_THE_WELL
	case "Gerudo Training Ground":
		return DUNGEON_TRAINING_GROUNDS
	case "Thieves Hideout":
		return DUNGEON_HIDEOUT
	case "Ganons Castle":
		return DUNGEON_GANON_CASTLE
	case "Treasure Chest Game":
		return DUNGEON_TREASURE_CHEST_GAME
	case "Ice Cavern":
		return DUNGEON_ICE_CAVERN
	default:
		panic(fmt.Errorf("unsupported key group %q", name))
	}
}

func ParseSilverRupeePuzzle(name string) SilverRupeePuzzle {
	parsed := groupName.FindStringSubmatch(name)
	switch parsed[1] {
	case "Dodongos Cavern Staircase":
		return SR_DC_STAIRCASE
	case "Ice Cavern Spinning Scythe":
		return SR_ICE_SPINNING_SCYTHE
	case "Ice Cavern Push Block":
		return SR_ICE_PUSH_BLOCK
	case "Bottom of the Well Basement":
		return SR_WELL_BASEMENT
	case "Shadow Temple Scythe Shortcut":
		return SR_SHADOW_SCYTHE_SHORTCUT
	case "Shadow Temple Invisible Blades":
		return SR_SHADOW_INVISIBLE_BLADES
	case "Shadow Temple Huge Pit":
		return SR_SHADOW_HUGE_PIT
	case "Shadow Temple Invisible Spikes":
		return SR_SHADOW_INVISIBLE_SPIKES
	case "Gerudo Training Ground Slopes":
		return SR_GTG_SLOPES
	case "Gerudo Training Ground Lava":
		return SR_GTG_LAVA
	case "Gerudo Training Ground Water":
		return SR_GTG_WATER
	case "Spirit Temple Child Early Torches":
		return SR_SPIRIT_CHILD_TORCHES
	case "Spirit Temple Adult Boulders":
		return SR_SPIRIT_ADULT_BOULDERS
	case "Spirit Temple Lobby and Lower Adult":
		return SR_SPIRIT_LOBBY
	case "Spirit Temple Sun Block":
		return SR_SPIRIT_SUN_BLOCK
	case "Spirit Temple Adult Climb":
		return SR_SPIRIT_ADULT_CLIMB
	case "Ganons Castle Spirit Trial":
		return SR_GC_SPIRIT_TRIAL
	case "Ganons Castle Light Trial":
		return SR_GC_LIGHT_TRIAL
	case "Ganons Castle Fire Trial":
		return SR_GC_FIRE_TRIAL
	case "Ganons Castle Shadow Trial":
		return SR_GC_SHADOW_TRIAL
	case "Ganons Castle Water Trial":
		return SR_GC_WATER_TRIAL
	case "Ganons Castle Forest Trial":
		return SR_GC_FOREST_TRIAL
	default:
		panic(fmt.Errorf("unsupported silver rupee group %q", name))
	}
}

var groupName = regexp.MustCompile(`^*\((.*)\)$`)
