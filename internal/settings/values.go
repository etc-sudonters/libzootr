package settings

import (
	"fmt"

	"github.com/etc-sudonters/substrate/slipup"
)

const (
	Forest Medallions = 0b00000001
	Fire   Medallions = 0b00000010
	Water  Medallions = 0b00000100
	Shadow Medallions = 0b00001000
	Spirit Medallions = 0b00010000
	Light  Medallions = 0b00100000
)

const (
	TrialsEnabledNone   TrialsEnabled = 0b00000000
	TrialsEnabledAll                  = 0b11111111
	TrialsEnabledRandom               = 0b10101010
	TrialsEnabledForest               = TrialsEnabled(Forest)
	TrialsEnabledFire                 = TrialsEnabled(Fire)
	TrialsEnabledWater                = TrialsEnabled(Water)
	TrialsEnabledShadow               = TrialsEnabled(Shadow)
	TrialsEnabledSpirit               = TrialsEnabled(Spirit)
	TrialsEnabledLight                = TrialsEnabled(Light)
)

func (this TrialsEnabled) Count() uint8 {
	switch this {
	case TrialsEnabledNone:
		return 0
	case TrialsEnabledAll:
		return 6
	case TrialsEnabledRandom:
		panic("cannot count random trials")
	default:
		var count uint8
		every := []TrialsEnabled{TrialsEnabledForest, TrialsEnabledFire, TrialsEnabledWater, TrialsEnabledShadow, TrialsEnabledSpirit, TrialsEnabledLight}
		for _, trial := range every {
			if Has(this, trial) {
				count++
			}
		}
		return count
	}
}

const (
	KeyRingsRandom        Keyrings = 0b1010101010101010
	KeyRingsOff           Keyrings = 0b0000000000000000
	KeyringFortress       Keyrings = 0b0000000000000001
	KeyringChestGame      Keyrings = 0b0000000000000010
	KeyringForest         Keyrings = 0b0000000000000100
	KeyringFire           Keyrings = 0b0000000000001000
	KeyringWater          Keyrings = 0b0000000000010000
	KeyringShadow         Keyrings = 0b0000000000100000
	KeyringSpirit         Keyrings = 0b0000000001000000
	KeyringWell           Keyrings = 0b0000000010000000
	KeyringTrainingGround Keyrings = 0b0000000100000000
	KeyringGanonsCastle   Keyrings = 0b0000001000000000
	KeyRingsAll           Keyrings = 0b0000001111111111
	KeyRingsGiveBossKey   Keyrings = 0b0100000000000000
)

const (
	LogicGlitchess LogicRuleSet = 2
	LogicGlitched  LogicRuleSet = 4
	LogicNone      LogicRuleSet = 8
)

func (this LogicRuleSet) String() string {
	switch this {
	case LogicGlitchess:
		return "glitchless"
	case LogicGlitched:
		return "glitched"
	case LogicNone:
		return "none"
	default:
		panic(fmt.Errorf("unknown logic set %x", uint8(this)))
	}
}

const (
	ReachableAll       ReachableLocations = 2
	ReachableRequired  ReachableLocations = 4
	ReachableGoalsOnly ReachableLocations = 8
)

func (this ReachableLocations) String() string {
	switch this {
	case ReachableAll:
		return "all"
	case ReachableGoalsOnly:
		return "goals"
	case ReachableRequired:
		return "beatable"
	default:
		panic(fmt.Errorf("unknown reachable setting %x", uint8(this)))
	}
}

const (
	GanonBKRemove GanonBKShuffleKind = 1 << iota
	GanonBKVanilla
	GanonBKDungeon
	GanonBKRegional
	GanonBKOverworld
	GanonBKAnyDungeon
	GanonBKKeysanity
	GanonBKOnLacs
	GanonBKStones
	GanonBKMedallions
	GanonBKDungeonRewards
	GanonBKTokens
	GanonBKHearts
	GanonBKTriforcePieces
)

const (
	DungeonRewardVanilla DungeonRewardShuffle = 1 << iota
	DungeonRewardComplete
	DungeonRewardDungeon
	DungeonRewardRegional
	DungeonRewardOverworld
	DungeonRewardAnyDungeon
	DungeonRewardAnywhere
)

