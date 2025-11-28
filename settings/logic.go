package settings

type Logic struct {
	Set              LogicSetting
	BlueFireArrows   bool
	FixBrokenDrops   bool
	HintsRevealed    HintsRevealed
	FreeBombchuDrops bool

	Spawns struct {
		TimeOfDay  TimeOfDay
		AdultSpawn string
		ChildSpawn string
		StartAge   StartAge
		Hearts     int
		Items      map[string]int
	}

	Shuffling struct {
		DungeonRewards ShuffleDungeonRewards
		Songs          ShuffleSongs
		Shops          ShuffleShop
		ShopPrices     ShuffleShopPrices
		SkullTokens    PartitionedShuffle
		Scrubs         ShuffleScrub
		Freestandings  PartitionedShuffle
		Pots           PartitionedShuffle
		Crates         PartitionedShuffle
		Loach          ShuffleLoachReward

		SongComposition ShuffleSongComposition

		RemoveCollectibleHearts bool

		Flags ShufflingFlags
	}

	Damage struct {
		Multiplier DamageMultiplier
		Bonks      DamageMultiplier
	}

	Locations struct {
		Reachability LocationsReachable
		Flags        LocationFlags
		Disabled     []string
	}

	Trade struct {
		AdultItems AdultTradeItems
		ChildItems ChildTradeItems

		AdultTradeShuffle bool
	}

	Dungeon struct {
		OneMajorItemPerDungeon bool
		Bosses                 BossShuffle
		Keys                   ShuffleKeys
		BossKey                ShuffleKeys
		SilverRupees           ShuffleKeys
		KeyRings               bool
		KeyRingsIncludeBossKey bool
		SilverRupeePouches     bool
		MapCompass             ShuffleMapCompass
		Empty                  Flags
		MasterQuest            Flags
		Shortcuts              Flags
		GerudoFortressKeys     ShuffleKeys
		GerudoFortress         GerudoFortressCarpenterRescue
		GanonBossKeyShuffle    GanonBossKeyShuffle
	}

	Minigames struct {
		TreasureChestGameKeys ShuffleKeys
		KakarikoChickenGoal   int
		BigPoeGoal            int
	}

	Connections struct {
		OpenKakarikoGate OpenKakarikoGate
		OpenKokriForest  OpenForest
		OpenZoraFountain OpenZoraFountain

		Flags     ConnectionFlag
		Interior  InteriorShuffle
		Overworld OverworldShuffle
	}

	WinConditions struct {
		TriforceHunt  bool
		TriforceCount int
		TriforceGoal  int
		Lacs          ConditionedAmount
		Bridge        ConditionedAmount
		GanonBossKey  ConditionedAmount
		Trials        TrialFlag
	}

	Tricks map[string]bool
}

func finalizeLogic(_ *Zootr, _ *Logic) error {
	return notImpled
}
