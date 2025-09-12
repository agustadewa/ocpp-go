package ocppj

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/trace"
)

// SpanContext holds both context and span for request-response correlation
type SpanContext struct {
	Ctx  context.Context
	Span trace.Span
}

// ContextMap provides thread-safe storage for request contexts and spans
type ContextMap struct {
	data sync.Map
}

// NewContextMap creates a new ContextMap instance
func NewContextMap() *ContextMap {
	return &ContextMap{}
}

// Store saves a context and span pair for a given request ID
func (cm *ContextMap) Store(key string, ctx context.Context, span trace.Span) {
	cm.data.Store(key, SpanContext{Ctx: ctx, Span: span})
}

// Delete removes a context and span pair for a given request ID
func (cm *ContextMap) Delete(key string) {
	cm.data.Delete(key)
}

// LoadAndDelete retrieves and removes a context and span pair for a given request ID
func (cm *ContextMap) LoadAndDelete(key string) (context.Context, trace.Span, bool) {
	if value, exists := cm.data.LoadAndDelete(key); exists {
		spanCtx := value.(SpanContext)
		return spanCtx.Ctx, spanCtx.Span, true
	}
	return nil, nil, false
}
