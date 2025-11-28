package settings

import (
	"errors"
	"fmt"
	"math/bits"
)

type LogicSetting uint8

func (this LogicSetting) String() string {
	switch this {
	case LogicNone:
		return "none"
	case LogicGlitched:
		return "glitched"
	case LogicGlitchless:
		return "glitchless"
	default:
		panic("unreachable")
	}
}

type LocationsReachable uint8

func (this LocationsReachable) String() string {
	switch this {
	case ReachableAll:
		return "all"
	case ReachableGoals:
		return "goal"
	case ReachableNecessary:
		return "beatable"
	default:
		panic("unreachable")
	}
}

type ConditionedAmount uint64
type ConditionKind uint32

type BossShuffle uint8
type InteriorShuffle uint8
type OverworldShuffle uint8
type DungeonDoorShuffle uint8

type Enum uint8

type Flags uint64
type Flag uint64

func CountFlags[U ~uint64](union U) int {
	return bits.OnesCount64(uint64(union))
}

func HasFlag[U ~uint64](union U, flag U) bool {
	return uint64(union)&uint64(flag) == uint64(flag)
}

func EncodeConditionedAmount(kind ConditionKind, qty uint32) ConditionedAmount {
	return ConditionedAmount(uint64(kind)<<32 | uint64(qty))
}

func (this ConditionedAmount) Kind() ConditionKind {
	return ConditionKind(uint64(this) >> 32)
}

func (this ConditionedAmount) Amount() uint32 {
	return uint32(this)
}

func (this ConditionedAmount) Decode() (ConditionKind, uint32) {
	kind := ConditionKind(this >> 32)
	qty := uint32(kind)
	return kind, qty
}

type TrialFlag uint64

func ParseLogic(raw string) (LogicSetting, error) {
	switch raw {
	case "none":
		return LogicNone, nil
	case "glitched":
		return LogicGlitched, nil
	case "glitchless":
		return LogicGlitchless, nil
	default:
		return LogicGlitchless, fmt.Errorf("unknown logic setting: %q", raw)
	}
}

func ParseReachability(value string) (LocationsReachable, error) {
	switch value {
	case "all":
		return ReachableAll, nil
	case "goals":
		return ReachableGoals, nil
	case "beatable":
		return ReachableNecessary, nil
	default:
		return ReachableAll, fmt.Errorf("unknown reachable_locations: %q", value)
	}

}

const (
	LogicGlitchless LogicSetting = iota
	LogicNone
	LogicGlitched

	ReachableAll LocationsReachable = iota
	ReachableGoals
	ReachableNecessary

	CondUnitialized ConditionKind = iota
	CondDefault
	CondMedallions
	CondStones
	CondRewards
	CondTokens
	CondHearts
	CondOpen
	CondVanilla
	CondTriforce

	TrialForest TrialFlag = 1 << iota
	TrialFire
	TrialWater
	TrialShadow
	TrialSpirit
	TrialLight
	TrialAll = TrialForest | TrialFire | TrialWater | TrialShadow | TrialSpirit | TrialLight
)

func (this ConditionKind) String() string {
	switch this {
	case CondUnitialized:
		panic(errors.New("uninitialized condition"))
	case CondMedallions:
		return "medallions"
	case CondStones:
		return "stones"
	case CondRewards:
		return "rewards"
	case CondTokens:
		return "tokens"
	case CondHearts:
		return "hearts"
	case CondDefault:
		return "default"
	case CondOpen:
		return "open"
	case CondVanilla:
		return "vanilla"
	case CondTriforce:
		return "triforce"
	default:
		panic(fmt.Errorf("unknown condition flag %x", uint8(this)))
	}
}

func ParseCondition(raw string) (ConditionKind, error) {
	switch raw {
	case "medallions":
		return CondMedallions, nil
	case "stones":
		return CondStones, nil
	case "rewards":
		return CondRewards, nil
	case "tokens":
		return CondTokens, nil
	case "hearts":
		return CondHearts, nil
	case "default":
		return CondDefault, nil
	case "open":
		return CondOpen, nil
	case "vanilla":
		return CondVanilla, nil
	case "triforce":
		return CondTriforce, nil
	default:
		return CondUnitialized, fmt.Errorf("unknown condition %q", raw)
	}
}

