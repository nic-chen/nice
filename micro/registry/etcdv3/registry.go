package etcdv3

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/nic-chen/nice/micro/registry"
	"go.etcd.io/etcd/clientv3"
)

// Prefix should start and end with no slash
var Prefix = "nice://"
var Deregister = make(chan struct{})

type etcdv3Registry struct {
	client  *clientv3.Client
	options *registry.Options
	leases  map[string]clientv3.LeaseID
}

func NewRegistry(options *registry.Options) (r registry.Registry, err error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: strings.Split(options.Ssrv, ","),
	})

	if err != nil {
		return nil, err
	}

	r = &etcdv3Registry{
		client:  cli,
		options: options,
	}

	return r, nil
}

func init() {
	registry.DefaultRegistry = NewRegistry
}

// Register
func (t *etcdv3Registry) Register() error {
	serviceValue := net.JoinHostPort(t.options.Host, t.options.Port)
	serviceKey := fmt.Sprintf("%s/%s/%s", Prefix, t.options.Name, serviceValue)

	resp, err := t.client.Grant(context.TODO(), int64(t.options.TTL))
	if err != nil {
		return fmt.Errorf("grpclb: create clientv3 lease failed: %v", err)
	}

	if _, err := t.client.Put(context.TODO(), serviceKey, serviceValue, clientv3.WithLease(resp.ID)); err != nil {
		return fmt.Errorf("grpclb: set service '%s' with TTL to clientv3 failed: %s", t.options.Name, err.Error())
	}

	if _, err := t.client.KeepAlive(context.TODO(), resp.ID); err != nil {
		return fmt.Errorf("grpclb: refresh service '%s' with TTL to clientv3 failed: %s", t.options.Name, err.Error())
	}

	// wait deregister then delete
	go func() {
		<-Deregister
		t.client.Delete(context.Background(), serviceKey)
		Deregister <- struct{}{}
	}()

	return nil
}

// UnRegister delete registered service from etcd
func (t *etcdv3Registry) UnRegister() {
	Deregister <- struct{}{}
	<-Deregister
}
