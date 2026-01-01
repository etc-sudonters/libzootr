package settings

type flagged interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64
}

func Has[F flagged](setting, expecting F) bool {
	return setting&expecting == expecting

}

type Zootr struct {
	Seed            uint64
	Worlds          uint8
	LogicRules      LogicRuleSet
	TriforceHunt    TriforceHunt
	LacsCondition   LacsCondition
	BridgeCondition BridgeCondition
	KeyShuffle      KeyShuffling
	Locations       Locations
	Dungeons        Dungeons
	Entrances       EntranceRandomizer
	Spawns          SpawnSettings
	Shuffling       Shuffling
	Skills          Skills
	Starting        Starting
	Skips           Skips
	Minigames       Minigames
	Damage          Damage
	Trades          Trades
	ItemPool        ItemPool

	// uncategorized
	BlueFireArrows       bool
	DisabledLocations    []string
	FixBrokenDrops       bool
	FreeBombchuDrops     bool
	HintsRevealed        HintsRevealed
	ClearerHints         bool
	EnhanceMapAndCompass bool
	UsefulCutscenes      bool
	FastChests           bool
	NoCollectibleHearts  bool
	ScarecrowBehavior    ScarecrowBehavior
}

type ScarecrowBehavior uint8
type LogicRuleSet uint8
type Medallions uint8

type StartingAge uint8
type Spawn uint64

type GanonBKShuffleKind uint16
type KeyShuffle uint8
type Keyrings uint16
type SilverRupeePouches uint32

type ReachableLocations uint8
type OpenForest uint8
type OpenKak uint8
type OpenZoraFountain uint8
type GerudoFortress uint8

type TrialsEnabled uint8
type DungeonShortcuts uint16
type MasterQuestDungeons uint16
type DungeonRewardShuffle uint16
type CompletedDungeons uint16
type MapsCompasses uint16
type ForestTemplePoes uint8

type InteriorShuffle uint8
type DungeonEntranceShuffle uint8
type BossShuffle uint8

type ShuffleSongs uint8
type ShuffleShops uint16
type ShuffleTokens uint8
type ShuffleScrubs uint8
type ShuffleFreestandings uint8
type ShufflePots uint8
type ShuffleCrates uint8
type ShuffleLoachReward uint8

type DamageMultiplier uint8
type BonkDamage uint8

type ShuffleSongPatterns uint8
type HintsRevealed uint8
type StartingTimeOfDay uint8

type ItemPool uint8
type IceTraps uint8

type ShuffleTradeAdult uint16
type ShuffleTradeChild uint16

type Trades struct {
	Adult         ShuffleTradeAdult
	Child         ShuffleTradeChild
	DisableRevert bool
}

type Damage struct {
	Multiplier DamageMultiplier
	Bonk       BonkDamage
}

type Starting struct {
	PlantBeans        bool
	Hearts            uint8
	RauruReward       bool
	Rupees            uint16
	TimeOfDay         StartingTimeOfDay
	Tokens            []string // ootr uses equip, song and inventory separately
	WithConsumables   bool
	CompleteMaskQuest bool
}

type Skills struct {
	Tricks   map[string]bool
	Glitches map[string]bool

	ShadowFireArrowEntry uint8
}

type Skips struct {
	TowerEscape bool
	EponaRace   bool
	ChildZelda  bool

	RutoAlreadyOnFloor1 bool
	HyruleCastleStealth bool
}

type Minigames struct {
	CollapsePhases bool
	KakChickens    uint8
	BigPoeCount    uint8

	TreasureChestGameRequiresLens bool
}

type Shuffling struct {
	Beans                                bool
	Beehives                             bool
	Cows                                 bool
	Crates                               ShuffleCrates
	ExpensiveMerchants                   bool
	Freestandings                        ShuffleFreestandings
	FrogRupeeRewards                     bool
	GerudoCard                           bool
	KokriSword                           bool
	LoachReward                          ShuffleLoachReward
	NightTokensWithoutSuns               bool
	OcarinaNotes                         bool
	Ocarinas                             bool
	Pots                                 ShufflePots
	Scrubs                               ShuffleScrubs
	Shops                                ShuffleShops
	SongPatterns                         ShuffleSongPatterns
	Songs                                ShuffleSongs
	Tokens                               ShuffleTokens
	WonderItems                          bool
	IncludeEmptyPots, IncludeEmptyCrates bool
}

type SpawnSettings struct {
	StartingAge StartingAge
	AdultSpawn  Spawn
	ChildSpawn  Spawn
}

func (spawn SpawnSettings) Randomized() bool {
	return spawn.AdultSpawn != SpawnVanilla || spawn.ChildSpawn != SpawnVanilla
}

func (er EntranceRandomizer) ShufflingAny() bool {
	return er.Interior != InteriorShuffleOff ||
		er.DungeonEntrances != DungeonEntranceShuffleOff ||
		er.Bosses != BossShuffleOff ||
		er.HideoutEntrances ||
		er.Grottos ||
		er.Overworld ||
		er.RiverExit ||
		er.OwlDrops ||
		er.WarpSongs
}

type EntranceRandomizer struct {
	Interior         InteriorShuffle
	DungeonEntrances DungeonEntranceShuffle
	Bosses           BossShuffle
	HideoutEntrances bool
	Grottos          bool
	Overworld        bool
	RiverExit        bool
	OwlDrops         bool
	WarpSongs        bool
	Tower            bool
	ValleyExit       bool
}

func (this EntranceRandomizer) AffectedTodChecks() bool {
	return this.Overworld || (this.Interior != InteriorShuffleOff)
}

type Dungeons struct {
	Trials        TrialsEnabled
	Shortcuts     DungeonShortcuts
	MasterQuest   MasterQuestDungeons
	Rewards       DungeonRewardShuffle
	MapsCompasses MapsCompasses
	OneItemPer    bool
	Completed     CompletedDungeons

	RandomTrials bool

	ForestTemplePoes
}

type Locations struct {
	ReachableLocations ReachableLocations
	KokriForest        OpenForest
	Kakariko           OpenKak
	OpenDoorOfTime     bool
	ZoraFountain       OpenZoraFountain
	GerudoFortress     GerudoFortress
	Disabled           []string
}

type TriforceHunt struct {
	CountPerWorld, GoalPerWorld uint
}

type KeyShuffling struct {
	BossKeys           KeyShuffle
	TreasureChestGame  KeyShuffle
	GanonBKCondition   GanonBKCondition
	GanonBKShuffle     GanonBKShuffleKind
	HideoutKeys        KeyShuffle
	Keyrings           Keyrings
	SilverRupeePouches SilverRupeePouches
	SilverRupees       KeyShuffle
	SmallKeys          KeyShuffle
}
