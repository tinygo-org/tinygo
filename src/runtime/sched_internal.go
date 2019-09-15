package runtime

// noret is a placeholder that can be used to indicate that an async function is not going to directly return here
func noret() {}

func getParentHandle() *task {
	panic("NOPE")
}

func fakeCoroutine(dst **task) {
	*dst = getCoroutine()
	for {
		yield()
	}
}
