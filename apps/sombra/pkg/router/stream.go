package router

import (
	"context"
	"fmt"
)

// Streamer is an optional interface that model adapters may implement to
// support token-level streaming. If an adapter does not implement Streamer,
// Router.SendStream falls back to buffered Send() and emits the full response
// as a single delta — protocol-correct but without real streaming latency.
type Streamer interface {
	// SendStream calls onDelta for each text chunk as it arrives from the
	// upstream model. The call to onDelta must be sequential (not concurrent).
	// Returning a non-nil error from onDelta aborts the stream.
	SendStream(ctx context.Context, messages []Message, opts ModelOpts, onDelta func(string) error) error
}

// SendStream routes the messages to the named model and calls onDelta for
// each text chunk as it arrives.
//
// If the selected adapter implements Streamer, true token-level streaming is
// used. Otherwise the adapter's Send() is called and the complete response is
// delivered as a single onDelta call — transparent fallback for Gemini and
// any future adapter that hasn't implemented streaming yet.
func (r *Router) SendStream(ctx context.Context, modelName string, messages []Message, opts ModelOpts, onDelta func(string) error) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if modelName == "" {
		modelName = r.fallback
	}

	adapter, ok := r.adapters[modelName]
	if !ok {
		return fmt.Errorf("router: unknown model %q (registered: %v)", modelName, r.modelNames())
	}

	if !isApprovedDomain(adapter.Endpoint(), r.allowedDomains) {
		return fmt.Errorf("OCULTAR Zero-Egress Block: domain %q not in approved list (fail-closed)", adapter.Endpoint())
	}

	if streamer, ok := adapter.(Streamer); ok {
		return streamer.SendStream(ctx, messages, opts, onDelta)
	}

	// Fallback: buffered response emitted as one delta.
	text, err := adapter.Send(ctx, messages, opts)
	if err != nil {
		return err
	}
	return onDelta(text)
}
