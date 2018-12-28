package micro

import (
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	
	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	//"github.com/nic-chen/nice/micro/registry"
	//"github.com/nic-chen/nice/micro/tracing"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
)

type Server struct {
	Name   string
	Option *serverOptions
}

func NewServer(name string, opts ...Option) (*Server, error) {
	var err error
	sOptions := &serverOptions{}
	sOptions.applyOption(opts...)
	srv := &Server{
		Name:   name,
		Option: sOptions,
	}

	if sOptions.prometheus {
		sOptions.applyOption(WithUnaryServerInterceptor(grpc_prometheus.UnaryServerInterceptor))
		sOptions.applyOption(WithStreamServerInterceptor(grpc_prometheus.StreamServerInterceptor))
	}

	if sOptions.authFunc != nil {
		sOptions.applyOption(WithUnaryServerInterceptor(grpc_auth.UnaryServerInterceptor(sOptions.authFunc)))
		sOptions.applyOption(WithStreamServerInterceptor(grpc_auth.StreamServerInterceptor(sOptions.authFunc)))
	}

	if sOptions.recovery == nil {
		sOptions.recovery = RecoveryWithLogger(sOptions)
		sOptions.applyOption(WithUnaryServerInterceptor(grpc_recovery.UnaryServerInterceptor(
			grpc_recovery.WithRecoveryHandler(sOptions.recovery))))
		sOptions.applyOption(WithStreamServerInterceptor(grpc_recovery.StreamServerInterceptor(
			grpc_recovery.WithRecoveryHandler(sOptions.recovery))))
	}

	return srv, err
}

func (t Server) BuildGrpcServer() *grpc.Server {
	var opts = t.Option.grpcOptions

	if t.Option.tracer != nil {
		opts = append(opts, grpc.UnaryInterceptor(
			otgrpc.OpenTracingServerInterceptor(t.Option.tracer),
		))
	}

	rpcSrv := grpc.NewServer(opts...)
	return rpcSrv
}

func (t Server) Run(rpcSrv *grpc.Server, listen string) error {
	lis, err := net.Listen("tcp", listen)
	if err != nil {
		panic(err)
	}

	if t.Option.prometheus {
		t.StartPrometheus(rpcSrv)
	}

	if t.Option.register != nil {
		if err = t.Option.register.Register(); err != nil {
			return err
		}
	} else {
		reflection.Register(rpcSrv)
	}

	log.Printf("%s tcp server will be ready for listening at:%s", t.Name, listen)
	return rpcSrv.Serve(lis)
}

func (t Server) StartPrometheus(rpcSrv *grpc.Server) {
	// After all your registrations, make sure all of the Prometheus metrics are initialized.
	grpc_prometheus.Register(rpcSrv)
	// standalone http server
	if t.Option.prometheusListen != "" {
		// Register Prometheus metrics handler.
		httpServer := &http.Server{
			Handler: promhttp.Handler(),
			Addr:    t.Option.prometheusListen,
		}
		go func() {
			log.Printf("starting prometheus http server at:%s", httpServer.Addr)
			if err := httpServer.ListenAndServe(); err != nil {
				log.Fatal("Unable to start a http server.")
			}
		}()
	}
}
