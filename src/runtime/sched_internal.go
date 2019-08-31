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

func getFakeCoroutine() *task {
	// this isnt defined behavior, but this is what our implementation does
	// this is really a horrible hack
	var t *task
	go fakeCoroutine(&t)

	// the first line of fakeCoroutine will have completed by now
	return t
}
