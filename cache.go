package nice

type Cache interface {
	Open() error
	Close() error
	Do(command string, args ...interface{}) (interface{}, error)
}