func ConditionFrom(raw string, qty uint32) (ConditionedAmount, error) {
	cond, err := ParseCondition(raw)
	if err != nil {
		return 0, err
	}

	return EncodeConditionedAmount(cond, qty), nil
}

type OpenForest uint8

const (
	KokriForestClosed OpenForest = iota
	KokriForestOpen
	KokriForestClosedDeku
)

func (this OpenForest) String() string {
	switch this {
	case KokriForestOpen:
		return "open"
	case KokriForestClosedDeku:
		return "closed_deku"
	case KokriForestClosed:
		return "closed"
	default:
		panic("unreachable")
	}
}

func ParseOpenForest(raw string) (OpenForest, error) {
	switch raw {
	case "open":
		return KokriForestOpen, nil
	case "closed_deku":
		return KokriForestClosedDeku, nil
	case "closed":
		return KokriForestClosed, nil
	default:
		return 0, fmt.Errorf("unknown open forest setting: %q", raw)
	}
}

type OpenKakarikoGate uint8

const (
	KakGateClosed OpenKakarikoGate = iota
	KakGateOpen
	KakGateZelda
)

func (this OpenKakarikoGate) String() string {
	switch this {
	case KakGateOpen:
		return "open"
	case KakGateZelda:
		return "zelda"
	case KakGateClosed:
		return "closed"
	default:
		panic("unreachable")
	}
}

func ParseKakarikoGate(raw string) (OpenKakarikoGate, error) {
	switch raw {
	case "open":
		return KakGateOpen, nil
	case "zelda":
		return KakGateZelda, nil
	case "closed":
		return KakGateClosed, nil
	default:
		return 0, fmt.Errorf("unknown open kakariko gate setting: %q", raw)
	}
}

type OpenZoraFountain uint8

const (
	ZoraFountainClosed OpenZoraFountain = iota
	ZoraFountainOpenAdult
	ZoraFountainOpen
)

func (this OpenZoraFountain) String() string {
	switch this {
	case ZoraFountainClosed:
		return "closed"
	case ZoraFountainOpenAdult:
		return "adult"
	case ZoraFountainOpen:
		return "open"
	default:
		panic("unreachable")
	}
}

func ParseOpenZoraFountain(raw string) (OpenZoraFountain, error) {
	switch raw {
	case "closed":
		return ZoraFountainClosed, nil
	case "adult":
		return ZoraFountainOpenAdult, nil
	case "open":
		return ZoraFountainOpen, nil
	default:
		return 0, fmt.Errorf("unknown open zora fountain setting: %q", raw)
	}
}

type GerudoFortressCarpenterRescue uint8

const (
	RescueAllCarpenters GerudoFortressCarpenterRescue = iota
	RescueOneCarpenters
	RescueZeroCarpenters
)

func (this GerudoFortressCarpenterRescue) String() string {
	switch this {
	case RescueAllCarpenters:
		return "normal"
	case RescueOneCarpenters:
		return "fast"
	case RescueZeroCarpenters:
		return "open"
	default:
		panic("unreachable")
	}
}

func ParseGerudoFortressCarpenterRescue(raw string) (GerudoFortressCarpenterRescue, error) {
	switch raw {
	case "normal":
		return RescueAllCarpenters, nil
	case "fast":
		return RescueOneCarpenters, nil
	case "open":
		return RescueZeroCarpenters, nil
	default:
		return 0, fmt.Errorf("unknown gerudo fortress carpenter rescue setting: %q", raw)
	}
}

type ShuffleScrub uint8

const (
	ShuffleUpgradeScrub ShuffleScrub = iota
	ShuffleScrubsAffordable
	ShuffleScrubsExpensive
	ShuffleScrubsRandomPrices
)

func (this ShuffleScrub) String() string {
	switch this {
	case ShuffleUpgradeScrub:
		return "off"
	case ShuffleScrubsAffordable:
		return "low"
	case ShuffleScrubsExpensive:
		return "regular"
	case ShuffleScrubsRandomPrices:
		return "random"
	default:
		panic("unreachable")
	}
}