func (this DungeonRewardShuffle) String() string {
	switch this {
	case DungeonRewardVanilla:
		return "vanilla"
	case DungeonRewardComplete:
		return "complete"
	case DungeonRewardDungeon:
		return "dungeon"
	case DungeonRewardRegional:
		return "regional"
	case DungeonRewardOverworld:
		return "overworld"
	case DungeonRewardAnyDungeon:
		return "any_dungeon"
	case DungeonRewardAnywhere:
		return "anywhere"
	default:
		panic(fmt.Errorf("unknown dungeon reward shuffle value %x", uint8(this)))
	}

}

const (
	KeysVanilla KeyShuffle = iota
	KeysRemove
	KeysDungeon
	KeysRegional
	KeysOverworld
	KeysAnyDungeon
	KeysAnywhere
)

func (this KeyShuffle) String() string {
	switch this {
	case KeysVanilla:
		return "vanilla"
	case KeysRemove:
		return "remove"
	case KeysDungeon:
		return "dungeon"
	case KeysRegional:
		return "regional"
	case KeysOverworld:
		return "overworld"
	case KeysAnyDungeon:
		return "any_dungeon"
	case KeysAnywhere:
		return "keysanity"

	default:
		panic(fmt.Errorf("unknown keyshuffle setting %x", uint8(this)))
	}

}

const (
	SilverRupeesOff              SilverRupeePouches = 0b00000000000000000000000000000000
	SilverRupeesAll              SilverRupeePouches = 0b11111111111111111111111111111111
	SilverRupeesRandom           SilverRupeePouches = 0b11100000000000000000000000000000
	SilverRupeesDodongoCavern    SilverRupeePouches = 0b00000000000000000000000000000001
	SilverRupeesIceCavernScythe  SilverRupeePouches = 0b00000000000000000000000000000010
	SilverRupeesIceCavernPush    SilverRupeePouches = 0b00000000000000000000000000000100
	SilverRupeesWellBasement     SilverRupeePouches = 0b00000000000000000000000000010000
	SilverRupeesShadowShortcut   SilverRupeePouches = 0b00000000000000000000000000100000
	SilverRupeesShadowBlades     SilverRupeePouches = 0b00000000000000000000000001000000
	SilverRupeesShadowHugePit    SilverRupeePouches = 0b00000000000000000000000010000000
	SilverRupeesShadowSpikes     SilverRupeePouches = 0b00000000000000000000000100000000
	SilverRupeesTrainingSlopes   SilverRupeePouches = 0b00000000000000000000001000000000
	SilverRupeesTrainingLava     SilverRupeePouches = 0b00000000000000000000010000000000
	SilverRupeesTrainingWater    SilverRupeePouches = 0b00000000000000000000100000000000
	SilverRupeesSpiritTorches    SilverRupeePouches = 0b00000000000000000001000000000000
	SilverRupeesSpiritBoulders   SilverRupeePouches = 0b00000000000000000010000000000000
	SilverRupeesSpiritSunBlock   SilverRupeePouches = 0b00000000000000000100000000000000
	SilverRupeesSpiritAdultClimb SilverRupeePouches = 0b00000000000000001000000000000000
	SilverRupeesTowerForestTrial SilverRupeePouches = 0b00000000000000010000000000000000
	SilverRupeesTowerFireTrial   SilverRupeePouches = 0b00000000000000100000000000000000
	SilverRupeesTowerWaterTrial  SilverRupeePouches = 0b00000000000001000000000000000000
	SilverRupeesTowerShadowTrial SilverRupeePouches = 0b00000000000010000000000000000000
	SilverRupeesTowerSpiritTrial SilverRupeePouches = 0b00000000000100000000000000000000
	SilverRupeesTowerLightTrial  SilverRupeePouches = 0b00000000001000000000000000000000
)

const (
	MapsCompassesVanilla                  = MapsCompasses(KeysVanilla)
	MapsCompassesRemove                   = MapsCompasses(KeysRemove)
	MapsCompassesDungeon                  = MapsCompasses(KeysDungeon)
	MapsCompassesRegional                 = MapsCompasses(KeysRegional)
	MapsCompassesOverworld                = MapsCompasses(KeysOverworld)
	MapsCompassesAnyDungeon               = MapsCompasses(KeysAnyDungeon)
	MapsCompassesAnywhere                 = MapsCompasses(KeysAnywhere)
	MapsCompassesStartWith                = MapsCompasses(1 << 8)
	MapsCompassesEnhanced   MapsCompasses = 0x0F00
)

