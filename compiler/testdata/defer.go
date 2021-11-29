package main

func external()

func deferSimple() {
	defer func() {
		print(3)
	}()
	external()
}

func deferMultiple() {
	defer func() {
		print(3)
	}()
	defer func() {
		print(5)
	}()
	external()
}
