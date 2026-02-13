package bff

import (
	"context"
	"encoding/json"
	"net/http"
)

// AgentHeadersKey is the context key for agent headers to forward to the agent.
type AgentHeadersKey struct{}

const agentHeadersHeader = "X-A2A-Agent-Headers"

// AgentHeadersFromContext returns the agent headers from context, or nil if not set.
func AgentHeadersFromContext(ctx context.Context) map[string]string {
	v := ctx.Value(AgentHeadersKey{})
	if v == nil {
		return nil
	}
	m, _ := v.(map[string]string)
	return m
}

// ExtractAgentHeaders reads X-A2A-Agent-Headers from the request, parses it as JSON
// (map[string]string), and returns the result. Returns nil if the header is absent or invalid.
func ExtractAgentHeaders(r *http.Request) map[string]string {
	h := r.Header.Get(agentHeadersHeader)
	if h == "" {
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(h), &m); err != nil {
		return nil
	}
	return m
}

// WithAgentHeaders returns a context with the given headers attached.
func WithAgentHeaders(ctx context.Context, headers map[string]string) context.Context {
	if headers == nil {
		return ctx
	}
	return context.WithValue(ctx, AgentHeadersKey{}, headers)
}
