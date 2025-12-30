package settings

import (
	"fmt"
	"math"
)

type Reader struct {
	*Model
}

func unknown(name string) error {
	return fmt.Errorf("%q is not a known setting", name)
}

func (this Reader) String(name string) (string, error) {
	var val string
	switch name {
	case "logic_rules":
		val = this.Logic.Set.String()
	case "reachable_locations":
		val = this.Logic.Locations.Reachability.String()
	case "lacs_condition":
		cond := this.Logic.WinConditions.Lacs.Kind()
		val = cond.String()
	case "bridge":
		cond := this.Logic.WinConditions.Bridge.Kind()
		val = cond.String()
	case "shuffle_ganon_bosskey":
		val = this.Logic.Dungeon.GanonBossKeyShuffle.String()
	case "open_forest":
		val = this.Logic.Connections.OpenKokriForest.String()
	case "open_kakariko":
		val = this.Logic.Connections.OpenKakarikoGate.String()
	case "zora_fountain":
		val = this.Logic.Connections.OpenKakarikoGate.String()
	case "gerudo_fortress":
		val = this.Logic.Dungeon.GerudoFortress.String()
	case "shuffle_scrubs":
		val = this.Logic.Shuffling.Scrubs.String()
	case "shuffle_pots":
		shuffle := PartitionedShuffle(this.Logic.Shuffling.Pots)
		val = shuffle.String()
	case "shuffle_crates":
		shuffle := PartitionedShuffle(this.Logic.Shuffling.Crates)
		val = shuffle.String()
	case "shuffle_dungeon_rewards":
		val = this.Logic.Shuffling.DungeonRewards.String()
	case "shuffle_tcgkeys":
		val = this.Logic.Minigames.TreasureChestGameKeys.String()
	case "hints":
		val = this.Logic.HintsRevealed.String()
	case "damage_multiplier":
		val = this.Logic.Damage.Multiplier.String()
	case "deadly_bonks":
		val = this.Logic.Damage.Bonks.String()
	case "shuffle_gerudo_fortress_heart_piece":
		val = "remove"
	case "scarecrow_behavior":
		val = this.Logic.Scarecrow.String()
	default:
		return val, unknown(name)
	}
	return val, nil

}

func expectCondition(cond ConditionedAmount, expect ConditionKind) float64 {
	kind, qty := cond.Decode()
	if kind != expect {
		return math.MaxFloat64
	}
	return float64(qty)
}

func (this Reader) Number(name string) (float64, error) {
	var val float64
	var err error
	switch name {
	case "triforce_count_per_world":
		val = float64(this.Logic.WinConditions.TriforceCount)
	case "triforce_goal_per_world":
		val = float64(this.Logic.WinConditions.TriforceGoal)
	case "lacs_medallions":
		val = expectCondition(this.Logic.WinConditions.Lacs, CondMedallions)
	case "lacs_stones":
		val = expectCondition(this.Logic.WinConditions.Lacs, CondStones)
	case "lacs_rewards":
		val = expectCondition(this.Logic.WinConditions.Lacs, CondRewards)
	case "lacs_tokens":
		val = expectCondition(this.Logic.WinConditions.Lacs, CondTokens)
	case "lacs_hearts":
		val = expectCondition(this.Logic.WinConditions.Lacs, CondHearts)
	case "bridge_medallions":
		val = expectCondition(this.Logic.WinConditions.Lacs, CondMedallions)
	case "bridge_stones":
		val = expectCondition(this.Logic.WinConditions.Lacs, CondStones)
	case "bridge_rewards":
		val = expectCondition(this.Logic.WinConditions.Lacs, CondRewards)
	case "bridge_tokens":
		val = expectCondition(this.Logic.WinConditions.Lacs, CondTokens)
	case "bridge_hearts":
		val = expectCondition(this.Logic.WinConditions.Lacs, CondHearts)
	case "trials":
		val = float64(CountFlags(this.Logic.WinConditions.Trials))
	case "ganon_bosskey_medallions":
		val = expectCondition(this.Logic.WinConditions.Lacs, CondMedallions)
	case "ganon_bosskey_stones":
		val = expectCondition(this.Logic.WinConditions.Lacs, CondStones)
	case "ganon_bosskey_rewards":
		val = expectCondition(this.Logic.WinConditions.Lacs, CondRewards)
	case "ganon_bosskey_tokens":
		val = expectCondition(this.Logic.WinConditions.Lacs, CondTokens)
	case "ganon_bosskey_hearts":
		val = expectCondition(this.Logic.WinConditions.Lacs, CondHearts)
	case "chicken_count":
		val = float64(this.Logic.Minigames.KakarikoChickenGoal)
	case "big_poe_count":
		val = float64(this.Logic.Minigames.BigPoeGoal)
	default:
		return val, unknown(name)
	}

	return val, err
}

func (this Reader) Bool(name string) (bool, error) {
	var val bool
	switch name {
	case "triforce_hunt":
		val = this.Logic.WinConditions.TriforceHunt
	case "open_door_of_time":
		val = HasFlag(this.Logic.Connections.Flags, ConnectionOpenDoorOfTime)
	case "free_bombchu_drops":
		val = this.Logic.FreeBombchuDrops
	case "shuffle_individual_ocarina_notes":
		val = HasFlag(this.Logic.Shuffling.Flags, ShuffleOcarinaNotes)
	case "shuffle_silver_rupees":
		val = this.Logic.Dungeon.SilverRupees != ShuffleKeysVanilla
	case "skip_child_zelda":
		val = HasFlag(this.Logic.Locations.Flags, LocationSkipChildZelda)
	case "skip_reward_from_rauru":
		val = HasFlag(this.Logic.Locations.Flags, LocationSkipRauruReward)
	case "shuffle_hideout_entrances":
		val = HasFlag(this.Logic.Connections.Flags, ConnectionShuffleHideoutEntrances)
	case "plant_beans":
		val = HasFlag(this.Logic.Locations.Flags, LocationsPlantBeans)
	case "adult_trade_shuffle":
		val = this.Logic.Trade.AdultItems != 0
	case "free_scarecrow":
		val = HasFlag(this.Logic.Locations.Flags, LocationsFreeScarecrow)
	case "shuffle_interior_entrances":
		val = this.Logic.Connections.Interior != 0
	case "shuffle_expensive_merchants":
		val = HasFlag(this.Logic.Shuffling.Flags, ShuffleExpensiveMerchants)
	case "shuffle_overworld_entrances":
		val = this.Logic.Connections.Overworld != 0
	case "complete_mask_quest":
		val = HasFlag(this.Logic.Locations.Flags, LocationsCompleteMaskQuest)
	case "shuffle_empty_pots":
		val = HasFlag(this.Logic.Shuffling.Flags, ShuffleEmptyPots)
	case "fix_broken_drops":
		val = this.Logic.FixBrokenDrops
	case "warp_songs":
		val = HasFlag(this.Logic.Connections.Flags, ConnectionShuffleWarpSongDestinations)
	case "disable_trade_revert":
		//self.shuffle_interior_entrances or settings.shuffle_overworld_entrances or settings.adult_trade_shuffle
		val = this.Logic.Connections.Interior != 0 || this.Logic.Connections.Overworld != 0 || this.Logic.Trade.AdultTradeShuffle
	case "blue_fire_arrows":
		val = this.Logic.BlueFireArrows
	default:
		return val, unknown(name)
	}

	return val, nil
}
