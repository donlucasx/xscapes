//go:build !windows

package notify

// directPlay is nil here: these platforms play a file by running a player
// binary (afplay, paplay), which playerBin names. Windows has no such binary
// and plays the file itself; see sound_windows.go.
func directPlay() func(string) error { return nil }