const (
	KokriForestClosed OpenForest = iota
	KokriForestOpen
	KokriForestDekuClosed
)

func (this OpenForest) String() string {
	switch this {
	case KokriForestClosed:
		return "closed"
	case KokriForestOpen:
		return "open"
	case KokriForestDekuClosed:
		return "closed_deku"
	default:
		panic(fmt.Errorf("unknown open forest value %x", uint8(this)))
	}
}

const (
	KakGateClosed OpenKak = iota
	KakGateLetter
	KakGateOpen
)

func (this OpenKak) String() string {
	switch this {
	case KakGateClosed:
		return "closed"
	case KakGateLetter:
		return "zelda"
	case KakGateOpen:
		return "open"
	default:
		panic(fmt.Errorf("unknown open kak setting %x", uint8(this)))
	}
}

const (
	ZoraFountainClosed OpenZoraFountain = iota
	// move KZ for adult but not for child
	ZoraFountainOpenAdult
	// if KZ is moved for child, he's moved for adult
	ZoraFountainOpenAlways
)

func (z OpenZoraFountain) String() string {
	switch z {
	case ZoraFountainClosed:
		return "closed"
	case ZoraFountainOpenAdult:
		return "adult"
	case ZoraFountainOpenAlways:
		return "open"
	default:
		panic(slipup.Createf("unknown zora fountain setting %x", uint8(z)))
	}
}

const (
	GerudoFortressNormal GerudoFortress = iota
	GerudoFortressFast
	GerudoFortressOpen
)

func (this GerudoFortress) String() string {
	switch this {
	case GerudoFortressNormal:
		return "normal"
	case GerudoFortressFast:
		return "fast"
	case GerudoFortressOpen:
		return "open"
	default:
		panic(fmt.Errorf("unknown Gerudo fortress setting %x", uint8(this)))
	}
}

const (
	ShortcutsOff    DungeonShortcuts = 0b0000000000000000
	ShortcutsAll    DungeonShortcuts = 0b0000000011111111
	ShortcutsRandom DungeonShortcuts = 0b1010101010101010
	ShortcutsDeku   DungeonShortcuts = 0b0000000000000001
	ShortcutsCavern DungeonShortcuts = 0b0000000000000010
	ShortcutsJabu   DungeonShortcuts = 0b0000000000000100
	ShortcutsForest DungeonShortcuts = 0b0000000000001000
	ShortcutsFire   DungeonShortcuts = 0b0000000000010000
	ShortcutsWater  DungeonShortcuts = 0b0000000000100000
	ShortcutsShadow DungeonShortcuts = 0b0000000001000000
	ShortcutsSpirit DungeonShortcuts = 0b0000000010000000
)

const (
	StartAgeChild StartingAge = iota
	StartAgeAdult
	StartAgeRandom
)

const (
	// mask for upper 4 bits
	MasterQuestDungeonsNone     MasterQuestDungeons = 0b0000000000000000
	MasterQuestDungeonsAll      MasterQuestDungeons = 0b1000000000000000
	MasterQuestDungeonsSpecific MasterQuestDungeons = 0b0100000000000000
	MasterQuestDungeonsCount    MasterQuestDungeons = 0b0010000000000000
	MasterQuestDungeonsRandom   MasterQuestDungeons = 0b0001000000000000
	MasterQuestDekuTree         MasterQuestDungeons = 0b0000000000000001
	MasterQuestDodongoCavern    MasterQuestDungeons = 0b0000000000000010
	MasterQuestJabu             MasterQuestDungeons = 0b0000000000000100
	MasterQuestForest           MasterQuestDungeons = 0b0000000000001000
	MasterQuestFire             MasterQuestDungeons = 0b0000000000010000
	MasterQuestWater            MasterQuestDungeons = 0b0000000000100000
	MasterQuestShadow           MasterQuestDungeons = 0b0000000001000000
	MasterQuestSpirit           MasterQuestDungeons = 0b0000000010000000
	MasterQuestWell             MasterQuestDungeons = 0b0000000100000000
	MasterQuestIceCavern        MasterQuestDungeons = 0b0000001000000000
	MasterQuestTrainingGround   MasterQuestDungeons = 0b0000010000000000
	MasterQuestGanonsCastle     MasterQuestDungeons = 0b0000100000000000
)

