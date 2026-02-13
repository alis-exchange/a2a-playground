package bff

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// ServerConfig holds configuration for the BFF server.
type ServerConfig struct {
	Port        int
	AgentURL    string
	Protocol    Protocol
	Dev         bool
	NoOpen      bool
	AppDir      string
	OpenBrowser bool
}

// Server represents the BFF HTTP server.
type Server struct {
	cfg    ServerConfig
	server *http.Server
}

// NewServer creates and configures the BFF server.
func NewServer(ctx context.Context, cfg ServerConfig) (*Server, error) {
	fsys, err := distFS(cfg.Dev, cfg.AppDir)
	if err != nil {
		return nil, fmt.Errorf("dist fs: %w", err)
	}

	var proxy A2AServiceHandler
	if cfg.Protocol == ProtocolJSONRPC {
		proxy = NewJSONRPCProxy(cfg.AgentURL)
	} else {
		proxy = NewGrpcProxy(cfg.AgentURL)
	}

	a2aPath, a2aHandler := proxy.Handler()

	// Middleware to inject X-A2A-Agent-Headers into request context for proxy forwarding
	a2aWithHeaders := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := ExtractAgentHeaders(r)
		ctx := WithAgentHeaders(r.Context(), headers)
		a2aHandler.ServeHTTP(w, r.WithContext(ctx))
	})

	mux := mux.NewRouter()
	mux.PathPrefix(a2aPath).Handler(a2aWithHeaders)
	mux.PathPrefix("/").Handler(SPAHandler(fsys))

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return &Server{
		cfg:    cfg,
		server: srv,
	}, nil
}

// Addr returns the server address.
func (s *Server) Addr() string {
	return s.server.Addr
}

// Start starts the server in a goroutine.
func (s *Server) Start() error {
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return s.server.Shutdown(shutdownCtx)
}
