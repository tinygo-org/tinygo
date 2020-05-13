package os

type Signal interface {
	String() string
	Signal() // to distinguish from other Stringers
}