type PartitionedShuffle uint8

type ShuffleSkullTokens PartitionedShuffle
type ShufflePots PartitionedShuffle
type ShuffleCrates PartitionedShuffle
type ShuffleFreestanding PartitionedShuffle

const (
	ShufflePartionOff PartitionedShuffle = iota
	ShufflePartitionDungeons
	ShufflePartitionOverworld
	ShufflePartionAll
)

func (this PartitionedShuffle) String() string {
	switch this {
	case ShufflePartionOff:
		return "off"
	case ShufflePartitionDungeons:
		return "dungeons"
	case ShufflePartitionOverworld:
		return "overworld"
	case ShufflePartionAll:
		return "all"
	default:
		panic("unreachable")
	}
}

func ParsePartitionedShuffle(raw string) (PartitionedShuffle, error) {
	switch raw {
	case "off":
		return ShufflePartionOff, nil
	case "dungeons":
		return ShufflePartitionDungeons, nil
	case "overworld":
		return ShufflePartitionOverworld, nil
	case "all":
		return ShufflePartionAll, nil
	default:
		return 0, fmt.Errorf("unknown partition %q", raw)
	}
}

type ShuffleDungeonRewards uint8

const (
	ShuffleDungeonRewardsReward ShuffleDungeonRewards = iota
	ShuffleDungeonRewardsVanilla
	ShuffleDungeonRewardsOwnDungeon
	ShuffleDungeonRewardsRegional
	ShuffleDungeonRewardsOverworld
	ShuffleDungeonRewardsAnyDungeon
	ShuffleDungeonRewardsAnywhere
)

func (this ShuffleDungeonRewards) String() string {

	switch this {
	case ShuffleDungeonRewardsVanilla:
		return "vanilla"
	case ShuffleDungeonRewardsReward:
		return "reward"
	case ShuffleDungeonRewardsOwnDungeon:
		return "dungeon"
	case ShuffleDungeonRewardsRegional:
		return "regional"
	case ShuffleDungeonRewardsOverworld:
		return "overworld"
	case ShuffleDungeonRewardsAnyDungeon:
		return "any_dungeon"
	case ShuffleDungeonRewardsAnywhere:
		return "anywhere"
	default:
		panic("unreachable")
	}
}

func ParseShuffleDungeonReward(raw string) (ShuffleDungeonRewards, error) {

	switch raw {
	case "vanilla":
		return ShuffleDungeonRewardsVanilla, nil
	case "reward":
		return ShuffleDungeonRewardsReward, nil
	case "dungeon":
		return ShuffleDungeonRewardsOwnDungeon, nil
	case "regional":
		return ShuffleDungeonRewardsRegional, nil
	case "overworld":
		return ShuffleDungeonRewardsOverworld, nil
	case "any_dungeon":
		return ShuffleDungeonRewardsAnyDungeon, nil
	case "anywhere":
		return ShuffleDungeonRewardsAnywhere, nil
	default:
		return 0, fmt.Errorf("unknown shuffle_dungeon_rewards: %q", raw)
	}
}

type ShuffleKeys uint8

const (
	ShuffleKeyOwnDungeon ShuffleKeys = iota
	ShuffleKeysVanilla
	ShuffleKeysRemove
	ShuffleKeyRegional
	ShuffleKeyOverworld
	ShuffleKeyAnyDungeon
	ShuffleKeysAnywhere
)

func (this ShuffleKeys) String() string {
	switch this {
	case ShuffleKeysVanilla:
		return "vanilla"
	case ShuffleKeysRemove:
		return "remove"
	case ShuffleKeyOwnDungeon:
		return "dungeon"
	case ShuffleKeyRegional:
		return "regional"
	case ShuffleKeyOverworld:
		return "overworld"
	case ShuffleKeyAnyDungeon:
		return "any_dungeon"
	case ShuffleKeysAnywhere:
		return "keysanity"
	default:
		panic("unreachable")
	}
}

