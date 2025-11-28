package oldsettings

import (
	"fmt"
	"math"
)

func (this *Zootr) String(name string) (string, error) {
	var val string
	switch name {
	case "logic_rules":
		val = this.LogicRules.String()
	case "reachable_locations":
		val = this.Locations.ReachableLocations.String()
	case "lacs_condition":
		cond, _ := DecodeCondition(this.LacsCondition)
		val = cond.String()
	case "bridge":
		cond, _ := DecodeCondition(this.BridgeCondition)
		val = cond.String()
	case "shuffle_ganon_bosskey":
		cond, _ := DecodeCondition(this.KeyShuffle.GanonBKCondition)
		val = cond.String()
	case "open_forest":
		val = this.Locations.KokriForest.String()
	case "open_kakariko":
		val = this.Locations.Kakariko.String()
	case "zora_fountain":
		val = this.Locations.ZoraFountain.String()
	case "gerudo_fortress":
		val = this.Locations.GerudoFortress.String()
	case "shuffle_scrubs":
		val = this.Shuffling.Scrubs.String()
	case "shuffle_pots":
		val = this.Shuffling.Pots.String()
	case "shuffle_crates":
		val = this.Shuffling.Crates.String()
	case "shuffle_dungeon_rewards":
		val = this.Dungeons.Rewards.String()
	case "shuffle_tcgkeys":
		val = this.KeyShuffle.TreasureChestGame.String()
	case "hints":
		val = this.HintsRevealed.String()
	case "damage_multiplier":
		val = this.Damage.Multiplier.String()
	case "deadly_bonks":
		val = this.Damage.Bonk.String()
	case "shuffle_gerudo_fortress_heart_piece":
		val = "remove"
	default:
		return val, unknown(name)
	}
	return val, nil
}

func (this *Zootr) Float64(name string) (float64, error) {
	var val float64
	switch name {
	case "triforce_count_per_world":
		val = float64(this.TriforceHunt.CountPerWorld)
	case "triforce_goal_per_world":
		val = float64(this.TriforceHunt.GoalPerWorld)
	case "lacs_medallions":
		qty, isCond := ExpectedCondition(this.LacsCondition, CondMedallions)
		if isCond {
			val = float64(qty)
		} else {
			val = math.MaxFloat64
		}
	case "lacs_stones":
		qty, isCond := ExpectedCondition(this.LacsCondition, CondStones)
		if isCond {
			val = float64(qty)
		} else {
			val = math.MaxFloat64
		}
	case "lacs_rewards":
		qty, isCond := ExpectedCondition(this.LacsCondition, CondRewards)
		if isCond {
			val = float64(qty)
		} else {
			val = math.MaxFloat64
		}
	case "lacs_tokens":
		qty, isCond := ExpectedCondition(this.LacsCondition, CondTokens)
		if isCond {
			val = float64(qty)
		} else {
			val = math.MaxFloat64
		}
	case "lacs_hearts":
		qty, isCond := ExpectedCondition(this.LacsCondition, CondHearts)
		if isCond {
			val = float64(qty)
		} else {
			val = math.MaxFloat64
		}
	case "bridge_medallions":
		qty, isCond := ExpectedCondition(this.BridgeCondition, CondMedallions)
		if isCond {
			val = float64(qty)
		} else {
			val = math.MaxFloat64
		}
	case "bridge_stones":
		qty, isCond := ExpectedCondition(this.BridgeCondition, CondStones)
		if isCond {
			val = float64(qty)
		} else {
			val = math.MaxFloat64
		}
	case "bridge_rewards":
		qty, isCond := ExpectedCondition(this.BridgeCondition, CondRewards)
		if isCond {
			val = float64(qty)
		} else {
			val = math.MaxFloat64
		}
	case "bridge_tokens":
		qty, isCond := ExpectedCondition(this.BridgeCondition, CondTokens)
		if isCond {
			val = float64(qty)
		} else {
			val = math.MaxFloat64
		}
	case "bridge_hearts":
		qty, isCond := ExpectedCondition(this.BridgeCondition, CondHearts)
		if isCond {
			val = float64(qty)
		} else {
			val = math.MaxFloat64
		}
	case "trials":
		val = float64(this.Dungeons.Trials.Count())
	case "ganon_bosskey_medallions":
		qty, isCond := ExpectedCondition(this.KeyShuffle.GanonBKCondition, CondMedallions)
		if isCond {
			val = float64(qty)
		} else {
			val = math.MaxFloat64
		}
	case "ganon_bosskey_stones":
		qty, isCond := ExpectedCondition(this.KeyShuffle.GanonBKCondition, CondStones)
		if isCond {
			val = float64(qty)
		} else {
			val = math.MaxFloat64
		}
	case "ganon_bosskey_rewards":
		qty, isCond := ExpectedCondition(this.KeyShuffle.GanonBKCondition, CondRewards)
		if isCond {
			val = float64(qty)
		} else {
			val = math.MaxFloat64
		}
	case "ganon_bosskey_tokens":
		qty, isCond := ExpectedCondition(this.KeyShuffle.GanonBKCondition, CondTokens)
		if isCond {
			val = float64(qty)
		} else {
			val = math.MaxFloat64
		}
	case "ganon_bosskey_hearts":
		qty, isCond := ExpectedCondition(this.KeyShuffle.GanonBKCondition, CondHearts)
		if isCond {
			val = float64(qty)
		} else {
			val = math.MaxFloat64
		}
	case "chicken_count":
		val = float64(this.Minigames.KakChickens)
	case "big_poe_count":
		val = float64(this.Minigames.BigPoeCount)
	default:
		return val, unknown(name)
	}

	return val, nil
}

