// Package main - a2a-playground CLI. Run: ./a2a-playground --agent-url=localhost:8080
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/alis-exchange/a2a-playground/internal/bff"
	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var (
	version   string // set via -ldflags at build
	agentURL  string
	useJSONRPC bool
	port      int
	noOpen    bool
	dev       bool
)

var rootCmd = &cobra.Command{
	Use:   "a2a-playground",
	Short: "A2A agent playground - serves the frontend and proxies to an A2A agent",
	RunE:  runServe,
}

func init() {
	rootCmd.Flags().StringVar(&agentURL, "agent-url", "localhost:8080", "Agent endpoint: for gRPC use host:port; for JSON-RPC use full URL (e.g. http://localhost:8080/jsonrpc)")
	rootCmd.Flags().BoolVar(&useJSONRPC, "jsonrpc", false, "Use JSON-RPC transport instead of gRPC")
	if version != "" {
		rootCmd.Version = version
	}
	rootCmd.Flags().IntVar(&port, "port", 3000, "HTTP port for the BFF")
	rootCmd.Flags().BoolVar(&noOpen, "no-open", false, "Do not open the browser on start")
	rootCmd.Flags().BoolVar(&dev, "dev", false, "Serve from app/dist on disk instead of embedded files")
}

// runServe starts the BFF server and blocks until interrupt.
func runServe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Validate and normalize agent-url by protocol
	proto := bff.ProtocolGRPC
	if useJSONRPC {
		proto = bff.ProtocolJSONRPC
	}
	normalizedURL := normalizeAgentURL(agentURL, proto)
	if normalizedURL == "" {
		return fmt.Errorf("invalid agent-url %q for protocol %s: JSON-RPC requires http:// or https:// scheme", agentURL, proto)
	}

	appDir, err := findAppDir()
	if err != nil {
		return err
	}

	cfg := bff.ServerConfig{
		Port:        port,
		AgentURL:    normalizedURL,
		Protocol:    proto,
		Dev:         dev,
		NoOpen:      noOpen,
		AppDir:      appDir,
		OpenBrowser: !noOpen,
	}

	srv, err := bff.NewServer(ctx, cfg)
	if err != nil {
		return err
	}

	if err := srv.Start(); err != nil {
		return err
	}

	url := fmt.Sprintf("http://localhost:%d", port)
	fmt.Printf("Serving at %s (agent: %s, protocol: %s)\n", url, normalizedURL, proto)

	if cfg.OpenBrowser {
		_ = bff.OpenBrowser(url)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	fmt.Println("Shutting down...")
	shutdownCtx := context.Background()
	return srv.Shutdown(shutdownCtx)
}

// normalizeAgentURL validates and normalizes agent-url by protocol.
// For gRPC: strips http(s):// so "http://localhost:8080" becomes "localhost:8080".
// For JSON-RPC: requires http:// or https://; returns "" if invalid.
func normalizeAgentURL(url string, proto bff.Protocol) string {
	url = strings.TrimSpace(url)
	if url == "" {
		return ""
	}
	if proto == bff.ProtocolGRPC {
		// Strip scheme for gRPC
		if strings.HasPrefix(url, "http://") {
			return strings.TrimPrefix(url, "http://")
		}
		if strings.HasPrefix(url, "https://") {
			return strings.TrimPrefix(url, "https://")
		}
		return url // bare host:port is fine
	}
	// JSON-RPC: must have http:// or https://
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	// Bare host:port for JSON-RPC - treat as http by default (plan says reject, but being helpful)
	// Plan says: "reject bare host:port when --jsonrpc"
	return "" // reject bare host:port for JSON-RPC
}

// findAppDir locates the project root (containing app/) for serving app/dist in dev mode.
func findAppDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// Prefer app dir next to the binary (works when run from repo root)
	for _, base := range []string{wd, filepath.Dir(wd)} {
		appDir := filepath.Join(base, "app")
		if st, err := os.Stat(appDir); err == nil && st.IsDir() {
			return base, nil
		}
	}
	return wd, nil
}
