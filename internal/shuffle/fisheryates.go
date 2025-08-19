package shuffle

import (
	"errors"
	"math/rand/v2"
)

var ErrEmptyQueue = errors.New("empty queue")
var ErrCannotRequeue = errors.New("cannot requeue item")

// FISO first in shrug out queue -- items are dequeued when RNGesus sees fit
// uses Fisher-Yates and underlying data is never discarded instead the slice
// is "sorted" in place. Unlike most in place FY implementations, this
// structure moves swapped elements to the front of the slice rather than the
// end. this allows new items to be added after dequeues have happened
func Empty[T comparable](rng *rand.Rand) *Q[T] {
	r := new(Q[T])
	r.rng = rng
	return r
}

// the passed slice is immediately shuffled
func From[T comparable](rng *rand.Rand, ts []T) *Q[T] {
	r := new(Q[T])
	r.rng = rng
	r.members = ts
	r.ShuffleRemaining()
	return r
}

// A "queue" that dequeues elements via incremental Fisher-Yates shuffling.
// Unlike most Fisher Yates implementations, this queue swaps items to
// _beginning_ of the slice.  This allows the queue to grow after some items
// have been dequeued. Additionally data is never discarded which provides a
// convenient and fast dequeue and requeue method for the fillers.
type Q[T comparable] struct {
	dqCount int
	members []T
	rng     *rand.Rand
}

func (this *Q[T]) Len() int {
	return len(this.members) - this.dqCount
}

func (this *Q[T]) Dequeue() (T, error) {
	var empty T
	curLen := len(this.members)

	if curLen == 0 || this.dqCount == curLen {
		return empty, ErrEmptyQueue
	}

	return this.dequeue(), nil
}

func (this *Q[T]) dequeue() T {
	// IntN is the half open range 0 (inclusive) to N (exclusive)
	// r.Len() is len(r.members) - r.dqCount and provides N.
	//   note: len(r.members) is the exclusive upper bound, any number 0
	//   (inclusive) to len(..) (exclusive) is a valid index in the slice
	// Since we swap to the _front_ of the slice we add the dqCount count to
	// this generated index to get the actual swapping index.
	// do the swap and spit out the value at the current index
	swap := this.rng.IntN(this.Len()) + this.dqCount
	current := this.dqCount
	this.dqCount += 1
	this.swap(current, swap)
	return this.members[current]
}

func (this *Q[T]) Enqueue(t T) {
	this.members = append(this.members, t)
}

func (this *Q[T]) EnqueueSlice(ts []T) {
	members := make([]T, 0, len(this.members)+len(ts))
	members = append(members, this.members...)
	members = append(members, ts...)
	this.members = members
}

func (this *Q[T]) Requeue(t T) error {
	if this.dqCount == 0 {
		return ErrEmptyQueue
	}

	if this.members[this.dqCount] != t {
		return ErrCannotRequeue
	}

	this.dqCount -= 1
	return nil
}

// Shuffles only indexes eligble for dequeuing.
func (this *Q[T]) ShuffleRemaining() {
	this.rng.Shuffle(this.Len(), func(i, j int) {
		this.swap(i+this.dqCount, j+this.dqCount)
	})
}

// Returns slices of dequeued and remaining T
func (r *Q[T]) Parts() (dequeued []T, remaining []T) {
	dequeued = make([]T, r.dqCount)
	remaining = make([]T, r.Len())

	copy(dequeued, r.members[:r.dqCount])
	copy(remaining, r.members[r.dqCount:])

	return dequeued, remaining
}

func (this *Q[T]) swap(i, j int) {
	this.members[i], this.members[j] = this.members[j], this.members[i]
}

// Convenience iterator that repeatedly randomly dequeues until all items are
// dequeued. Items can be requeued as normal, however care must be taken to
// avoid accidentally creating infinite loops:
//
//	```golang
//	q := RandomQueueFrom(rng, theSlice)
//	for something := range q.All {
//	    if !predicate(something) {
//	        q.RequeueLast()
//	    }
//	}
//	```
//
// Depending on how `predicate` determines something should be requeued it may
// be possible to end up in a situation where the predicate _always_ rejects
// the item, placing it back into the queue and eventually the queue consists
// exclusively of items the predicate will reject.
func (this *Q[T]) All(yield func(T) bool) {
	for this.Len() > 0 {
		if !yield(this.dequeue()) {
			return
		}
	}
}
