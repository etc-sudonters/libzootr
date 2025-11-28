package internal

func Todo[T any]() T {
	return *(*T)(nil)
}

func NeedsErrorHandling(e error) {
	PanicOnError(e)
}

func PanicOnError(e error) {
	if e != nil {
		panic(e)
	}
}
