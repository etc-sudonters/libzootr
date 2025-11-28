package magicbean

import (
	"fmt"
	"sudonters/libzootr/mido/objects"
	"sudonters/libzootr/settings"
	"sudonters/libzootr/zecs"
)

func ConstBool(b bool) objects.BuiltInFunction {
	obj := objects.PackedTrue
	if !b {
		obj = objects.PackedFalse
	}
	return func(*objects.Table, []objects.Object) (objects.Object, error) {
		return obj, nil
	}
}

type BuiltIns struct {
	CheckTodAccess    objects.BuiltInFunction `libzootr:"check_tod_access,params=1"`
	Has               objects.BuiltInFunction `libzootr:"has,params=2"`
	HasAnyOf          objects.BuiltInFunction `libzootr:"has_anyof,params=-1"`
	HasBottle         objects.BuiltInFunction `libzootr:"has_bottle,params=0"`
	HasDungeonRewards objects.BuiltInFunction `libzootr:"has_dungeon_rewards,params=1"`
	HasEvery          objects.BuiltInFunction `libzootr:"has_every,params=-1"`
	HasHearts         objects.BuiltInFunction `libzootr:"has_hearts,params=1"`
	HasMedallions     objects.BuiltInFunction `libzootr:"has_medallions,params=1"`
	HasNotesForSong   objects.BuiltInFunction `libzootr:"has_notes_for_song,params=1"`
	HasStones         objects.BuiltInFunction `libzootr:"has_stones,params=1"`
	IsAdult           objects.BuiltInFunction `libzootr:"is_adult,params=0"`
	IsChild           objects.BuiltInFunction `libzootr:"is_child,params=0"`
	IsStartingAge     objects.BuiltInFunction `libzootr:"is_starting_age,params=0"`
}

func (this BuiltIns) Table() objects.BuiltInFunctions {
	return objects.BuiltInFunctions{
		this.CheckTodAccess,
		this.Has,
		this.HasAnyOf,
		this.HasBottle,
		this.HasDungeonRewards,
		this.HasEvery,
		this.HasHearts,
		this.HasMedallions,
		this.HasNotesForSong,
		this.HasStones,
		this.IsAdult,
		this.IsChild,
		this.IsStartingAge,
	}
}

func CreateBuiltInDefs() []objects.BuiltInFunctionDef {
	return []objects.BuiltInFunctionDef{
		{Name: "check_tod_access", Params: 1},
		{Name: "has", Params: 2},
		{Name: "has_anyof", Params: -1},
		{Name: "has_bottle", Params: 0},
		{Name: "has_dungeon_rewards", Params: 1},
		{Name: "has_every", Params: -1},
		{Name: "has_hearts", Params: 1},
		{Name: "has_medallions", Params: 1},
		{Name: "has_notes_for_song", Params: 1},
		{Name: "has_stones", Params: 1},
		{Name: "is_adult", Params: 0},
		{Name: "is_child", Params: 0},
		{Name: "is_starting_age", Params: 0},
	}
}

func CreateBuiltInHasFuncs(builtins *BuiltIns, pocket *Pocket, flags settings.ShufflingFlags) {
	builtins.Has = func(tbl *objects.Table, args []objects.Object) (objects.Object, error) {
		if len(args) != 2 {
			return objects.Null, fmt.Errorf("has expects 2 arguments, got %d", len(args))
		}

		ptr := objects.UnpackPtr32(args[0])
		qty := objects.UnpackU32(args[1])
		result := pocket.Has(zecs.Entity(ptr.Addr), int(qty))
		return objects.PackBool(result), nil
	}

	builtins.HasAnyOf = func(_ *objects.Table, args []objects.Object) (objects.Object, error) {
		items := make([]zecs.Entity, len(args))
		for i, arg := range args {
			ptr := objects.UnpackPtr32(arg)
			items[i] = zecs.Entity(ptr.Addr)
		}

		result := pocket.HasAny(items)
		return objects.PackBool(result), nil
	}

	builtins.HasEvery = func(_ *objects.Table, args []objects.Object) (objects.Object, error) {
		items := make([]zecs.Entity, len(args))
		for i, arg := range args {
			ptr := objects.UnpackPtr32(arg)
			items[i] = zecs.Entity(ptr.Addr)
		}

		result := pocket.HasEvery(items)
		return objects.PackBool(result), nil
	}

	builtins.HasBottle = func(_ *objects.Table, _ []objects.Object) (objects.Object, error) {
		return objects.PackBool(pocket.HasBottle()), nil
	}

	builtins.HasDungeonRewards = func(_ *objects.Table, args []objects.Object) (objects.Object, error) {
		qty := objects.UnpackU32(args[0])
		return objects.PackBool(pocket.HasDungeonRewards(int(qty))), nil
	}

	builtins.HasHearts = func(_ *objects.Table, args []objects.Object) (objects.Object, error) {
		qty := objects.UnpackF64(args[0])
		return objects.PackBool(pocket.HasHearts(qty)), nil
	}

	builtins.HasStones = func(_ *objects.Table, args []objects.Object) (objects.Object, error) {
		qty := objects.UnpackU32(args[0])
		return objects.PackBool(pocket.HasStones(int(qty))), nil
	}

	builtins.HasMedallions = func(_ *objects.Table, args []objects.Object) (objects.Object, error) {
		qty := objects.UnpackU32(args[0])
		return objects.PackBool(pocket.HasMedallions(int(qty))), nil
	}

	if settings.HasFlag(flags, settings.ShuffleOcarinaNotes) {
		builtins.HasNotesForSong = func(_ *objects.Table, args []objects.Object) (objects.Object, error) {
			ptr := objects.UnpackPtr32(args[0])
			return objects.PackBool(pocket.HasAllNotes(zecs.Entity(ptr.Addr))), nil
		}
	} else {
		builtins.HasNotesForSong = ConstBool(true)
	}
}
