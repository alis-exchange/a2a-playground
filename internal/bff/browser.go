package bff

import (
	"fmt"

	"github.com/pkg/browser"
)

// OpenBrowser opens the default browser to the given URL.
func OpenBrowser(url string) error {
	if err := browser.OpenURL(url); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}
	return nil
}
