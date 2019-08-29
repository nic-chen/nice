package etcdv3_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nic-chen/nice/micro/mock"
	"github.com/nic-chen/nice/micro/mock/prototest"
	"github.com/nic-chen/nice/micro/registry"
	"github.com/nic-chen/nice/micro/registry/etcdv3"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"testing"
	"time"
)

func init() {
	go mock.NewMockDiscoveryServer(":12345", "default")
}

func TestParse(t *testing.T) {
	js := `{"auto-sync-interval":100}`
	cnf := &clientv3.Config{}
	json.Unmarshal([]byte(js), cnf)
	t.Log(cnf.AutoSyncInterval)
}

func TestEtcdv3Registry_Build(t *testing.T) {
	r, err := etcdv3.NewRegistry(registry.Dsn("127.0.0.1:2379"))
	if err != nil {
		panic(err)
	}

	b := r.(resolver.Builder)
	resolver.Register(b)

	_, err = grpc.Dial(mock.TestSvrName, grpc.WithBalancerName("round_robin"), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
}

func TestNewRegistry(t *testing.T) {
	if _, err := etcdv3.NewRegistry(registry.Dsn("127.0.0.1:2379")); err != nil {
		t.Error(err)
	}

	r, err := etcdv3.NewRegistry(
		registry.Dsn("127.0.0.1:2379?username=qsli&password=123456&dial-timeout=7200"),
	)
	if err != nil {
		t.Error(err)
	}
	if r == nil {
		t.Error("instance init error")
	}

	t.Log(r)
}

func TestEtcdv3Registry_Register(t *testing.T) {
	r, err := etcdv3.NewRegistry(registry.Dsn("127.0.0.1:2379"))
	if err != nil {
		t.Error(err)
	}
	err = r.Register(mock.TestSvrName, &registry.Node{Id: mock.TestSvrName, Address: ":12345"})
	if err != nil {
		t.Error(err)
	}
}

func TestMulti(t *testing.T) {
	//
	go mock.NewMockDiscoveryServer(":12346", "n2")
	go mock.NewMockDiscoveryServer(":12347", "n3")
	r, _ := etcdv3.NewRegistry(registry.Dsn("127.0.0.1:2379"))
	b := r.(resolver.Builder)
	resolver.Register(b)
	//d1
	d1, err := grpc.Dial(b.Scheme()+"://author/"+mock.TestSvrName, grpc.WithBalancerName("round_robin"), grpc.WithInsecure())
	client := prototest.NewSayClient(d1)
	if err != nil {
		panic(err)
	}
	for {
		out, err := client.Hello(context.Background(), &prototest.Request{Name: "test"})
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(out)
		}
		<-time.After(time.Second)
	}
}
