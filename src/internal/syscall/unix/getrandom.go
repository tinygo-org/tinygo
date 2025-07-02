package unix

type GetRandomFlag uintptr

const (
	GRND_NONBLOCK GetRandomFlag = 0x0001
	GRND_RANDOM   GetRandomFlag = 0x0002
)

func GetRandom(p []byte, flags GetRandomFlag) (n int, err error) {
	panic("todo: unix.GetRandom")
}
