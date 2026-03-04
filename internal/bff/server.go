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

	factory := NewProxyFactory(cfg.AgentURL, cfg.Protocol)
	// Prime the cache and get the A2A path prefix
	a2aPath, _ := factory.GetHandler(cfg.AgentURL, cfg.Protocol)

	// Per-request: read agent config from headers (fallback to CLI default), get proxy handler, inject agent headers, serve
	a2aWithHeaders := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agentURL, protocol := AgentConfigFromRequest(r)
		_, handler := factory.GetHandler(agentURL, protocol)
		headers := ExtractAgentHeaders(r)
		reqCtx := WithAgentHeaders(r.Context(), headers)
		handler.ServeHTTP(w, r.WithContext(reqCtx))
	})

	mux := mux.NewRouter()
	mux.HandleFunc("/auth/callback", HandleAuthCallback).Methods(http.MethodGet)
	mux.HandleFunc("/auth/refresh", HandleAuthRefresh).Methods(http.MethodPost)
	mux.HandleFunc("/auth/start", HandleAuthStart).Methods(http.MethodPost)
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