func ParseShuffleKeys(raw string) (ShuffleKeys, error) {
	switch raw {
	case "vanilla":
		return ShuffleKeysVanilla, nil
	case "remove":
		return ShuffleKeysRemove, nil
	case "dungeon":
		return ShuffleKeyOwnDungeon, nil
	case "regional":
		return ShuffleKeyRegional, nil
	case "overworld":
		return ShuffleKeyOverworld, nil
	case "any_dungeon":
		return ShuffleKeyAnyDungeon, nil
	case "keysanity":
		return ShuffleKeysAnywhere, nil
	default:
		return ShuffleKeyOwnDungeon, fmt.Errorf("unknown key shuffle: %s", raw)
	}
}

type GanonBossKeyShuffle uint8

const (
	GanonBossKeyInDungeon GanonBossKeyShuffle = iota
	GanonBossKeyRemove
	GanonBossKeyVanilla
	GanonBossKeyRegional
	GanonBossKeyOverworld
	GanonBossKeyAnyDungeon
	GanonBossKeyAnywhere
	GanonBossKeyOnLacs
	GanonBossKeyStones
	GanonBossKeyMedallions
	GanonBossKeyTokens
	GanonBossKeyHearts
)

func (this GanonBossKeyShuffle) String() string {
	switch this {
	case GanonBossKeyInDungeon:
		return "dungeon"
	case GanonBossKeyRemove:
		return "remove"
	case GanonBossKeyVanilla:
		return "vanilla"
	case GanonBossKeyRegional:
		return "regional"
	case GanonBossKeyOverworld:
		return "overworld"
	case GanonBossKeyAnyDungeon:
		return "any_dungeon"
	case GanonBossKeyAnywhere:
		return "keysanity"
	case GanonBossKeyOnLacs:
		return "on_lacs"
	case GanonBossKeyStones:
		return "stones"
	case GanonBossKeyMedallions:
		return "medallions"
	case GanonBossKeyTokens:
		return "tokens"
	case GanonBossKeyHearts:
		return "hearts"
	default:
		panic("unreachable")
	}

}

func ParseGanonBossKeyShuffle(raw string) (GanonBossKeyShuffle, error) {
	switch raw {
	case "dungeon":
		return GanonBossKeyInDungeon, nil
	case "remove":
		return GanonBossKeyRemove, nil
	case "vanilla":
		return GanonBossKeyVanilla, nil
	case "regional":
		return GanonBossKeyRegional, nil
	case "overworld":
		return GanonBossKeyOverworld, nil
	case "any_dungeon":
		return GanonBossKeyAnyDungeon, nil
	case "keysanity":
		return GanonBossKeyAnywhere, nil
	case "on_lacs":
		return GanonBossKeyOnLacs, nil
	case "stones":
		return GanonBossKeyStones, nil
	case "medallions":
		return GanonBossKeyMedallions, nil
	case "tokens":
		return GanonBossKeyTokens, nil
	case "hearts":
		return GanonBossKeyHearts, nil
	default:
		return GanonBossKeyInDungeon, fmt.Errorf("unknown ganon boss key shuffle: %q", raw)
	}
}

type HintsRevealed uint8

const (
	HintsRevealedNever HintsRevealed = iota
	HintsRevealedWithMask
	HintsRevealedWithStone
	HintsRevealedAlways
)

func (this HintsRevealed) String() string {
	switch this {
	case HintsRevealedNever:
		return "none"
	case HintsRevealedWithMask:
		return "mask"
	case HintsRevealedWithStone:
		return "agony"
	case HintsRevealedAlways:
		return "always"
	default:
		panic("unreachable")
	}
}

type DamageMultiplier uint8

const (
	DamageMultiplierNormal DamageMultiplier = iota
	DamageMultiplierHalf
	DamageMultiplierDouble
	DamageMultiplierQuadruple
	DamageMultiplierOHKO
	DamageMultiplierNone // deadly bonks
)

func (this DamageMultiplier) String() string {
	switch this {
	case DamageMultiplierNormal:
		return "normal"
	case DamageMultiplierHalf:
		return "half"
	case DamageMultiplierDouble:
		return "double"
	case DamageMultiplierQuadruple:
		return "quadruple"
	case DamageMultiplierOHKO:
		return "ohko"
	case DamageMultiplierNone:
		return "none"
	default:
		panic("unreachable")
	}
}

