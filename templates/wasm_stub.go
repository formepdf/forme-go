//go:build !forme_wasm

package templates

import "fmt"

// FormeRenderError is returned when the WASM engine returns an error.
type FormeRenderError struct {
	Message string
}

func (e *FormeRenderError) Error() string {
	return e.Message
}

func renderPDF(_ string) ([]byte, error) {
	return nil, fmt.Errorf("WASM rendering not available: build with -tags forme_wasm (requires forme.wasm)")
}
