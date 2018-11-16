package registry

import (
	"crypto/tls"
	"time"
)

// The registry provides an interface for service discovery
type Registry interface {
	Register(serviceName string, node *Node, opts ...RegisterOption) error
	Unregister(serviceName string, node *Node) error
	GetClient() interface{}
}

type Node struct {
	Id       string            `json:"id"`
	Address  string            `json:"address"`
	Metadata map[string]string `json:"metadata"`
}

var DefaultRegistry func(opts ...Option) (Registry, error)

type Option func(*Options)

type Options struct {
	Timeout   time.Duration
	TLSConfig *tls.Config
	Dsn       string
}

type RegisterOption func(*RegisterOptions)

type RegisterOptions struct {
	TTL time.Duration
}

// Specify TLS Config
func TLSConfig(t *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = t
	}
}

func Dsn(dsn string) Option {
	return func(o *Options) {
		o.Dsn = dsn
	}
}

func RegisterTTL(t time.Duration) RegisterOption {
	return func(o *RegisterOptions) {
		o.TTL = t
	}
}
