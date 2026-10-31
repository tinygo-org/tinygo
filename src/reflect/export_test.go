package reflect

type OtherPkgFields struct {
	OtherExported   int
	otherUnexported int
}

type Buffer struct {
	buf []byte
}
