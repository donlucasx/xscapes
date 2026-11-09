//go:build !ghosttyvt

package host

import "errors"

// EngineAvailable reports whether this binary carries the Ghostty engine.
const EngineAvailable = false

func (h *Host) runVT() error {
	return errors.New("this xscapes was built without the Ghostty engine (go build -tags ghosttyvt)")
}
