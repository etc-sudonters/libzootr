package settings

type Model struct {
	Seed       uint64
	Logic      Logic
	Cosmetics  Cosmetics
	Generation Generation
	Rom        Rom
}

func Finalize(zootr *Zootr) (Model, error) {
	var m Model
	return m, notImpled
}

func FromString(encoded string) (Model, error) {
	z, err := decodeSettingStr(encoded)
	if err != nil {
		return Model{}, err
	}

	return Finalize(&z)
}

func Default() Model {
	var m Model
	// most values have the correct 0 value, but some don't
	m.Logic.Dungeon.GerudoFortressKeys = ShuffleKeysVanilla
	m.Logic.Dungeon.SilverRupees = ShuffleKeysVanilla
	m.Logic.Minigames.BigPoeGoal = 10
	m.Logic.Minigames.KakarikoChickenGoal = 7
	m.Logic.Minigames.TreasureChestGameKeys = ShuffleKeysVanilla
	m.Logic.Shuffling.Flags = ShuffleKokiriSword
	m.Logic.Spawns.Hearts = 3
	m.Logic.Trade.AdultItems = AdultTradeItemsAll
	m.Logic.WinConditions.Bridge = EncodeConditionedAmount(CondMedallions, 6)
	m.Logic.WinConditions.Lacs = EncodeConditionedAmount(CondVanilla, 0)
	m.Logic.WinConditions.Trials = TrialAll
	return m
}