func ParseDamageMultiplier(raw string) (DamageMultiplier, error) {
	switch raw {
	case "normal":
		return DamageMultiplierNormal, nil
	case "half":
		return DamageMultiplierHalf, nil
	case "double":
		return DamageMultiplierDouble, nil
	case "quadruple":
		return DamageMultiplierQuadruple, nil
	case "ohko":
		return DamageMultiplierOHKO, nil
	case "none":
		return DamageMultiplierNone, nil
	default:
		return 0, fmt.Errorf("unknown damage multiplier %q", raw)
	}
}

type GerudoFortressHeartPiece uint

const (
	GerudoFortressHeartPieceRemove GerudoFortressHeartPiece = iota
	GerudoFortressHeartPieceVanilla
	GerudoFortressHeartPieceShuffle
)

func (this GerudoFortressHeartPiece) String() string {
	switch this {
	case GerudoFortressHeartPieceRemove:
		return "remove"
	case GerudoFortressHeartPieceVanilla:
		return "vanilla"
	case GerudoFortressHeartPieceShuffle:
		return "shuffle"
	default:
		panic("unreachable")
	}
}

type ShufflingFlags uint64

const (
	ShuffleEmptyPots ShufflingFlags = 1 << iota
	ShuffleEmptyCrates
	ShuffleCows
	ShuffleOcarinaNotes
	ShuffleBeehives
	ShuffleWonderItems
	ShuffleKokiriSword
	ShuffleOcarinas
	ShuffleGerudoCard
	ShuffleBeans
	ShuffleExpensiveMerchants
	ShuffleFrogRupees
)

type ShuffleLoachReward uint8

const (
	ShuffleLoachRewardOff ShuffleLoachReward = iota
	ShuffleLoachRewardVanilla
	ShuffleLoachRewardEasy
)

func ParseShuffleLoachReward(raw string) (ShuffleLoachReward, error) {
	switch raw {
	case "off":
		return ShuffleLoachRewardOff, nil
	case "vanilla":
		return ShuffleLoachRewardVanilla, nil
	case "easy":
		return ShuffleLoachRewardEasy, nil
	default:
		return 0, fmt.Errorf("unknown loach shuffle setting: %q", raw)
	}
}

type TimeOfDay uint8

const (
	TimeOfDayDefault TimeOfDay = iota
	TimeOfDaySunrise
	TimeOfDayMorning
	TimeOfDayNoon
	TimeOfDayAfternoon
	TimeOfDaySunset
	TimeOfDayEvening
	TimeOfDayMidnight
	TimeOfDayWitching
)

func (this TimeOfDay) IsNight() bool {
	switch this {
	case TimeOfDayEvening,
		TimeOfDaySunset,
		TimeOfDayMidnight,
		TimeOfDayWitching:
		return true
	default:
		return false
	}
}

func ParseTimeOfDay(raw string) (TimeOfDay, error) {
	switch raw {
	case "default":
		return TimeOfDayDefault, nil
	case "sunrise":
		return TimeOfDaySunrise, nil
	case "morning":
		return TimeOfDayMorning, nil
	case "noon":
		return TimeOfDayNoon, nil
	case "afternoon":
		return TimeOfDayAfternoon, nil
	case "sunset":
		return TimeOfDaySunset, nil
	case "evening":
		return TimeOfDayEvening, nil
	case "midnight":
		return TimeOfDayMidnight, nil
	case "witching-hour":
		return TimeOfDayWitching, nil
	default:
		return 0, fmt.Errorf("unknown time of day: %q", raw)
	}

}

type ChildTradeItems uint64

type AdultTradeItems uint64

