// Package main - a2a-playground CLI. Run: ./a2a-playground --agent-url=localhost:8080
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

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
	version  string // set via -ldflags at build
	agentURL string
	port     int
	noOpen   bool
	dev      bool
)

var rootCmd = &cobra.Command{
	Use:   "a2a-playground",
	Short: "A2A agent playground - serves the frontend and proxies to an A2A agent",
	RunE:  runServe,
}

func init() {
	rootCmd.Flags().StringVar(&agentURL, "agent-url", "localhost:8080", "gRPC URL of the A2A agent")
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

	appDir, err := findAppDir()
	if err != nil {
		return err
	}

	cfg := bff.ServerConfig{
		Port:        port,
		AgentURL:    agentURL,
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
	fmt.Printf("Serving at %s (agent: %s)\n", url, agentURL)

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