func (this *Zootr) Bool(name string) (bool, error) {
	var val bool
	switch name {
	case "trials_random":
		val = this.Dungeons.RandomTrials
	case "triforce_hunt":
		val = this.TriforceHunt.CountPerWorld != 0 && this.TriforceHunt.GoalPerWorld != 0
	case "open_door_of_time":
		val = this.Locations.OpenDoorOfTime
	case "shuffle_hideout_entrances":
		val = this.Entrances.HideoutEntrances
	case "shuffle_grotto_entrances":
		val = this.Entrances.Grottos
	case "shuffle_ganon_tower":
		val = this.Entrances.Tower
	case "shuffle_overworld_entrances":
		val = this.Entrances.Overworld
	case "shuffle_gerudo_valley_river_exit":
		val = this.Entrances.ValleyExit
	case "owl_drops":
		val = this.Entrances.OwlDrops
	case "free_bombchu_drops":
		val = this.FreeBombchuDrops
	case "warp_songs":
		val = this.Entrances.WarpSongs
	case "adult_trade_shuffle":
		val = Has(this.Trades.Adult, AdultTradeShuffle)
	case "shuffle_empty_pots":
		val = this.Shuffling.IncludeEmptyPots
	case "shuffle_empty_crates":
		val = this.Shuffling.IncludeEmptyCrates
	case "shuffle_cows":
		val = this.Shuffling.Cows
	case "shuffle_beehives":
		val = this.Shuffling.Beehives
	case "shuffle_wonderitems":
		val = this.Shuffling.WonderItems
	case "shuffle_kokiri_sword":
		val = this.Shuffling.KokriSword
	case "shuffle_ocarinas":
		val = this.Shuffling.Ocarinas
	case "shuffle_gerudo_card":
		val = this.Shuffling.GerudoCard
	case "shuffle_beans":
		val = this.Shuffling.Beans
	case "shuffle_expensive_merchants":
		val = this.Shuffling.ExpensiveMerchants
	case "shuffle_frog_song_rupees":
		val = this.Shuffling.FrogRupeeRewards
	case "shuffle_individual_ocarina_notes":
		val = this.Shuffling.OcarinaNotes
	case "keyring_give_bk":
		val = Has(this.KeyShuffle.Keyrings, KeyRingsGiveBossKey)
	case "enhance_map_compass":
		val = this.EnhanceMapAndCompass
	case "start_with_consumables":
		val = this.Starting.WithConsumables
	case "start_with_rupees":
		val = this.Starting.Rupees != 0
	case "skip_reward_from_rauru":
		val = this.Starting.RauruReward
	case "no_escape_sequence":
		val = this.Skips.TowerEscape
	case "no_guard_stealth":
		val = this.Skips.HyruleCastleStealth
	case "no_epona_race":
		val = this.Skips.EponaRace
	case "skip_some_minigame_phases":
		val = this.Minigames.CollapsePhases
	case "skip_child_zelda":
		val = this.Skips.ChildZelda
	case "complete_mask_quest":
		val = this.Starting.CompleteMaskQuest
	case "useful_cutscenes":
		val = this.UsefulCutscenes
	case "fast_chests":
		val = this.FastChests
	case "free_scarecrow":
		val = this.Starting.Scarecrow
	case "plant_beans":
		val = this.Starting.PlantBeans
	case "easier_fire_arrow_entry":
		val = this.Tricks.ShadowFireArrowEntry != 0
	case "ruto_already_f1_jabu":
		val = this.Skips.RutoAlreadyOnFloor1
	case "chicken_count_random":
		val = this.Minigames.KakChickens != 0xFF
	case "clearer_hints":
		val = this.ClearerHints
	case "blue_fire_arrows":
		val = this.BlueFireArrows
	case "fix_broken_drops":
		val = this.FixBrokenDrops
	case "tcg_requires_lens":
		val = this.Minigames.TreasureChestGameRequiresLens
	case "no_collectible_hearts":
		val = this.NoCollectibleHearts
	case "one_item_per_dungeon":
		val = this.Dungeons.OneItemPer
	case "shuffle_interior_entrances":
		val = this.Entrances.Interior != InteriorShuffleOff
	case "shuffle_silver_rupees":
		val = this.KeyShuffle.SilverRupees != KeysVanilla
	default:
		return val, unknown(name)
	}

	return val, nil
}

func unknown(name string) error {
	return fmt.Errorf("%q is not a known setting", name)
}
