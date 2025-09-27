package tracing

import (
	"io"
	"log"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

var (
	defaultSampleRatio float64 = 1
)

// Init returns a newly configured tracer and closer
func Init(serviceName, host string) (opentracing.Tracer, io.Closer, error) {
	ratio := defaultSampleRatio
	log.Printf("jaeger: tracing sample ratio %f", ratio)
	cfg := jaegercfg.Configuration{
		ServiceName: serviceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  "probabilistic",
			Param: ratio,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  host,
		},
	}
	logger := jaegerlog.StdLogger
	tracer, closer, err := cfg.NewTracer(jaegercfg.Logger(logger))
	if err != nil {
		return nil, nil, err
	}
	return tracer, closer, nil
}
