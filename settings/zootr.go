package settings

import "errors"

type Zootr struct {
	ShowSeedInfo      bool
	UserMessage       string
	WorldCount        int
	CreateSpoiler     bool
	PasswordLock      bool
	RandomizeSettings bool
	EnhanceMapCompass bool
	UseItemPool       string
	ShuffleIceTraps   string

	LogicRules         string
	ReachableLocations string
	StartingAge        string
	ShuffleSpawns      []string

	TriforceHunt           bool
	LacsCondition          string
	BridgeCondition        string
	TrialsRandom           bool
	TrialsEnabled          int
	ShuffleGanonBossKey    string
	GanonBossKeyMedallions int

	ShuffleBossKeys        string
	ShuffleSmallKeys       string
	ShuffleHideoutKeys     string
	ShuffleTCGKeys         string
	KeyRings               string
	ShuffleSilverRupees    string
	ShuffleMapCompass      string
	OneMajorItemPerDungeon bool
	MasterQuestDungonMode  string
	DungeonShortcuts       string
	EmptyDungeons          string

	OpenForest       string
	OpenKakariko     string
	OpenDoorOfTime   bool
	OpenZoraFountain string
	GerudoFortress   string

	ShuffleInteriorEntrances    string
	ShuffleGrottoEntrances      bool
	ShuffleDungeonEntrances     string
	ShuffleBosses               string
	ShuffleOverworldEntrances   bool
	ShuffleValleyRiverExit      bool
	ShuffleOwlDrops             bool
	ShuffleWarpSongDestinations bool

	FreeBombchuDrops bool

	ShuffleDungeonRewards     string
	ShuffleSongs              string
	ShopShuffle               string
	GoldSkullTokenShuffle     string
	ShuffleScrubs             string
	ShuffleFreestandingItems  string
	ShufflePots               string
	ShuffleCows               bool
	ShuffleBeehives           bool
	ShuffleWonderItems        bool
	ShuffleKokriSword         bool
	ShuffleOcarinas           bool
	ShuffleGerudoCard         bool
	ShuffleBeans              bool
	ShuffleExpensiveMerchants bool
	ShuffleFrogRewards        bool
	ShuffleOcarinaNotes       bool
	ShuffleFishingReward      string
	ShuffleAllAdultTradeItems bool
	ShuffledAdultTradeItems   []string
	ShuffleChildTrade         []string
	RandomizeSongNotes        string

	SkipRauruLightMedallionReward bool
	CompleteMaskQuest             bool
	FreeScarecrowSong             bool
	PlantBeans                    bool

	KakChickenCount int
	BigPoeCount     bool

	DamageMultiplier string
	BonksCauseDamage string

	StartingHearts       int
	StartWith            []StartingItem
	StartWithConsumabled bool
	StartWithRupees      bool

	RemoveCollectibleHearts bool
	StartingTimeOfDay       string
	UseBlueFireArrows       bool
	FixBrokenDrops          bool

	AllowedTricks []string

	DisabledLocations []string

	DisableGanonTowerCollapse      bool
	DisableHyruleCastleStealth     bool
	DisableEponaRace               bool
	SkipSomeMinigamePhases         bool
	KeepUsefulCutscenes            bool
	FastChests                     bool
	FastBunnyHood                  bool
	AutoEquipMasks                 bool
	RandomKakChickenCount          bool
	RandomBigPoeCount              bool
	EasierFireArrowEntry           bool
	RutoOnF1OfJabu                 bool
	ChestAppearanceMatchesContents string
	EnabledChestTextures           []string
	PlaceMinorItemsInMajorChest    []string
	MakeAllChestsInvisible         bool // invisible_chests -- wording is a little vague
	PotApperanceMatchesContents    string
	KeyApperanceMatchesDungeon     bool
	ClearerHints                   bool
	HintsRevealed                  string
	HintDistribution               string
	ItemHints                      []string
	UserHintDistrbution            map[string]any
	IncludeHintsAt                 []string
	ShuffleText                    string

	DisguiseIceTraps string
}

type StartingItem struct {
	Name string
	Qty  int
}

var notImpled = errors.New("unimplemented")

func decodeSettingStr(encoded string) (Zootr, error) {
	var z Zootr
	return z, notImpled
}
