package settings

func Default() Zootr {
	var s Zootr
	s.LogicRules = LogicGlitchess
	s.TriforceHunt = TriforceHunt{}
	s.LacsCondition = CreateLacs(CondVanilla, 0)
	s.BridgeCondition = CreateBridge(CondMedallions, 6)
	s.BlueFireArrows = false
	s.DisabledLocations = nil
	s.FixBrokenDrops = true
	s.FreeBombchuDrops = false
	s.HintsRevealed = HintsRevealedAlways

	s.KeyShuffle.BossKeys = KeysDungeon
	s.KeyShuffle.GanonBKShuffle = GanonBKRemove
	s.KeyShuffle.GanonBKCondition = CreateGanonBK(CondMedallions, 6)
	s.KeyShuffle.HideoutKeys = KeysRemove
	s.KeyShuffle.Keyrings = KeyRingsOff
	s.KeyShuffle.SilverRupeePouches = SilverRupeesOff
	s.KeyShuffle.SilverRupees = KeysVanilla
	s.KeyShuffle.SmallKeys = KeysDungeon

	s.Locations.ReachableLocations = ReachableAll
	s.Locations.KokriForest = KokriForestDekuClosed
	s.Locations.Kakariko = KakGateOpen
	s.Locations.OpenDoorOfTime = true
	s.Locations.ZoraFountain = ZoraFountainClosed
	s.Locations.GerudoFortress = GerudoFortressFast
	s.Locations.Disabled = []string{"Deku Theater Mask of Truth"}

	s.Dungeons.Trials = TrialsEnabledNone
	s.Dungeons.Shortcuts = ShortcutsOff
	s.Dungeons.MasterQuest = MasterQuestDungeonsNone
	s.Dungeons.Rewards = DungeonRewardAnyDungeon
	s.Dungeons.MapsCompasses = MapsCompassesRemove
	s.Dungeons.OneItemPer = false
	s.Dungeons.Completed = CompletedDungeonsNone
	s.Dungeons.ForestTemplePoes = ForestTemplePoesNone

	s.Entrances.Interior = InteriorShuffleOff
	s.Entrances.DungeonEntrances = DungeonEntranceShuffleOff
	s.Entrances.Bosses = BossShuffleOff
	s.Entrances.HideoutEntrances = false
	s.Entrances.Grottos = false
	s.Entrances.Overworld = false
	s.Entrances.RiverExit = false
	s.Entrances.OwlDrops = false
	s.Entrances.WarpSongs = false

	s.Spawns.AdultSpawn = SpawnVanilla
	s.Spawns.ChildSpawn = SpawnVanilla
	s.Spawns.StartingAge = StartAgeRandom

	s.Shuffling.Beans = false
	s.Shuffling.Beehives = false
	s.Shuffling.Cows = false
	s.Shuffling.Crates = ShuffleCratesOff
	s.Shuffling.ExpensiveMerchants = false
	s.Shuffling.Freestandings = ShuffleFreestandingsOff
	s.Shuffling.FrogRupeeRewards = false
	s.Shuffling.GerudoCard = false
	s.Shuffling.KokriSword = true
	s.Shuffling.LoachReward = ShuffleLoachRewardOff
	s.Shuffling.NightTokensWithoutSuns = true
	s.Shuffling.OcarinaNotes = false
	s.Shuffling.Ocarinas = false
	s.Shuffling.Pots = ShufflePotsOff
	s.Shuffling.Scrubs = ShuffleScrubsUpgradeOnly
	s.Shuffling.Shops = ShuffleShopsOff
	s.Shuffling.SongPatterns = ShuffleSongPatternsOff
	s.Shuffling.Songs = ShuffleSongsOnSong
	s.Shuffling.Tokens = ShuffleGoldTokenOff
	s.Shuffling.WonderItems = false

	s.Skills.Tricks = map[string]bool{
		"fewer_tunic_requirements":    true,
		"grottos_without_agony":       true,
		"child_deadhand":              true,
		"man_on_roof":                 true,
		"dc_jump":                     true,
		"rusted_switches":             true,
		"windmill_poh":                true,
		"crater_bean_poh_with_hovers": true,
		"forest_vines":                true,
		"lens_botw":                   true,
		"lens_castle":                 true,
		"lens_gtg":                    true,
		"lens_shadow":                 true,
		"lens_spirit":                 true,
	}

	s.Starting.PlantBeans = false
	s.Starting.Hearts = 3
	s.Starting.RauruReward = true
	s.Starting.Rupees = 0
	s.Starting.TimeOfDay = StartingTimeOfDayDefault
	s.Starting.Tokens = []string{
		"Deku Shield",
		"Ocarina",
		"Zeldas Letter",
	}
	s.Starting.WithConsumables = true

	s.Skips.EponaRace = true
	s.Skips.TowerEscape = true
	s.Skips.ChildZelda = true

	s.Minigames.BigPoeCount = 1
	s.Minigames.CollapsePhases = true
	s.Minigames.KakChickens = 4

	s.Damage.Bonk = BonkDamageNormal
	s.Damage.Multiplier = DamageMultiplierNormal

	s.Trades.Adult = AdultTradeStartClaimCheck
	s.Trades.Child = ChildTradeStartMaskKeaton

	s.ItemPool = ItemPoolDefault
	s.ScarecrowBehavior = ScarecrowBehaviorDefault

	return s
}