const (
	AdultTradePocketEgg AdultTradeItems = 1 << iota
	AdultTradePocketCucco
	AdultTradeOddMushroom
	AdultTradeOddPotion
	AdultTradePoachersSaw
	AdultTradeBrokenSword
	AdultTradePrescription
	AdultTradeEyeballFrog
	AdultTradeEyedrops
	AdultTradeClaimCheck
	AdultTradeItemsAll = AdultTradePocketEgg | AdultTradePocketCucco |
		AdultTradeOddMushroom | AdultTradeOddPotion |
		AdultTradePoachersSaw | AdultTradeBrokenSword |
		AdultTradePrescription | AdultTradeEyeballFrog |
		AdultTradeEyedrops | AdultTradeClaimCheck

	ChildTradeItemWeirdEgg ChildTradeItems = iota
	ChildTradeItemChicken
	ChildTradeItemZeldasLetter
	ChildTradeItemKeatonMask
	ChildTradeItemSkullMask
	ChildTradeItemSpookyMask
	ChildTradeItemBunnyHood
	ChildTradeItemGoronMask
	ChildTradeItemZoraMask
	ChildTradeItemGerudoMask
	ChildTradeItemMaskOfTruth
	ChildTradeItemsAll = ChildTradeItemWeirdEgg | ChildTradeItemChicken |
		ChildTradeItemZeldasLetter | ChildTradeItemKeatonMask |
		ChildTradeItemSkullMask | ChildTradeItemSpookyMask |
		ChildTradeItemBunnyHood | ChildTradeItemGoronMask | ChildTradeItemZoraMask |
		ChildTradeItemGerudoMask | ChildTradeItemMaskOfTruth
)

func ParseChildTradeItem(raw string) (ChildTradeItems, error) {
	switch raw {
	case "Weird Egg":
		return ChildTradeItemWeirdEgg, nil
	case "Chicken":
		return ChildTradeItemChicken, nil
	case "Zeldas Letter":
		return ChildTradeItemZeldasLetter, nil
	case "Keaton Mask":
		return ChildTradeItemKeatonMask, nil
	case "Skull Mask":
		return ChildTradeItemSkullMask, nil
	case "Spooky Mask":
		return ChildTradeItemSpookyMask, nil
	case "Bunny Hood":
		return ChildTradeItemBunnyHood, nil
	case "Goron Mask":
		return ChildTradeItemGoronMask, nil
	case "Zora Mask":
		return ChildTradeItemZoraMask, nil
	case "Gerudo Mask":
		return ChildTradeItemGerudoMask, nil
	case "Mask of Truth":
		return ChildTradeItemMaskOfTruth, nil
	default:
		return 0, fmt.Errorf("unknown child trade item %q", raw)
	}
}

func ParseAdultTradeItem(raw string) (AdultTradeItems, error) {
	switch raw {
	case "Pocket Egg":
		return AdultTradePocketEgg, nil
	case "Pocket Cucco":
		return AdultTradePocketCucco, nil
	case "Odd Mushroom":
		return AdultTradeOddMushroom, nil
	case "Odd Potion":
		return AdultTradeOddPotion, nil
	case "Poachers Saw":
		return AdultTradePoachersSaw, nil
	case "Broken Sword":
		return AdultTradeBrokenSword, nil
	case "Prescription":
		return AdultTradePrescription, nil
	case "Eyeball Frog":
		return AdultTradeEyeballFrog, nil
	case "Eyedrops":
		return AdultTradeEyedrops, nil
	case "Claim Check":
		return AdultTradeClaimCheck, nil
	default:
		return 0, fmt.Errorf("unknown adult trade shuffle item %q", raw)
	}
}

type StartAge bool

const (
	StartAgeAdult StartAge = true
	StartAgeChild StartAge = false
)

type LocationFlags uint64

const (
	LocationSkipRauruReward LocationFlags = 1 << iota
	LocationSkipChildZelda
	LocationsFreeScarecrow
	LocationsPlantBeans
	LocationsCompleteMaskQuest
)

type ShuffleSongs uint8

const (
	ShuffleSongsOnSongs ShuffleSongs = iota
	ShuffleSongsOnDungeonRewards
	ShuffleSongsAnywhere
)

