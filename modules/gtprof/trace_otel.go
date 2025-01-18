// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package gtprof

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type traceOtelStarter struct{}

type traceOtelSpan struct {
	tsShim *TraceSpan
	tsOtel trace.Span

	internalSpanIdx int
}

func toOtelAttributes(attrs ...*TraceAttribute) []attribute.KeyValue {
	otelAttrs := make([]attribute.KeyValue, len(attrs))
	for i, a := range attrs {
		v := a.Value.v
		switch v.(type) {
		case float32, float64:
			otelAttrs[i] = attribute.Float64(a.Key, a.Value.AsFloat64())
		case int, int64:
			otelAttrs[i] = attribute.Int64(a.Key, a.Value.AsInt64())
		default:
			otelAttrs[i] = attribute.String(a.Key, a.Value.AsString())
		}
	}
	return otelAttrs
}

func (t *traceOtelSpan) addEvent(name string, cfg *EventConfig) {
	t.tsOtel.AddEvent(name, trace.WithAttributes(toOtelAttributes(cfg.attributes...)...))
}

func (t *traceOtelSpan) recordError(err error, cfg *EventConfig) {
	t.tsOtel.RecordError(err, trace.WithAttributes(toOtelAttributes(cfg.attributes...)...))
}

func (t *traceOtelSpan) end() {
	t.tsShim.mu.RLock()
	defer t.tsShim.mu.RUnlock()
	t.tsOtel.SetName(t.tsShim.name)
	if t.tsShim.statusCode != 0 {
		t.tsOtel.SetStatus(codes.Code(t.tsShim.statusCode), t.tsShim.statusDesc)
	}
	if len(t.tsShim.attributes) > 0 {
		t.tsOtel.SetAttributes(toOtelAttributes(t.tsShim.attributes...)...)
	}
	t.tsOtel.End()
}

var tracer = otel.Tracer("gitea")

func (t *traceOtelStarter) start(ctx context.Context, traceSpan *TraceSpan, internalSpanIdx int) (context.Context, traceSpanInternal) {
	ctx, span := tracer.Start(ctx, traceSpan.name)
	return ctx, &traceOtelSpan{tsShim: traceSpan, tsOtel: span, internalSpanIdx: internalSpanIdx}
}

func init() {
	globalTraceStarters = append(globalTraceStarters, &traceOtelStarter{})
}
