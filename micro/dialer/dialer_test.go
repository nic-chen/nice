package dialer_test

import (
	"github.com/opentracing/opentracing-go"
	"github.com/nic-chen/nice/micro/dialer"
	"google.golang.org/grpc"
	"strings"
	"testing"
	"time"
)

type tracer struct {
}

func (tracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	return nil
}
func (tracer) Inject(sm opentracing.SpanContext, format interface{}, carrier interface{}) error {
	return nil
}
func (tracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	return nil, nil
}

func TestWithTracer(t *testing.T) {
	o := dialer.WithTracer(&tracer{})
	os := &dialer.Options{}
	o(os)
	if os.Tracer == nil {
		t.Errorf("dialer with tracer fail")
	}
}

func TestWithDialOption(t *testing.T) {
	o := dialer.WithDialOption(grpc.WithInsecure())
	os := &dialer.Options{}
	o(os)
	if os.DialOptions[0] == nil {
		t.Errorf("dialer with tracer fail")
	}
}

func TestWithUnaryClientInterceptor(t *testing.T) {
	os := &dialer.Options{}
	if os.UnaryClientInterceptors[0] == nil {
		t.Errorf("dialer with tracer fail")
	}
}

func TestDial(t *testing.T) {
	_, err := dialer.Dial("test",
		dialer.WithDialOption(grpc.WithTimeout(1*time.Second)),
		dialer.WithTracer(&tracer{}),
	)
	if strings.Index(err.Error(), "failed to dial") == 0 {
		return
	}
	t.Errorf("dial get err:%s", err)
}
