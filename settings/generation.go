package settings

type Generation struct {
	StartWithConsumables      bool
	StartWithRupees           bool
	SkipTowerEscape           bool
	SkipCastleStealth         bool
	SkipEponaRace             bool
	SkipSomeMinigamePhases    bool
	KeepGlitchUsefulCutscenes bool
	FastChests                bool
	AutoEquipMasks            bool
	RandomChickenCount        bool
	RandomPoeCount            bool
	RandomStartingAge         bool
	RandomTrialsEnabled       bool
	TrialCount                int
	RandomStartTimeOfDay      bool
	FastBunnyHood             bool

	EasierFireArrowEntry bool
	RutoAlreadyOnFloor1  bool
	CAMC                 bool //chest appearance matches contents
	PAMC                 bool // pot ...
}

func finalizeGeneration(z *Zootr, g *Generation) error {
	return notImpled
}
