package micro

import (
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	//"github.com/nic-chen/nice/micro/registry"
	"github.com/nic-chen/nice/micro/tracing"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
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

	uins := []grpc.UnaryServerInterceptor{grpc_ctxtags.UnaryServerInterceptor()}
	sins := []grpc.StreamServerInterceptor{grpc_ctxtags.StreamServerInterceptor()}

	if sOptions.tracer != nil {
		uins = append(uins, grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithTracer(sOptions.tracer)))
		sins = append(sins, grpc_opentracing.StreamServerInterceptor(grpc_opentracing.WithTracer(sOptions.tracer)))
	}
	if sOptions.logger == nil {
		sOptions.logger, err = zap.NewDevelopment()
		if err != nil {
			return nil, err
		}
	}
	uins = append(uins, grpc_zap.UnaryServerInterceptor(sOptions.logger))

	sins = append(sins, grpc_zap.StreamServerInterceptor(sOptions.logger))
	grpc_zap.ReplaceGrpcLogger(sOptions.logger)
	// if tracer is nil then set a id use for log request id
	if sOptions.logger != nil {
		uins = append(uins, tracing.UnaryServerInterceptor())
		sins = append(sins, tracing.StreamServerInterceptor())
	}

	// tag and tracer must at first
	sOptions.unaryServerInterceptors = append(uins, sOptions.unaryServerInterceptors...)
	sOptions.streamServerInterceptors = append(sins, sOptions.streamServerInterceptors...)

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
	}

	sOptions.applyOption(WithUnaryServerInterceptor(grpc_recovery.UnaryServerInterceptor(
		grpc_recovery.WithRecoveryHandler(sOptions.recovery))))
	sOptions.applyOption(WithStreamServerInterceptor(grpc_recovery.StreamServerInterceptor(
		grpc_recovery.WithRecoveryHandler(sOptions.recovery))))

	return srv, err
}

func (t Server) BuildGrpcServer() *grpc.Server {
	var opts = t.Option.grpcOptions
	if len(t.Option.unaryServerInterceptors) > 0 {
		opts = append(opts, grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(t.Option.unaryServerInterceptors...),
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
