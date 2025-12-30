package boot

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sudonters/libzootr/components"
	"sudonters/libzootr/internal/json"
	"sudonters/libzootr/magicbean/tracking"
	"sudonters/libzootr/settings"

	"github.com/etc-sudonters/substrate/slipup"
)

func malformed(err error) error {
	return fmt.Errorf("malformed spoilers document: %w", err)
}

func LoadSpoilerData(
	ctx context.Context,
	r io.Reader,
	these *settings.Model,
	nodes *tracking.Nodes,
	tokens *tracking.Tokens) error {

	reader := json.NewParser(json.NewScanner(r))
	obj, err := reader.ReadObject()
	if err != nil {
		return malformed(err)
	}
	var property string
	var propertyErr error

	defer func() {
		if recovered := recover(); recovered != nil {
			slog.ErrorContext(ctx, "failed to import spoiler data", "property", property, "parser", reader.Dump())
			panic(recovered)
		}
	}()

	for obj.More() {
		property, propertyErr = obj.ReadPropertyName()
		if propertyErr != nil {
			return malformed(propertyErr)
		}

		slog.DebugContext(ctx, "importing from spoiler log", "kind", property)

		switch property {
		case "settings", "randomized_settings":
			obj, err := obj.ReadObject()
			if err != nil {
				return malformed(err)
			}
			if err := CopySettings(these, obj); err != nil {
				return malformed(err)
			}
		case "locations":
			obj, err := obj.ReadObject()
			if err != nil {
				return malformed(err)
			}
			if err := CopyPlacements(tokens, nodes, obj); err != nil {
				return malformed(err)
			}
		case ":skipped_locations":
			obj, err := obj.ReadObject()
			if err != nil {
				return malformed(err)
			}
			if err := CopySkippedLocations(nodes, obj); err != nil {
				return malformed(err)
			}
		default:
			if err := obj.DiscardValue(); err != nil {
				return malformed(err)
			}
		}
	}

	if err := obj.ReadEnd(); err != nil {
		return malformed(err)
	}

	return nil
}

func CopySettings(these *settings.Model, obj *json.ObjectParser) error {
	for obj.More() {
		setting, err := obj.ReadPropertyName()
		if err != nil {
			return err
		}
		slog.Debug("reading setting property", "name", setting)
		if err := readSetting(setting, these, obj); err != nil {
			slog.Error("failed to read property", "name", setting, "error", err)
			return invalidSettingFor(setting, err)
		}
	}
	return obj.ReadEnd()
}

func parseConditionInto(r json.Reader, dest *settings.ConditionedAmount) error {
	cond, err := json.ParseStringWith(r, settings.ParseCondition)
	if err == nil {
		_, qty := dest.Decode()
		*dest = settings.EncodeConditionedAmount(cond, qty)
	}
	return err

}

func parseConditionedAmountInto(
	r json.Reader,
	unparsedCond string,
	dest *settings.ConditionedAmount,
) error {
	qty, qtyErr := r.ReadInt()
	if qtyErr != nil {
		return qtyErr
	}

	cond, condErr := settings.ConditionFrom(unparsedCond, uint32(qty))
	if condErr == nil {
		*dest = cond
	}
	return condErr

}

func readBoolIntoFlags[F ~uint64](r json.Reader, dest *F, set F) error {
	b, err := r.ReadBool()
	if err == nil && b {
		*dest |= set
	}
	return err
}

func parsePrefixedConditionInto(
	r json.Reader,
	prefix string,
	unparsedCond string,
	dest *settings.ConditionedAmount,
) error {
	condStr, _ := strings.CutPrefix(unparsedCond, prefix)
	return parseConditionedAmountInto(r, condStr, dest)

}

