package bff

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// AgentHeadersKey is the context key for agent headers to forward to the agent.
type AgentHeadersKey struct{}

const (
	agentHeadersHeader   = "X-A2A-Agent-Headers"
	agentURLHeader       = "X-A2A-Agent-URL"
	agentProtocolHeader  = "X-A2A-Agent-Protocol"
)

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
// (map[string]string), and merges in the request's Authorization header so the BFF
// forwards OAuth Bearer tokens to the agent. Returns nil if no headers to forward.
func ExtractAgentHeaders(r *http.Request) map[string]string {
	var m map[string]string
	if h := r.Header.Get(agentHeadersHeader); h != "" {
		if err := json.Unmarshal([]byte(h), &m); err != nil {
			m = nil
		}
	}
	if m == nil {
		m = make(map[string]string)
	}
	if auth := r.Header.Get("Authorization"); auth != "" {
		m["Authorization"] = auth
	}
	if len(m) == 0 {
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

// AgentConfigFromRequest reads X-A2A-Agent-URL and X-A2A-Agent-Protocol from the request.
// Returns (url, protocol). Empty url means use server default.
func AgentConfigFromRequest(r *http.Request) (url string, protocol Protocol) {
	url = strings.TrimSpace(r.Header.Get(agentURLHeader))
	protoStr := strings.TrimSpace(r.Header.Get(agentProtocolHeader))
	if protoStr == string(ProtocolJSONRPC) {
		return url, ProtocolJSONRPC
	}
	return url, ProtocolGRPC
}
