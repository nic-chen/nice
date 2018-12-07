package registry

type Options struct {
	Name string
	Host string
	Port string
	TTL  int
	Ssrv string
}

// The registry provides an interface for service discovery
type Registry interface {
	Register() error
	UnRegister()
}

var DefaultRegistry func(*Options) (Registry, error)

