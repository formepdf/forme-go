//go:build forme_wasm

package templates

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed forme.wasm
var wasmBytes []byte

// FormeRenderError is returned when the WASM engine returns an error.
type FormeRenderError struct {
	Message string
}

func (e *FormeRenderError) Error() string {
	return e.Message
}

type formeEngine struct {
	runtime wazero.Runtime
	mod     wazero.CompiledModule
	mu      sync.Mutex
}

var (
	engine     *formeEngine
	engineOnce sync.Once
	engineErr  error
)

func initEngine() (*formeEngine, error) {
	engineOnce.Do(func() {
		ctx := context.Background()

		rt := wazero.NewRuntime(ctx)
		wasi_snapshot_preview1.MustInstantiate(ctx, rt)

		compiled, err := rt.CompileModule(ctx, wasmBytes)
		if err != nil {
			engineErr = fmt.Errorf("failed to compile WASM module: %w", err)
			rt.Close(ctx)
			return
		}

		engine = &formeEngine{
			runtime: rt,
			mod:     compiled,
		}
	})
	return engine, engineErr
}

func renderPDF(jsonStr string) ([]byte, error) {
	eng, err := initEngine()
	if err != nil {
		return nil, err
	}

	eng.mu.Lock()
	defer eng.mu.Unlock()

	ctx := context.Background()

	// Instantiate a fresh module for each render (WASM is single-use due to static state)
	mod, err := eng.runtime.InstantiateModule(ctx, eng.mod, wazero.NewModuleConfig().
		WithStdout(nil).WithStderr(nil).WithName(""))
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate WASM module: %w", err)
	}
	defer mod.Close(ctx)

	alloc := mod.ExportedFunction("forme_alloc")
	dealloc := mod.ExportedFunction("forme_dealloc")
	render := mod.ExportedFunction("forme_render_pdf")
	resultPtr := mod.ExportedFunction("forme_get_result_ptr")
	resultLen := mod.ExportedFunction("forme_get_result_len")
	errorPtr := mod.ExportedFunction("forme_get_error_ptr")
	errorLen := mod.ExportedFunction("forme_get_error_len")
	freeResult := mod.ExportedFunction("forme_free_result")

	jsonBytes := []byte(jsonStr)
	length := uint64(len(jsonBytes))

	// Allocate input buffer
	results, err := alloc.Call(ctx, length, 1)
	if err != nil {
		return nil, fmt.Errorf("forme_alloc failed: %w", err)
	}
	inputPtr := results[0]
	if inputPtr == 0 {
		return nil, &FormeRenderError{Message: "Failed to allocate WASM memory for input"}
	}

	// Write JSON into WASM memory
	if !mod.Memory().Write(uint32(inputPtr), jsonBytes) {
		dealloc.Call(ctx, inputPtr, length, 1)
		return nil, &FormeRenderError{Message: "Failed to write to WASM memory"}
	}

	// Call render
	results, err = render.Call(ctx, inputPtr, length)
	if err != nil {
		dealloc.Call(ctx, inputPtr, length, 1)
		return nil, fmt.Errorf("forme_render_pdf failed: %w", err)
	}
	status := results[0]

	if status != 0 {
		// Read error
		ePtrResults, _ := errorPtr.Call(ctx)
		eLenResults, _ := errorLen.Call(ctx)
		ePtr := uint32(ePtrResults[0])
		eLen := uint32(eLenResults[0])
		errMsg := "Unknown render error"
		if ePtr != 0 && eLen > 0 {
			if data, ok := mod.Memory().Read(ePtr, eLen); ok {
				errMsg = string(data)
			}
		}
		dealloc.Call(ctx, inputPtr, length, 1)
		return nil, &FormeRenderError{Message: errMsg}
	}

	// Read result
	rPtrResults, _ := resultPtr.Call(ctx)
	rLenResults, _ := resultLen.Call(ctx)
	rPtr := uint32(rPtrResults[0])
	rLen := uint32(rLenResults[0])

	if rPtr == 0 || rLen == 0 {
		dealloc.Call(ctx, inputPtr, length, 1)
		return nil, &FormeRenderError{Message: "Render returned empty result"}
	}

	pdfBytes, ok := mod.Memory().Read(rPtr, rLen)
	if !ok {
		dealloc.Call(ctx, inputPtr, length, 1)
		return nil, &FormeRenderError{Message: "Failed to read PDF from WASM memory"}
	}

	// Copy before freeing
	result := make([]byte, len(pdfBytes))
	copy(result, pdfBytes)

	freeResult.Call(ctx)
	dealloc.Call(ctx, inputPtr, length, 1)

	return result, nil
}