func readSetting(propertyName string, these *settings.Model, obj *json.ObjectParser) error {
	var err error

	switch propertyName {
	case "logic_rules":
		return json.ParseStringInto(obj, &these.Logic.Set, settings.ParseLogic)
	case "reachable_locations":
		return json.ParseStringInto(obj, &these.Logic.Locations.Reachability, settings.ParseReachability)
	case "triforce_hunt":
		return json.ReadBoolInto(obj, &these.Logic.WinConditions.TriforceHunt)
	case "triforce_goal_per_world":
		return json.ReadIntInto(obj, &these.Logic.WinConditions.TriforceGoal)
	case "triforce_count_per_world":
		return json.ReadIntInto(obj, &these.Logic.WinConditions.TriforceCount)
	case "lacs_condition":
		return parseConditionInto(obj, &these.Logic.WinConditions.Lacs)
	case "lacs_stones",
		"lacs_medallions",
		"lacs_rewards",
		"lacs_tokens",
		"lacs_hearts":
		if these.Logic.WinConditions.Lacs.Amount() != 0 {
			return errors.New("light arrow cutscene condition already set")
		}
		return parsePrefixedConditionInto(obj, "lacs_", propertyName, &these.Logic.WinConditions.Lacs)
	case "bridge":
		return parseConditionInto(obj, &these.Logic.WinConditions.Bridge)
	case "bridge_stones",
		"bridge_medallions",
		"bridge_rewards",
		"bridge_tokens",
		"bridge_hearts":
		return parsePrefixedConditionInto(
			obj,
			"bridge_",
			propertyName,
			&these.Logic.WinConditions.Bridge,
		)
	case "trials":
		return json.ReadIntInto(obj, &these.Generation.TrialCount)
	case "shuffle_ganon_bosskey":
		return json.ParseStringInto(obj, &these.Logic.Dungeon.GanonBossKeyShuffle, settings.ParseGanonBossKeyShuffle)
	case "ganon_bosskey_medallions",
		"ganon_bosskey_stones",
		"ganon_bosskey_rewards",
		"ganon_bosskey_hearts",
		"ganon_bosskey_tokens":
		if these.Logic.WinConditions.GanonBossKey != 0 {
			return fmt.Errorf("ganon_bosskey condition is already set")
		}
		return parsePrefixedConditionInto(
			obj,
			"ganon_bosskey_",
			propertyName,
			&these.Logic.WinConditions.GanonBossKey,
		)
	case "shuffle_dungeon_rewards":
		return json.ParseStringInto(obj, &these.Logic.Shuffling.DungeonRewards, settings.ParseShuffleDungeonReward)
	case "shuffle_bosskeys":
		return json.ParseStringInto(obj, &these.Logic.Dungeon.BossKey, settings.ParseShuffleKeys)
	case "shuffle_smallkeys":
		return json.ParseStringInto(obj, &these.Logic.Dungeon.Keys, settings.ParseShuffleKeys)
	case "shuffle_hideoutkeys":
		return json.ParseStringInto(obj, &these.Logic.Dungeon.GerudoFortressKeys, settings.ParseShuffleKeys)
	case "shuffle_tcgkeys":
		return json.ParseStringInto(obj, &these.Logic.Minigames.TreasureChestGameKeys, settings.ParseShuffleKeys)
	case "shuffle_silver_rupees":
		return json.ParseStringInto(obj, &these.Logic.Dungeon.SilverRupees, settings.ParseShuffleKeys)
	case "shuffle_mapcompass":
		return json.ParseStringInto(obj, &these.Logic.Dungeon.MapCompass, settings.ParseMapCompass)
	case "open_forest":
		return json.ParseStringInto(obj, &these.Logic.Connections.OpenKokriForest, settings.ParseOpenForest)
	case "open_kakariko":
		return json.ParseStringInto(obj, &these.Logic.Connections.OpenKakarikoGate, settings.ParseKakarikoGate)
	case "zora_fountain":
		return json.ParseStringInto(obj, &these.Logic.Connections.OpenZoraFountain, settings.ParseOpenZoraFountain)
	case "gerudo_fortress":
		return json.ParseStringInto(obj, &these.Logic.Dungeon.GerudoFortress, settings.ParseGerudoFortressCarpenterRescue)
	case "starting_age":
		val, err := obj.ReadString()
		if err != nil {
			return invalidSettingFor(propertyName, err)
		}
		switch val {
		case "adult":
			these.Logic.Spawns.StartAge = settings.StartAgeAdult
		case "child":
			these.Logic.Spawns.StartAge = settings.StartAgeChild
		case "random":
			these.Generation.RandomStartingAge = true
		default:
			return fmt.Errorf("unknown start age: %q", val)
		}
	case "free_bombchu_drops":
		return json.ReadBoolInto(obj, &these.Logic.FreeBombchuDrops)
	case "one_item_per_dungeon":
		return json.ReadBoolInto(obj, &these.Logic.Dungeon.OneMajorItemPerDungeon)
	case "shuffle_song_items":
		return json.ParseStringInto(obj, &these.Logic.Shuffling.Songs, settings.ParseShuffleSong)
	case "shuffle_freestanding_items":
		return json.ParseStringInto(obj, &these.Logic.Shuffling.Freestandings, settings.ParsePartitionedShuffle)
	case "shuffle_pots":
		return json.ParseStringInto(obj, &these.Logic.Shuffling.Pots, settings.ParsePartitionedShuffle)
	case "shuffle_empty_pots":
		return readBoolIntoFlags(obj, &these.Logic.Shuffling.Flags, settings.ShuffleEmptyPots)
	case "shuffle_crates":
		return json.ParseStringInto(obj, &these.Logic.Shuffling.Crates, settings.ParsePartitionedShuffle)
	case "shuffle_empty_crates":
		return readBoolIntoFlags(obj, &these.Logic.Shuffling.Flags, settings.ShuffleEmptyCrates)
	case "shuffle_cows":
		return readBoolIntoFlags(obj, &these.Logic.Shuffling.Flags, settings.ShuffleCows)
	case "shuffle_beehives":
		return readBoolIntoFlags(obj, &these.Logic.Shuffling.Flags, settings.ShuffleBeehives)
	case "shuffle_wonderitems":
		return readBoolIntoFlags(obj, &these.Logic.Shuffling.Flags, settings.ShuffleWonderItems)
	case "shuffle_kokri_sword":
		return readBoolIntoFlags(obj, &these.Logic.Shuffling.Flags, settings.ShuffleKokiriSword)
	case "shuffle_ocarinas":
		return readBoolIntoFlags(obj, &these.Logic.Shuffling.Flags, settings.ShuffleOcarinas)
	case "shuffle_gerudo_card":
		return readBoolIntoFlags(obj, &these.Logic.Shuffling.Flags, settings.ShuffleGerudoCard)
	case "shuffle_beans":
		return readBoolIntoFlags(obj, &these.Logic.Shuffling.Flags, settings.ShuffleBeans)
	case "shuffle_expensive_merchants":
		return readBoolIntoFlags(obj, &these.Logic.Shuffling.Flags, settings.ShuffleExpensiveMerchants)
	case "shuffle_frog_song_rupees":
		return readBoolIntoFlags(obj, &these.Logic.Shuffling.Flags, settings.ShuffleFrogRupees)
	case "shuffle_individual_ocarina_notes":
		return readBoolIntoFlags(obj, &these.Logic.Shuffling.Flags, settings.ShuffleOcarinaNotes)
	case "warp_songs":
		return readBoolIntoFlags(obj, &these.Logic.Connections.Flags, settings.ConnectionShuffleWarpSongDestinations)
	case "shuffle_loach_reward":
		return json.ParseStringInto(obj, &these.Logic.Shuffling.Loach, settings.ParseShuffleLoachReward)
	case "shopsanity":
		return json.ParseStringInto(obj, &these.Logic.Shuffling.Shops, settings.ParseShuffleShop)
	case "shopsanity_prices":
		return json.ParseStringInto(obj, &these.Logic.Shuffling.ShopPrices, settings.ParseShuffleShopPrices)
	case "tokensanity":
		return json.ParseStringInto(obj, &these.Logic.Shuffling.SkullTokens, settings.ParsePartitionedShuffle)
	case "disabled_locations":
		return json.ReadStringArrayInto(obj, &these.Logic.Locations.Disabled)
	case "allowed_tricks":
		var err error
		if these.Logic.Tricks == nil {
			these.Logic.Tricks = make(map[string]bool, 16)
		}
		for _, enabled := range json.ReadStringArray(obj, &err) {
			trick, _ := strings.CutPrefix(enabled, "logic_")
			these.Logic.Tricks[trick] = true
		}
		return err
	case "starting_items":
		if these.Logic.Spawns.Items == nil {
			these.Logic.Spawns.Items = make(map[string]int, 8)
		}
		return json.ReadIntObjectInto(obj, these.Logic.Spawns.Items)
	case "start_with_consumables":
		return json.ReadBoolInto(obj, &these.Generation.StartWithConsumables)
	case "start_with_rupees":
		return json.ReadBoolInto(obj, &these.Generation.StartWithRupees)
	case "starting_hearts":
		return json.ReadIntInto(obj, &these.Logic.Spawns.Hearts)
	case "skip_child_zelda":
		return readBoolIntoFlags(obj, &these.Logic.Locations.Flags, settings.LocationSkipChildZelda)
	case "skip_reward_from_rauru":
		return readBoolIntoFlags(obj, &these.Logic.Locations.Flags, settings.LocationSkipRauruReward)
	case "no_escape_sequence":
		return json.ReadBoolInto(obj, &these.Generation.SkipTowerEscape)
	case "no_guard_stealth":
		return json.ReadBoolInto(obj, &these.Generation.SkipCastleStealth)
	case "no_epona_race":
		return json.ReadBoolInto(obj, &these.Generation.SkipEponaRace)
	case "skip_some_minigame_phases":
		return json.ReadBoolInto(obj, &these.Generation.SkipSomeMinigamePhases)
	case "complete_mask_quest":
		return readBoolIntoFlags(obj, &these.Logic.Locations.Flags, settings.LocationsCompleteMaskQuest)
	case "useful_cutscenes":
		return json.ReadBoolInto(obj, &these.Generation.KeepGlitchUsefulCutscenes)
	case "fast_chests":
		return json.ReadBoolInto(obj, &these.Generation.FastChests)
	case "free_scarecrow":
		return readBoolIntoFlags(obj, &these.Logic.Locations.Flags, settings.LocationsFreeScarecrow)
	case "fast_bunny_hood":
		return json.ReadBoolInto(obj, &these.Generation.FastBunnyHood)
	case "auto_equip_masks":
		return json.ReadBoolInto(obj, &these.Generation.AutoEquipMasks)
	case "plant_beans":
		return readBoolIntoFlags(obj, &these.Logic.Locations.Flags, settings.LocationsPlantBeans)
	case "chicken_count_random":
		return json.ReadBoolInto(obj, &these.Generation.RandomChickenCount)
	case "chicken_count":
		return json.ReadIntInto(obj, &these.Logic.Minigames.KakarikoChickenGoal)
	case "big_poe_count_random":
		return json.ReadBoolInto(obj, &these.Generation.RandomPoeCount)
	case "big_poe_count":
		return json.ReadIntInto(obj, &these.Logic.Minigames.BigPoeGoal)
	case "easier_fire_arrow_entry":
		return json.ReadBoolInto(obj, &these.Generation.EasierFireArrowEntry)
	case "ruto_already_f1_jabu":
		return json.ReadBoolInto(obj, &these.Generation.RutoAlreadyOnFloor1)
	case "ocarina_songs":
		return parseShuffledOcarinaSongs(obj, these)
	case "damage_multiplier":
		return json.ParseStringInto(obj, &these.Logic.Damage.Multiplier, settings.ParseDamageMultiplier)
	case "deadly_bonks":
		return json.ParseStringInto(obj, &these.Logic.Damage.Bonks, settings.ParseDamageMultiplier)
	case "no_collectible_hearts":
		return json.ReadBoolInto(obj, &these.Logic.Shuffling.RemoveCollectibleHearts)
	case "starting_tod":
		str, err := obj.ReadString()
		if err != nil {
			return err
		}
		if str == "random" {
			these.Generation.RandomStartTimeOfDay = true
			return nil
		}
		tod, err := settings.ParseTimeOfDay(str)
		if err == nil {
			these.Logic.Spawns.TimeOfDay = tod
		}
		return err
	case "blue_fire_arrows":
		return json.ReadBoolInto(obj, &these.Logic.BlueFireArrows)
	case "fix_broken_drops":
		return json.ReadBoolInto(obj, &these.Logic.FixBrokenDrops)
	case "shuffle_child_trade":
		var childTradeItems settings.ChildTradeItems
		for _, itemName := range json.ReadStringArray(obj, &err) {
			item, err := settings.ParseChildTradeItem(itemName)
			if err != nil {
				return err
			}
			childTradeItems |= item
		}
		these.Logic.Trade.ChildItems = childTradeItems
		return err
	case "adult_trade_shuffle":
		return json.ReadBoolInto(obj, &these.Logic.Trade.AdultTradeShuffle)
	case "adult_trade_start":
		var adultTradeItems settings.AdultTradeItems
		for _, itemName := range json.ReadStringArray(obj, &err) {
			item, err := settings.ParseAdultTradeItem(itemName)
			if err != nil {
				return err
			}
			adultTradeItems |= item
		}
		these.Logic.Trade.AdultItems = adultTradeItems
		return err
	case "scarecrow_behavior":
		return json.ParseStringInto(obj, &these.Logic.Scarecrow, settings.ParseScarecrownBehavior)
	case "correct_chest_appearances", "chest_textures_specific",
		"minor_items_as_major_chest", "correct_potcrate_appearances",
		"key_appearance_match_dungeon", "clearer_hints", "hints", "hint_dist",
		"item_hints", "hint_dist_user", "misc_hints", "text_shuffle",
		"item_pool_value", "junk_ice_traps", "ice_trap_appearance",
		"dungeon_shortcuts", "key_rings", "mq_dungeons_count", "mq_dungeons_mode",
		"mq_dungeons_specific", "empty_dungeons_mode", "empty_dungeons_count",
		"empty_dungeons_specific", "empty_dungeons_rewards",
		"shuffle_interior_entrances", "shuffle_hideout_entrances",
		"shuffle_gerudo_fortress_heart_piece", "shuffle_grotto_entrances",
		"shuffle_dungeon_entrances", "shuffle_bosses", "shuffle_ganon_tower",
		"shuffle_overworld_entrances", "shuffle_gerudo_valley_river_exit",
		"owl_drops", "spawn_positions":
		slog.Debug("discarding spoiler setting", "name", propertyName)
		return obj.DiscardValue()

	case "user_message":
		slog.Debug("not importing setting", "name", propertyName)
		obj.DiscardValue()
	default:
		slog.Debug("discarding spoiler setting", "name", propertyName)
		return obj.DiscardValue()
	}

	return nil
}

