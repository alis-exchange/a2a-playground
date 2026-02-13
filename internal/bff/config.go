package bff

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
