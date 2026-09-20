package interfacepkg

type Unexported = struct{ x int }
type Exported = struct{ X int }

func New() Unexported {
	return Unexported{x: 1}
}
