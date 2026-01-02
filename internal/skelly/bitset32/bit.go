package bitset32

import (
	"math/bits"
)

func Buckets(i uint32) int {
	return int(i / 32)
}

func BitIndex(i uint32) uint32 {
	return 1 << (i % 32)
}

func New(i int) Bitset {
	var b Bitset
	b.buckets = make([]uint32, i)
	return b
}

func Create(members ...uint32) Bitset {
	b := New(0)
	for _, m := range members {
		b.Set(m)
	}
	return b
}

func WithBucketsFor(i uint32) Bitset {
	return New(Buckets(i))
}

func FromRaw(parts ...uint32) Bitset {
	var b Bitset
	b.buckets = parts
	return b
}

func ToRawParts(b Bitset) []uint32 {
	ret := make([]uint32, len(b.buckets))
	copy(ret, b.buckets)
	return ret
}

type Bitset struct {
	buckets []uint32
}

func IsEmpty(b Bitset) bool {
	for i := range b.buckets {
		if b.buckets[i] != 0 {
			return false
		}
	}
	return true
}

func Copy(b Bitset) Bitset {
	var n Bitset
	n.buckets = make([]uint32, len(b.buckets))
	copy(n.buckets, b.buckets)
	return n
}

func (this *Bitset) resize(bucket int) {
	if bucket < len(this.buckets) {
		return
	}

	buckets := make([]uint32, bucket+1)
	copy(buckets, this.buckets)
	this.buckets = buckets
}

func (this *Bitset) Set(i uint32) bool {
	idx := Buckets(i)
	bit := BitIndex(i)
	this.resize(idx)
	bucket := this.buckets[idx]
	this.buckets[idx] = bucket | bit
	return bucket&bit == 0
}

func (this Bitset) Unset(i uint32) {
	bucket := Buckets(i)

	if bucket >= len(this.buckets) {
		return
	}

	this.buckets[bucket] &= ^BitIndex(i)
}

func (this Bitset) IsSet(i uint32) bool {
	bucket := Buckets(i)
	if bucket >= len(this.buckets) {
		return false
	}

	bit := BitIndex(i)
	return bit == (bit & this.buckets[bucket])
}

func (this Bitset) Complement() Bitset {
	n := Copy(this)
	for i, bits := range n.buckets {
		n.buckets[i] = ^bits
	}
	return n
}

func (this Bitset) Intersect(n Bitset) Bitset {
	buckets := min(len(this.buckets), len(n.buckets))
	r := Bitset{}
	r.buckets = make([]uint32, buckets)

	for i := range r.buckets {
		r.buckets[i] = this.buckets[i] & n.buckets[i]
	}

	return r
}

func (this Bitset) Union(n Bitset) Bitset {
	if len(n.buckets) > len(this.buckets) {
		this, n = n, this
	}

	ret := Copy(this)
	for i, bits := range n.buckets {
		ret.buckets[i] |= bits
	}

	return ret
}

func (this Bitset) Difference(n Bitset) Bitset {
	buckets := make([]uint32, max(len(this.buckets), len(n.buckets)))
	copy(buckets, n.buckets)
	n.buckets = buckets
	return this.Intersect(n.Complement())
}

func (this Bitset) Eq(n Bitset) bool {
	hi, lo := this.buckets, n.buckets

	if len(lo) > len(hi) {
		hi, lo = lo, hi
		for _, bucket := range hi[len(lo):] {
			if bucket != 0 {
				return false
			}
		}
	}

	for i := range lo {
		if lo[i] != hi[i] {
			return false
		}
	}

	return true
}

func (this Bitset) Len() int {
	var count int

	for _, bucket := range this.buckets {
		for ; bucket != 0; bucket &= bucket - 1 {
			count++
		}
	}

	return count
}

func (this Bitset) Elems() []uint32 {
	var elems []uint32

	for k, bucket := range this.buckets {
		k := uint32(k)
		for bucket != 0 {
			tz := uint32(bits.TrailingZeros32(bucket))
			elems = append(elems, k*32+tz)
			bucket ^= (1 << tz)
		}
	}

	return elems
}

func (this Bitset) Pop() uint32 {
	for k, bucket := range this.buckets {
		if bucket == 0 {
			continue
		}

		tz := uint32(bits.TrailingZeros32(bucket))
		this.buckets[k] = bucket ^ (1 << tz)
		return uint32(k*32) + tz
	}
	return 0
}

func (this Bitset) IsEmpty() bool {
	for _, bucket := range this.buckets {
		if bucket != 0 {
			return false
		}
	}
	return true
}