const (
	CompletedDungeonsNone     CompletedDungeons = 0b0000000000000000
	CompletedDungeonsSpecific CompletedDungeons = 0b1000000000000000
	CompletedDungeonsRewards  CompletedDungeons = 0b0100000000000000
	CompletedDungeonsCount    CompletedDungeons = 0b0010000000000000
	CompletedDekuTree         CompletedDungeons = 0b0000000000000001
	CompletedDodongoCavern    CompletedDungeons = 0b0000000000000010
	CompletedJabu             CompletedDungeons = 0b0000000000000100
	CompletedForest           CompletedDungeons = 0b0000000000001000
	CompletedFire             CompletedDungeons = 0b0000000000010000
	CompletedWater            CompletedDungeons = 0b0000000000100000
	CompletedShadow           CompletedDungeons = 0b0000000001000000
	CompletedSpirit           CompletedDungeons = 0b0000000010000000
)

const (
	InteriorShuffleOff    InteriorShuffle = 0
	InteriorShuffleSimple InteriorShuffle = 2
	InteriorShuffleAll    InteriorShuffle = 4
)

const (
	DungeonEntranceShuffleOff    DungeonEntranceShuffle = 0
	DungeonEntranceShuffleSimple DungeonEntranceShuffle = 2
	DungeonEntranceShuffleAll    DungeonEntranceShuffle = 4
)

const (
	BossShuffleOff    BossShuffle = 0
	BossShuffleSimple BossShuffle = 2
	BossShuffleAll    BossShuffle = 4
)

const (
	SpawnVanilla     Spawn = 0
	RandomSpawn      Spawn = 0xF0F0F0F0F0
	SetSpawnLocation Spawn = 0x0F0F0F0F0F
)

const (
	ShuffleSongsOnSong ShuffleSongs = iota
	ShuffleSongsOnRewards
	ShuffleSongsAnywhere
)

const (
	// upper bits mask -- how are shops shuffled
	ShuffleShopsOff           ShuffleShops = 0
	ShuffleShopsSpecialRandom ShuffleShops = 0b01010101 << 8
	ShuffleShopsSpecial0      ShuffleShops = 0b11000000 << 8
	ShuffleShopsSpecial1      ShuffleShops = 0b10100000 << 8
	ShuffleShopsSpecial2      ShuffleShops = 0b10010000 << 8
	ShuffleShopsSpecial3      ShuffleShops = 0b10001000 << 8
	ShuffleShopsSpecial4      ShuffleShops = 0b10000100 << 8

	// lower bit mask -- do we have shop price caps
	ShuffleShopPricesRandom       ShuffleShops = 0b0000000001010101
	ShuffleShopPricesStartWallet  ShuffleShops = 0b0000000000000011
	ShuffleShopPricesAdultWallet  ShuffleShops = 0b0000000000000110
	ShuffleShopPricesGiantWallet  ShuffleShops = 0b0000000000001010
	ShuffleShopPricesTycoonWallet ShuffleShops = 0b0000000000010010
	ShuffleShopPricesAffordable   ShuffleShops = 0b0000000000100010
)

const (
	ShuffleGoldTokenOff       ShuffleTokens = 0
	ShuffleGoldTokenDungeons  ShuffleTokens = 1
	ShuffleGoldTokenOverworld ShuffleTokens = 2
)

const (
	ShuffleScrubsOff         ShuffleScrubs = 0 // off off
	ShuffleScrubsUpgradeOnly ShuffleScrubs = 1 // OOTR off
	ShuffleScrubsAffordable  ShuffleScrubs = 2
	ShuffleScrubsExpensive   ShuffleScrubs = 3
	ShuffleScrubsRandom      ShuffleScrubs = 4
)

func (this ShuffleScrubs) String() string {
	switch this {
	case ShuffleScrubsOff:
		return "really off"
	case ShuffleScrubsUpgradeOnly:
		return "off"
	case ShuffleScrubsAffordable:
		return "low"
	case ShuffleScrubsExpensive:
		return "regular"
	case ShuffleScrubsRandom:
		return "random"
	default:
		panic(fmt.Errorf("unknown scrub shuffle setting %x", uint8(this)))
	}

}

