package mock

import (
	"context"
	"errors"
	"github.com/nic-chen/nice/micro"
	"github.com/nic-chen/nice/micro/mock/prototest"
	"github.com/nic-chen/nice/micro/registry"
	"github.com/nic-chen/nice/micro/registry/etcdv3"
	"google.golang.org/grpc"
	"log"
	"net"
)

//grpc server
type Hello struct {
	listen string
}

func (t *Hello) Hello(ctx context.Context, req *prototest.Request) (*prototest.Response, error) {
	res := &prototest.Response{}
	if req == nil {
		req = &prototest.Request{}
	}
	res.Msg = req.Name + ", from:" + t.listen
	return res, nil
}

func (t *Hello) HelloError(ctx context.Context, req *prototest.Request) (*prototest.Response, error) {
	return nil, errors.New("yes,it error")
	//panic("yes,it panic")
}

func (t *Hello) HelloPanic(ctx context.Context, req *prototest.Request) (*prototest.Response, error) {
	panic("yes,it panic")
}

const TestSvrName = "test"
const TestSvrListen = ":54321"

func NewMockDiscoveryServer(listen string, nodeId string) error {

	l, e := net.Listen("tcp", listen)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	server := grpc.NewServer()
	prototest.RegisterSayServer(server, &Hello{listen})

	r, err := etcdv3.NewRegistry(registry.Dsn("127.0.0.1:2379"))
	if err != nil {
		panic(err)
	}
	err = r.Register(TestSvrName, &registry.Node{Id: nodeId, Address: listen})
	if err != nil {
		panic(err)
	}
	server.Serve(l)
	return nil
}

func NewMicroServer(listen string) error {
	a, err := micro.NewServer("test")
	if err != nil {
		return err
	}

	arpc := a.BuildGrpcServer()
	prototest.RegisterSayServer(arpc, &Hello{})
	return a.Run(arpc, listen)
}
