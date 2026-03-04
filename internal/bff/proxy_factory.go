package bff

import (
	"net/http"
	"sync"
)

// ProxyFactory returns an http.Handler for a given (agentURL, protocol) by
// getting or creating a proxy and caching its handler.
type ProxyFactory struct {
	defaultURL      string
	defaultProtocol Protocol
	mu              sync.Mutex
	handlers        map[string]http.Handler
	path            string
}

// NewProxyFactory creates a factory that uses defaultURL and defaultProtocol
// when the request does not specify agent config.
func NewProxyFactory(defaultURL string, defaultProtocol Protocol) *ProxyFactory {
	return &ProxyFactory{
		defaultURL:      defaultURL,
		defaultProtocol: defaultProtocol,
		handlers:        make(map[string]http.Handler),
	}
}

// cacheKey returns a key for the (url, protocol) pair.
func (f *ProxyFactory) cacheKey(agentURL string, protocol Protocol) string {
	return agentURL + "\x00" + string(protocol)
}

// GetHandler returns the path prefix and the http.Handler for the given agentURL and protocol.
// If agentURL is empty, defaultURL and defaultProtocol are used.
// The returned handler is safe for concurrent use.
func (f *ProxyFactory) GetHandler(agentURL string, protocol Protocol) (path string, handler http.Handler) {
	if agentURL == "" {
		agentURL = f.defaultURL
		protocol = f.defaultProtocol
	}
	normalized := NormalizeAgentURL(agentURL, protocol)
	if normalized == "" {
		// Invalid combo; fall back to default
		agentURL = f.defaultURL
		protocol = f.defaultProtocol
		normalized = f.defaultURL
	}
	key := f.cacheKey(normalized, protocol)
	f.mu.Lock()
	defer f.mu.Unlock()
	if h, ok := f.handlers[key]; ok {
		return f.path, h
	}
	var proxy A2AServiceHandler
	if protocol == ProtocolJSONRPC {
		proxy = NewJSONRPCProxy(normalized)
	} else {
		proxy = NewGrpcProxy(normalized)
	}
	p, h := proxy.Handler()
	if f.path == "" {
		f.path = p
	}
	f.handlers[key] = h
	return f.path, h
}
