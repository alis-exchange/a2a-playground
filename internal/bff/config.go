package bff

import "strings"

// Protocol identifies the A2A agent transport protocol.
type Protocol string

const (
	ProtocolGRPC    Protocol = "grpc"
	ProtocolJSONRPC Protocol = "jsonrpc"
)

// AgentConfig holds agent connection configuration.
type AgentConfig struct {
	URL      string
	Protocol Protocol
}

// NormalizeAgentURL validates and normalizes agent URL by protocol.
// For gRPC: strips http(s):// so "http://localhost:8080" becomes "localhost:8080".
// For JSON-RPC: requires http:// or https://; returns "" if invalid.
func NormalizeAgentURL(url string, proto Protocol) string {
	url = strings.TrimSpace(url)
	if url == "" {
		return ""
	}
	if proto == ProtocolGRPC {
		if strings.HasPrefix(url, "http://") {
			return strings.TrimPrefix(url, "http://")
		}
		if strings.HasPrefix(url, "https://") {
			return strings.TrimPrefix(url, "https://")
		}
		return url
	}
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	return ""
}
