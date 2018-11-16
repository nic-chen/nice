package dialer

import (
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"github.com/nic-chen/nice/micro/tracing"
	"google.golang.org/grpc"
)

//
type Options struct {
	Tracer                  opentracing.Tracer
	UnaryClientInterceptors []grpc.UnaryClientInterceptor
	DialOptions             []grpc.DialOption
	TraceIdFunc             tracing.ClientTraceIdFunc
}

type Option func(*Options)

// WithTracer traces rpc calls,the Tracer interceptor must put the last
// because the context  get value limit
// if use jeagertracer,the context header key will be uber-trace-id,it can't log by zap logger.
// if you want auto log by CtxTag,the key must be TagTraceID value,set the jeager's configuration key headers.TraceContextHeaderName to `trace.traceid`
func WithTracer(t opentracing.Tracer) Option {
	return func(options *Options) {
		options.Tracer = t
	}
}

func WithDialOption(gopts ...grpc.DialOption) Option {
	return func(options *Options) {
		options.DialOptions = gopts
	}
}

func WithTraceIdFunc(idFunc tracing.ClientTraceIdFunc) Option {
	return func(options *Options) {
		options.TraceIdFunc = idFunc
	}
}

// Dial returns a load balanced grpc client conn with tracing interceptor
func Dial(name string, opts ...Option) (*grpc.ClientConn, error) {
	options := Options{}

	for _, v := range opts {
		v(&options)
	}

	if options.TraceIdFunc != nil {
		options.UnaryClientInterceptors = append(options.UnaryClientInterceptors, tracing.UnaryClientInterceptor(options.TraceIdFunc))
	}

	if options.Tracer != nil {
		// keep Tracer is last
		options.UnaryClientInterceptors = append(options.UnaryClientInterceptors, grpc_opentracing.UnaryClientInterceptor(grpc_opentracing.WithTracer(options.Tracer)))
	}

	uopt := grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(options.UnaryClientInterceptors...))

	conn, err := grpc.Dial(name, append(options.DialOptions, uopt)...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %v", name, err)
	}

	return conn, nil
}

func WithUnaryClientInterceptor(interceptors ...grpc.UnaryClientInterceptor) Option {
	return func(options *Options) {
		options.UnaryClientInterceptors = append(options.UnaryClientInterceptors, interceptors...)
	}
}