const (
	ShuffleFreestandingsOff       ShuffleFreestandings = 0
	ShuffleFreestandingsDungeon   ShuffleFreestandings = 1
	ShuffleFreestandingsOverworld ShuffleFreestandings = 2
)

const (
	ShufflePotsOff       ShufflePots = 0
	ShufflePotsAll       ShufflePots = 1
	ShufflePotsDungeons  ShufflePots = 2
	ShufflePotsOverworld ShufflePots = 4
)

func (this ShufflePots) String() string {
	switch this {
	case ShufflePotsOff:
		return "off"
	case ShufflePotsAll:
		return "all"
	case ShufflePotsDungeons:
		return "dungeons"
	case ShufflePotsOverworld:
		return "overworld"
	default:
		panic(fmt.Errorf("unknown pot shuffle value %x", uint8(this)))
	}
}

const (
	ShuffleCratesOff       ShuffleCrates = 0
	ShuffleCratesAll       ShuffleCrates = 1
	ShuffleCratesDungeons  ShuffleCrates = 2
	ShuffleCratesOverworld ShuffleCrates = 4
)

func (this ShuffleCrates) String() string {
	switch this {
	case ShuffleCratesOff:
		return "off"
	case ShuffleCratesAll:
		return "all"
	case ShuffleCratesDungeons:
		return "dungeons"
	case ShuffleCratesOverworld:
		return "overworld"
	default:
		panic(fmt.Errorf("unknown crate shuffle value %x", uint8(this)))
	}
}

const (
	ShuffleLoachRewardOff     ShuffleLoachReward = 0
	ShuffleLoachRewardVanilla ShuffleLoachReward = 1
	ShuffleLoachRewardEasy    ShuffleLoachReward = 2
)

const (
	ShuffleSongPatternsOff   ShuffleSongPatterns = 0
	ShuffleSongPatternsFrogs ShuffleSongPatterns = 1
	ShuffleSongPatternsWarps ShuffleSongPatterns = 2
)

const (
	HintsRevealedNever  HintsRevealed = 0
	HintsRevealedMask   HintsRevealed = 1
	HintsRevealedStone  HintsRevealed = 2
	HintsRevealedAlways HintsRevealed = 4
)

func (this HintsRevealed) String() string {
	switch this {
	case HintsRevealedNever:
		return "none"
	case HintsRevealedMask:
		return "mask"
	case HintsRevealedStone:
		return "agony"
	case HintsRevealedAlways:
		return "always"

	default:
		panic(fmt.Errorf("unknown hints value %x", uint8(this)))
	}
}

const (
	DamageMultiplierHalf   DamageMultiplier = 0
	DamageMultiplierNormal DamageMultiplier = 1
	DamageMultiplierDouble DamageMultiplier = 2
	DamageMultiplierQuad   DamageMultiplier = 4
	DamageMultiplierOhko   DamageMultiplier = 8
)

func (m DamageMultiplier) String() string {
	switch m {
	case DamageMultiplierHalf:
		return "half"
	case DamageMultiplierNormal:
		return "normal"
	case DamageMultiplierDouble:
		return "double"
	case DamageMultiplierQuad:
		return "quadruple"
	case DamageMultiplierOhko:
		return "ohko"
	default:
		panic(slipup.Createf("unknown damage multiple %x", uint(m)))
	}
}

const (
	BonkDamageNone   BonkDamage = 0
	BonkDamageHalf   BonkDamage = 1
	BonkDamageNormal BonkDamage = 2
	BonkDamageDouble BonkDamage = 4
	BonkDamageQuad   BonkDamage = 8
	BonkDamageOhko   BonkDamage = 16
)

func (m BonkDamage) String() string {
	switch m {
	case BonkDamageNone:
		return "none"
	case BonkDamageHalf:
		return "half"
	case BonkDamageNormal:
		return "normal"
	case BonkDamageDouble:
		return "double"
	case BonkDamageQuad:
		return "quadruple"
	case BonkDamageOhko:
		return "ohko"
	default:
		panic(slipup.Createf("unknown bonk damage %x", uint(m)))
	}
}

