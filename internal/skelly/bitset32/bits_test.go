package bitset32

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/etc-sudonters/substrate/rng"
)

const (
	minU uint32 = 0
	maxU        = ^minU
)

func TestSetsBits(t *testing.T) {
	expected := []uint32{2, 2, 2}

	numbers := []uint32{
		1,
		65,
		129,
	}

	b := Bitset{}
	for i := range numbers {
		b.Set(numbers[i])
	}

	for i := range expected {
		if expected[i] != b.buckets[i] {
			t.FailNow()
		}
	}
}

func TestClearsBits(t *testing.T) {
	expected := []uint32{2, 0, 2}

	b := Bitset{}
	numbers := []uint32{1, 65, 129}
	for i := range numbers {
		b.Set(numbers[i])
	}
	b.Unset(65)

	for i := range expected {
		if expected[i] != b.buckets[i] {
			t.FailNow()
		}
	}
}

func TestTestBits(t *testing.T) {
	var b Bitset
	b.buckets = []uint32{2, 2, 2}

	if !b.IsSet(1) {
		t.Log("expected 1 to be set")
		t.Fail()
	}

	if !b.IsSet(65) {
		t.Log("expected 65 to be set")
		t.Fail()
	}

	if !b.IsSet(129) {
		t.Log("expected 129 to be set")
		t.Fail()
	}
}

func TestComplement(t *testing.T) {
	b := Bitset{}
	b.Set(1)
	b.Set(65)
	b.Set(129)

	comp := b.Complement().buckets
	expected := maxU ^ 2

	if comp[0] != expected || comp[1] != expected || comp[2] != expected {
		t.Fail()
	}
}

func TestIntersect(t *testing.T) {
	b1 := Bitset{}
	b2 := Bitset{}

	shared := []uint32{1, 65, 129}
	b1.Set(144)
	b2.Set(13)

	for i := range shared {
		b1.Set(shared[i])
		b2.Set(shared[i])
	}

	I := b1.Intersect(b2).buckets

	if I[0] != 2 || I[1] != 2 || I[2] != 2 {
		t.Fail()
	}
}

func TestUnion(t *testing.T) {
	b1 := Bitset{}
	b2 := Bitset{}
	b3 := Bitset{}

	b1.Set(1)
	b2.Set(65)
	b3.Set(129)

	b := b1.Union(b2).Union(b3)

	if !b.Eq(FromRaw(2, 2, 2)) {
		t.Fail()
	}
}

func TestDifference(t *testing.T) {
	b1 := Bitset{}
	b2 := Bitset{}

	b1.Set(1)
	b1.Set(65)
	b2.Set(65)
	b2.Set(129)

	b1DiffB2 := b1.Difference(b2)
	expected := FromRaw(2, 0, 0)

	if !b1DiffB2.Eq(expected) {
		t.Log("expected only 1 to be set")
		t.Fail()
	}

	b2DiffB1 := b2.Difference(b1)
	expected = FromRaw(0, 0, 2)

	if !b2DiffB1.Eq(expected) {
		t.Log("expected only 129 to be set")
		t.Fail()
	}
}

func TestLength(t *testing.T) {
	b := Bitset{}
	r := rand.New(rng.NewXoshiro256PPFromU64(0xbf58476d1ce4e5b9))
	for range 5000 {
		for !b.Set(r.Uint32N(math.MaxUint16)) {
		}
	}
	if l := b.Len(); l != 5000 {
		t.Fatalf("expected length of 5000 but got %d", l)
	}
}

func TestElems(t *testing.T) {
	b := Bitset{}
	b.Set(1)
	b.Set(65)
	b.Set(129)

	expected := []uint32{1, 65, 129}
	elems := b.Elems()

	if len(expected) != len(elems) {
		t.Fatalf("mismatched elems\nexpected:\t%+v\nactual:\t%+v", expected, elems)
	}

	for idx := range elems {
		a, b := expected[idx], elems[idx]
		if a != b {
			t.Logf("expected to find %d at index %d but found %d", a, idx, b)
			t.Fail()
		}
	}

	b = WithBucketsFor(10000)
	expected = make([]uint32, 0, 5000)

	for i := 0; i < 10000; i += 2 {
		b.Set(uint32(i))
		expected = append(expected, uint32(i))
	}

	elems = b.Elems()
	l := b.Len()

	if len(elems) != l {
		t.Fatalf("len(Elems()) and Len() disagree\nlen(Elems()) = %d\nLen() = %d", len(elems), l)
	}

	if l != len(expected) {
		t.Fatalf("expected length of %d but got %d", len(expected), l)
	}

	for idx := range elems {
		a, b := expected[idx], elems[idx]
		if a != b {
			t.Logf("expected to find %d at index %d but found %d", a, idx, b)
			t.Fail()
		}
	}
}

func TestEq(t *testing.T) {
	b1 := Bitset{}
	b1.Set(32)
	b2 := Copy(b1)
	b2.resize(3)

	if !b1.Eq(b2) {
		t.Log("expected b1 == b2")
		t.Fail()
	}

	if !b2.Eq(b1) {
		t.Log("expected b2 == b1")
		t.Fail()
	}
}
