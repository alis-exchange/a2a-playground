// Package assets provides the embedded frontend build output (app/dist).
// Build with `pnpm build` in app/, or use --dev to serve from disk.
package assets

import "embed"

//go:embed all:app/dist
var FS embed.FS
