package optimizer

import (
	"fmt"
	"sudonters/libzootr/internal"
	"sudonters/libzootr/internal/settings"
	"sudonters/libzootr/mido/ast"
	"sudonters/libzootr/mido/symbols"
)

type reader func(*settings.Zootr, string) ast.Node

func InlineSettings(these *settings.Zootr, symbols *symbols.Table) ast.Rewriter {
	inliner := newinliner(these, symbols)
	return ast.Rewriter{
		Identifier: inliner.Identifier,
	}
}

func str(these *settings.Zootr, name string) ast.Node {
	value, err := these.String(name)
	internal.PanicOnError(err)
	return ast.String(value)
}

func f64(these *settings.Zootr, name string) ast.Node {
	value, err := these.Float64(name)
	internal.PanicOnError(err)
	return ast.Number(value)
}

func boolean(these *settings.Zootr, name string) ast.Node {
	value, err := these.Bool(name)
	internal.PanicOnError(err)
	return ast.Boolean(value)
}

type settinginline struct {
	these   *settings.Zootr
	symbols *symbols.Table
	readers map[string]reader
}

func (this settinginline) Identifier(node ast.Identifier, _ ast.Rewriting) (ast.Node, error) {
	symbol := this.symbols.LookUpByIndex(node.AsIndex())
	if symbol == nil {
		return node, nil
	}

	reader, exists := this.readers[symbol.Name]
	if !exists {
		return node, nil
	}

	if symbol.Kind != symbols.SETTING {
		panic(fmt.Errorf("found setting reader for non-setting %q", symbol.Name))
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("panicked while handling %q\n", symbol.Name)
			if err, ok := r.(error); ok {
				fmt.Println(err)
			} else if str, ok := r.(string); ok {
				fmt.Println(str)
			}
			panic(r)
		}
	}()

	result := reader(this.these, symbol.Name)
	return result, nil
}

func newinliner(these *settings.Zootr, symbols *symbols.Table) settinginline {
	var inliner settinginline
	inliner.symbols = symbols
	inliner.these = these
	inliner.readers = map[string]reader{
		"logic_rules":                         str,
		"reachable_locations":                 str,
		"lacs_condition":                      str,
		"bridge":                              str,
		"shuffle_ganon_bosskey":               str,
		"open_forest":                         str,
		"open_kakariko":                       str,
		"zora_fountain":                       str,
		"gerudo_fortress":                     str,
		"shuffle_scrubs":                      str,
		"shuffle_pots":                        str,
		"shuffle_crates":                      str,
		"shuffle_dungeon_rewards":             str,
		"shuffle_tcgkeys":                     str,
		"hints":                               str,
		"damage_multiplier":                   str,
		"deadly_bonks":                        str,
		"shuffle_gerudo_fortress_heart_piece": str,
		"triforce_count_per_world":            f64,
		"triforce_goal_per_world":             f64,
		"lacs_medallions":                     f64,
		"lacs_stones":                         f64,
		"lacs_rewards":                        f64,
		"lacs_tokens":                         f64,
		"lacs_hearts":                         f64,
		"bridge_medallions":                   f64,
		"bridge_stones":                       f64,
		"bridge_rewards":                      f64,
		"bridge_tokens":                       f64,
		"bridge_hearts":                       f64,
		"trials":                              f64,
		"ganon_bosskey_medallions":            f64,
		"ganon_bosskey_stones":                f64,
		"ganon_bosskey_rewards":               f64,
		"ganon_bosskey_tokens":                f64,
		"ganon_bosskey_hearts":                f64,
		"chicken_count":                       f64,
		"big_poe_count":                       f64,
		"trials_random":                       boolean,
		"triforce_hunt":                       boolean,
		"open_door_of_time":                   boolean,
		"shuffle_hideout_entrances":           boolean,
		"shuffle_grotto_entrances":            boolean,
		"shuffle_ganon_tower":                 boolean,
		"shuffle_overworld_entrances":         boolean,
		"shuffle_gerudo_valley_river_exit":    boolean,
		"owl_drops":                           boolean,
		"free_bombchu_drops":                  boolean,
		"warp_songs":                          boolean,
		"adult_trade_shuffle":                 boolean,
		"shuffle_empty_pots":                  boolean,
		"shuffle_empty_crates":                boolean,
		"shuffle_cows":                        boolean,
		"shuffle_beehives":                    boolean,
		"shuffle_wonderitems":                 boolean,
		"shuffle_kokiri_sword":                boolean,
		"shuffle_ocarinas":                    boolean,
		"shuffle_gerudo_card":                 boolean,
		"shuffle_beans":                       boolean,
		"shuffle_expensive_merchants":         boolean,
		"shuffle_frog_song_rupees":            boolean,
		"shuffle_individual_ocarina_notes":    boolean,
		"keyring_give_bk":                     boolean,
		"enhance_map_compass":                 boolean,
		"start_with_consumables":              boolean,
		"start_with_rupees":                   boolean,
		"skip_reward_from_rauru":              boolean,
		"skip_child_zelda":                    boolean,
		"no_escape_sequence":                  boolean,
		"no_guard_stealth":                    boolean,
		"no_epona_race":                       boolean,
		"skip_some_minigame_phases":           boolean,
		"complete_mask_quest":                 boolean,
		"useful_cutscenes":                    boolean,
		"fast_chests":                         boolean,
		"free_scarecrow":                      boolean,
		"plant_beans":                         boolean,
		"easier_fire_arrow_entry":             boolean,
		"ruto_already_f1_jabu":                boolean,
		"chicken_count_random":                boolean,
		"clearer_hints":                       boolean,
		"blue_fire_arrows":                    boolean,
		"fix_broken_drops":                    boolean,
		"tcg_requires_lens":                   boolean,
		"no_collectible_hearts":               boolean,
		"one_item_per_dungeon":                boolean,
		"shuffle_interior_entrances":          boolean,
		"shuffle_silver_rupees":               boolean,
	}

	return inliner
}