func ParseShuffleSong(raw string) (ShuffleSongs, error) {
	switch raw {
	case "song":
		return ShuffleSongsOnSongs, nil
	case "dungeon":
		return ShuffleSongsOnDungeonRewards, nil
	case "any":
		return ShuffleSongsAnywhere, nil
	default:
		return 0, fmt.Errorf("unknown song shuffle: %q", raw)
	}

}

type ShuffleShop uint8

const (
	ShuffleShopsOff ShuffleShop = iota
	ShuffleShopsZero
	ShuffleShopsOne
	ShuffleShopTwo
	ShuffleShopThree
	ShuffleShopFour
)

func ParseShuffleShop(raw string) (ShuffleShop, error) {
	switch raw {
	case "off":
		return ShuffleShopsOff, nil
	case "0":
		return ShuffleShopsZero, nil
	case "1":
		return ShuffleShopsOne, nil
	case "2":
		return ShuffleShopTwo, nil
	case "3":
		return ShuffleShopFour, nil
	case "4":
		return ShuffleShopFour, nil
	default:
		return 0, fmt.Errorf("unknown shop shuffle setting %q", raw)
	}
}

type ShuffleShopPrices uint8

const (
	ShuffleShopPricesRandom ShuffleShopPrices = iota
	ShuffleShopPricesStartingWallet
	ShuffleShopPricesAdultWallet
	ShuffleShopPricesGiantWallet
	ShuffleShopPricesTycoonWallet
	ShuffleShopPricesAffordable
)

func ParseShuffleShopPrices(raw string) (ShuffleShopPrices, error) {
	switch raw {
	case "random":
		return ShuffleShopPricesRandom, nil
	case "random_starting":
		return ShuffleShopPricesStartingWallet, nil
	case "random_adult":
		return ShuffleShopPricesAdultWallet, nil
	case "random_giant":
		return ShuffleShopPricesGiantWallet, nil
	case "random_tycoon":
		return ShuffleShopPricesTycoonWallet, nil
	case "affordable":
		return ShuffleShopPricesAffordable, nil
	default:
		return 0, fmt.Errorf("unknown shop price shuffle %q", raw)
	}
}

type ShuffleMapCompass uint8

const (
	ShuffleMapCompassDungeon ShuffleMapCompass = iota
	ShuffleMapCompassRemove
	ShuffleMapCompassStartWith
	ShuffleMapCompassVanilla
	ShuffleMapCompassRegional
	ShuffleMapCompassOverworld
	ShuffleMapCompassAnyDungeon
	ShuffleMapCompassAnywhere
)

func ParseMapCompass(raw string) (ShuffleMapCompass, error) {
	switch raw {
	case "remove":
		return ShuffleMapCompassRemove, nil
	case "startwith":
		return ShuffleMapCompassStartWith, nil
	case "vanilla":
		return ShuffleMapCompassVanilla, nil
	case "dungeon":
		return ShuffleMapCompassDungeon, nil
	case "regional":
		return ShuffleMapCompassRegional, nil
	case "overworld":
		return ShuffleMapCompassOverworld, nil
	case "any_dungeon":
		return ShuffleMapCompassAnyDungeon, nil
	case "keysanity":
		return ShuffleMapCompassAnywhere, nil
	default:
		return 0, fmt.Errorf("unknown map & compass shuffle: %q", raw)
	}
}

type ConnectionFlag uint64

const (
	ConnectionOpenDoorOfTime ConnectionFlag = 1 << iota
	ConnectionShuffleHideoutEntrances
	ConnectionShuffleWarpSongDestinations
)

type ShuffleSongComposition uint8

const (
	ShuffleSongCompositionOff ShuffleSongComposition = iota
	ShuffleSongCompositionFrogs
	ShuffleSongCompositionWarp
	ShuffleSongCompositionAll
)

func ParseShuffleSongComposition(raw string) (ShuffleSongComposition, error) {
	switch raw {
	case "off":
		return ShuffleSongCompositionOff, nil
	case "frog":
		return ShuffleSongCompositionFrogs, nil
	case "warp":
		return ShuffleSongCompositionWarp, nil
	case "all":
		return ShuffleSongCompositionAll, nil
	default:
		return 0, fmt.Errorf("unknown shuffle song composition setting %q", raw)
	}
}
