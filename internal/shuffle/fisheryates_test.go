package shuffle

import (
	"math/rand/v2"
	"testing"

	"github.com/etc-sudonters/substrate/rng"
)

func TestDequeueAllItems(t *testing.T) {
	enqueued := []int{1, 2, 3, 4}
	rng := rng.NewXoshiro256PPFromU64(0xbf58476d1ce4e5b9)
	q := Empty[int](rand.New(rng))
	q.EnqueueSlice(enqueued)
	result := map[int]int{}

	for res := range q.All {
		result[res] = 1
	}

	checkWasDequeued := func(x int) {
		if _, exists := result[x]; !exists {
			t.Fail()
			t.Logf("expected %d to have been dequeued but was not", x)
		}
	}

	for _, num := range enqueued {
		checkWasDequeued(num)
	}
}
