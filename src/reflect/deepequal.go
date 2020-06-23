package reflect

func DeepEqual(x, y interface{}) bool {
	if x == nil || y == nil {
		return x == y
	}

	panic("unimplemented: reflect.DeepEqual()")
}
