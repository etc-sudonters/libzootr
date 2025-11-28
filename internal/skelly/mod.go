package skelly

import "github.com/etc-sudonters/substrate/skelly/bitset32"

type bs32 = bitset32.Bitset
type Bitset32Reducer func(bs32, bs32) bs32

func ReduceBitset32(op Bitset32Reducer, seed bs32, sets ...bs32) bs32 {
	if len(sets) == 0 {
		return bitset32.Copy(seed)
	}

	for _, set := range sets {
		seed = op(seed, set)
	}

	return seed
}

func IntersectAll(first bs32, sets ...bs32) bs32 {
	return ReduceBitset32(bitset32.Bitset.Intersect, first, sets...)
}

func UnionAll(first bs32, sets ...bs32) bs32 {
	return ReduceBitset32(bitset32.Bitset.Union, first, sets...)
}