func CopyPlacements(tokens *tracking.Tokens, nodes *tracking.Nodes, obj *json.ObjectParser) error {
	var err error

	for node, placed := range json.ReadObjectProperties(obj, readPlacedItem, &err) {
		node := nodes.Placement(components.Name(node))
		token := tokens.MustGet(components.Name(placed.name))
		if placed.price > -1 {
			token.Attach(components.Price(placed.price))
		}
		node.Holding(token)
	}
	return err
}

func CopySkippedLocations(nodes *tracking.Nodes, obj *json.ObjectParser) error {
	var err error
	for skipped := range obj.Keys(&err) {
		node := nodes.Placement(components.Name(skipped))
		node.Attach(components.Skipped{})
	}

	return err
}

type placed struct {
	name  string
	price int
}

func readPlacedItem(obj *json.ObjectParser) (placed, error) {
	var placed placed
	placed.price = -1
	switch kind := obj.Current().Kind; kind {
	case json.STRING:
		var readErr error
		placed.name, readErr = obj.ReadString()
		if readErr != nil {
			return placed, readErr
		}
		break
	case json.OBJ_OPEN:
		inner, readErr := obj.ReadObject()
		if readErr != nil {
			return placed, readErr
		}
		for inner.More() {
			prop, err := obj.ReadPropertyName()
			if err != nil {
				return placed, err
			}
			switch prop {
			case "item":
				placed.name, readErr = obj.ReadString()
				if readErr != nil {
					return placed, err
				}
				break
			case "price":
				placed.price, readErr = obj.ReadInt()
				if readErr != nil {
					return placed, readErr
				}
				break
			default:
				return placed, fmt.Errorf("unexpected property: %q", prop)
			}
		}

	}
	return placed, nil
}

func invalidSettingFor(name string, cause error) error {
	return fmt.Errorf("invalid value for %q: %w", name, cause)
}

func parseShuffledOcarinaSongs(obj *json.ObjectParser, these *settings.Model) error {
	switch obj.Current().Kind {
	case json.STRING:
		slog.Debug("parsing ocarina_songs as string")
		return json.ParseStringInto(obj, &these.Logic.Shuffling.SongComposition, settings.ParseShuffleSongComposition)
	case json.ARR_OPEN:
		slog.Debug("parsing ocarina_songs as array")
		result, err := json.ReduceStringArrayInto(
			obj,
			these.Logic.Shuffling.SongComposition,
			func(current settings.ShuffleSongComposition, input string) (settings.ShuffleSongComposition, error) {
				val, err := settings.ParseShuffleSongComposition(input)
				if err != nil {
					return current, err
				}
				return current | val, nil
			},
		)
		these.Logic.Shuffling.SongComposition = result
		return err
	default:
		return slipup.Describe(obj.Unexpected(), "expected ocarina_songs to be a string or array of strings")
	}
}