const (
	StartingTimeOfDayDefault StartingTimeOfDay = iota
	StartingTimeOfDayRandom
	StartingTimeOfDaySunrise
	StartingTimeOfDayMorning
	StartingTimeOfDayNoon
	StartingTimeOfDayAfternoon
	StartingTimeOfDaySunset
	StartingTimeOfDayEvening
	StartingTimeOfDayMidnight
	StartingTimeOfDayWitching
)

func (this StartingTimeOfDay) IsNight() bool {
	switch this {
	case StartingTimeOfDayEvening,
		StartingTimeOfDaySunset,
		StartingTimeOfDayMidnight,
		StartingTimeOfDayWitching:
		return true
	default:
		return false
	}
}

const (
	ItemPoolMinimal ItemPool = iota
	ItemPoolScarce
	ItemPoolDefault
	ItemPoolPlentiful
	ItemPoolLudicrous
)

const (
	IceTrapsOff IceTraps = 0
	IceTrapsNormal
	IceTrapsSomeExtraJunk
	IceTrapsAllExtraJunk
	IceTrapsAllJunk
)

const (
	AdultTradeShuffle           ShuffleTradeAdult = 0b1000000000000000
	AdultTradeShuffleDisabled   ShuffleTradeAdult = 0
	AdultTradeShuffleRandom     ShuffleTradeAdult = 0b1111111111111111
	AdultTradeStartPocketEgg    ShuffleTradeAdult = 0b0000000000000001
	AdultTradeStartPocketCucco  ShuffleTradeAdult = 0b0000000000000010
	AdultTradeStartCojiro       ShuffleTradeAdult = 0b0000000000000100
	AdultTradeStartOddMushroom  ShuffleTradeAdult = 0b0000000000001000
	AdultTradeStartOddPotion    ShuffleTradeAdult = 0b0000000000010000
	AdultTradeStartPoachersSaw  ShuffleTradeAdult = 0b0000000000100000
	AdultTradeStartBrokenSword  ShuffleTradeAdult = 0b0000000001000000
	AdultTradeStartPrescription ShuffleTradeAdult = 0b0000000010000000
	AdultTradeStartEyeballFrog  ShuffleTradeAdult = 0b0000000100000000
	AdultTradeStartEyedrops     ShuffleTradeAdult = 0b0000001000000000
	AdultTradeStartClaimCheck   ShuffleTradeAdult = 0b0000010000000000
)

const (
	ChildTradeShuffle  ShuffleTradeChild = 0b1000000000000000
	ChildTradeComplete ShuffleTradeChild = 0b1111111111111111

	ChildTradeStartWeirdEgg   ShuffleTradeChild = 0b0000000000000001
	ChildTradeStartChicken    ShuffleTradeChild = 0b0000000000000010
	ChildTradeStartLetter     ShuffleTradeChild = 0b0000000000000100
	ChildTradeStartMaskKeaton ShuffleTradeChild = 0b0000000000001000
	ChildTradeStartMaskSkull  ShuffleTradeChild = 0b0000000000010000
	ChildTradeStartMaskSpooky ShuffleTradeChild = 0b0000000000100000
	ChildTradeStartMaskBunny  ShuffleTradeChild = 0b0000000001000000
	ChildTradeStartMaskGoron  ShuffleTradeChild = 0b0000000010000000
	ChildTradeStartMaskZora   ShuffleTradeChild = 0b0000000100000000
	ChildTradeStartMaskGerudo ShuffleTradeChild = 0b0000001000000000
	ChildTradeStartMaskTruth  ShuffleTradeChild = 0b0000010000000000
)

const (
	ForestTemplePoesNone ForestTemplePoes = 0
	ForestTempleAmyMeg   ForestTemplePoes = 1
	ForestTempleJoBeth   ForestTemplePoes = 2
)

const (
	ScarecrowBehaviorDefault ScarecrowBehavior = 0
	ScarecrowBehaviorFast    ScarecrowBehavior = 1
	ScarecrowBehaviorFree    ScarecrowBehavior = 2
)

func (this ScarecrowBehavior) String() string {
	switch this {
	case ScarecrowBehaviorDefault:
		return "vanilla"
	case ScarecrowBehaviorFast:
		return "fast"
	case ScarecrowBehaviorFree:
		return "free"
	default:
		panic(slipup.Createf("unknown scarecrow behavior %x", uint(this)))
	}
}
