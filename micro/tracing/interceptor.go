package tracing

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type tracingTag struct{}

var (
	tracingTagKey = &tracingTag{}
)

const (
	// opentracing log key is trace.traceid
	ContextHeaderName = "nice-traceid"
	TraceIdKey        = "traceid"
)

type ClientTraceIdFunc func(context.Context) (context.Context, error)

// support get string key from context.such as gin
func DefaultClientTraceIdFunc(fromHttpHeader bool) ClientTraceIdFunc {
	return func(ctx context.Context) (context.Context, error) {
		var tid interface{}
		if fromHttpHeader {
			tid = ctx.Value(ContextHeaderName)
		} else {
			tid = ctx.Value(tracingTagKey)
		}
		if tid == nil {
			return ctx, nil
		}

		newCtx := metadata.AppendToOutgoingContext(ctx, ContextHeaderName, tid.(string))
		return newCtx, nil
	}
}

// append to context
func ToContext(ctx context.Context, traceId string) context.Context {
	return context.WithValue(ctx, tracingTagKey, traceId)
}

// UnaryClientInterceptor returns a new unary client interceptor that optionally logs the execution of external gRPC calls.
func UnaryClientInterceptor(cidFunc ClientTraceIdFunc) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var err error
		var newCtx context.Context
		if newCtx, err = cidFunc(ctx); err == nil {
			err = invoker(newCtx, method, req, reply, cc, opts...)
		}
		return err
	}
}

// 全局跟踪ID,本方式用于在不启用operation tracer情况下,仍然可以跟踪整个请求.
// 客户端通过metadata向服务端传递
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		v := metautils.ExtractIncoming(ctx).Get(ContextHeaderName)
		if v != "" {
			ctxzap.AddFields(ctx, zap.String(TraceIdKey, v))
		}
		newCtx := ToContext(ctx, v)
		return handler(newCtx, req)
	}
}

func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		v := metautils.ExtractIncoming(stream.Context()).Get(ContextHeaderName)
		if v != "" {
			ctxzap.AddFields(ctx, zap.String(TraceIdKey, v))
		}
		newCtx := ToContext(ctx, v)
		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx
		return handler(srv, wrapped)
	}
}
