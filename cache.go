package nice

type Cache interface {
	Open()
	Close() error
	Do(command string, args ...interface{}) (interface{}, error)
}