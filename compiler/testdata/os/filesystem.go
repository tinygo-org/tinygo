package os

type FileHandle interface {
	Close() error
}
