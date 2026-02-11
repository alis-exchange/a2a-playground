package bff

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/alis-exchange/a2a-playground"
)

// distFS returns an fs.FS for serving the SPA.
// When dev is true, serves from rootDir/app/dist on disk; otherwise uses embedded files.
func distFS(dev bool, rootDir string) (fs.FS, error) {
	if dev {
		distPath := filepath.Join(rootDir, "app", "dist")
		if _, err := os.Stat(distPath); err != nil {
			return nil, err
		}
		return os.DirFS(distPath), nil
	}
	sub, err := fs.Sub(assets.FS, "app/dist")
	if err != nil {
		return nil, err
	}
	return sub, nil
}

// SPAHandler serves the SPA, falling back to index.html for client-side routes.
func SPAHandler(fsys fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(fsys))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		// Check if asset path (has extension) or is under /assets/
		if strings.HasPrefix(path, "/assets/") || hasExtension(path) {
			fileServer.ServeHTTP(w, r)
			return
		}
		// SPA fallback: serve index.html for client-side routes
		r.URL.Path = "/index.html"
		fileServer.ServeHTTP(w, r)
	})
}

// hasExtension returns true if the path has a file extension.
func hasExtension(path string) bool {
	base := filepath.Base(path)
	return strings.Contains(base, ".")
}
