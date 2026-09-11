package notify

import (
	"bytes"
	_ "embed"
	"os"
	"path/filepath"
	"runtime"

	"github.com/donlucasx/xscapes/internal/event"
)

// The two cues xscapes ships with, synthesised by notes/s28-sound/make.py and
// picked by Lucas on 2026-09-11 from four auditioned families.
//
// They are EMBEDDED rather than installed beside the binary because "single
// static binary" is a locked decision in the brief, and because a cue that can
// go missing is worse than no cue at all: the player would exit non-zero and
// the knock would vanish with nothing said. Embedded bytes cannot be half
// there.
//
// Ask RISES and done FALLS, which is the same distinction Bubble and
// DoneBubble draw on screen -- shape, not pitch, so the pair stays readable
// through a laptop speaker the way the bubbles stay readable without colour.
//
//go:embed sounds/ask.wav
var askWAV []byte

//go:embed sounds/done.wav
var doneWAV []byte

// maxCueBytes is what a cue is allowed to add to the binary.
//
// At the format these are written in -- 16-bit mono, 44.1 kHz -- one second is
// 88,200 bytes, so 128 KB is about 1.5 seconds. The design rule the cues were
// written to is "under 500 ms", so this is three times the longest cue that
// could ever be in taste, and the pair still costs well under 1% of a ~6.6 MB
// binary. It exists to catch a wrong file being embedded (a stereo master, a
// whole ambience bed), not to shave bytes.
const maxCueBytes = 128 << 10

// cueDir is where the embedded bytes get written so a player can open them.
//
// afplay and paplay take a PATH, not a stream, so the bytes have to land on
// disk somewhere. NOT /tmp: a reboot on 2026-09-06 destroyed nine trace files
// there and cost a whole occurrence, and a cue that silently stops existing
// after a reboot is the same defect with a quieter failure. Under Home() it
// lives exactly as long as the rest of the scape's state, and the user can go
// and listen to it.
func cueDir() (string, error) {
	h, err := event.Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, "sounds"), nil
}

// materialiseCues writes both embedded cues out and returns their paths.
//
// ok is false for any reason at all -- no home, an unwritable directory, a
// full disk. The caller falls to the next rung rather than going silent, which
// is the whole point of doing this at New() time: a failure here is discovered
// once, at startup, and never inside Play.
func materialiseCues() (ask, done string, ok bool) {
	dir, err := cueDir()
	if err != nil {
		return "", "", false
	}
	// 0700 for the same reason the run directory is: everything under Home()
	// is this user's session state and nothing here is meant to be shared.
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", "", false
	}
	a, err := placeCue(filepath.Join(dir, "ask.wav"), askWAV)
	if err != nil {
		return "", "", false
	}
	d, err := placeCue(filepath.Join(dir, "done.wav"), doneWAV)
	if err != nil {
		return "", "", false
	}
	return a, d, true
}

// placeCue makes path hold want, writing only if it does not already.
//
// The freshness check is SIZE FIRST, THEN THE BYTES THEMSELVES -- not a hash,
// and not the size alone. Size alone is what a retune would defeat: the likely
// next edit to a cue is a different timbre at the same 260 ms, which leaves the
// file byte-for-byte the wrong sound at exactly the right length, forever, with
// nothing to notice it. A hash would catch that, but the embedded bytes are
// already in memory, so comparing them directly is the same read, exact instead
// of probabilistic, and needs no sidecar file to store the digest in. The stat
// is kept in front of it purely so the common case -- unchanged, every scape
// start -- costs one syscall instead of a 30 KB read.
func placeCue(path string, want []byte) (string, error) {
	if cueIsCurrent(path, want) {
		return path, nil
	}
	// Written to a temp file and RENAMED, because two scapes starting at the
	// same moment is normal -- he runs one per project -- and a plain
	// os.WriteFile lets the second open the file the first is halfway through.
	// Rename within a directory is atomic, so a reader sees the old cue or the
	// new one and never a truncated header.
	tmp, err := os.CreateTemp(filepath.Dir(path), ".cue-*.wav")
	if err != nil {
		return "", err
	}
	// A no-op once the rename has happened; it only fires on the error paths,
	// where it is what stops a failed write leaving litter behind.
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(want); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return "", err
	}
	return path, nil
}

func cueIsCurrent(path string, want []byte) bool {
	st, err := os.Stat(path)
	if err != nil || st.IsDir() || st.Size() != int64(len(want)) {
		return false
	}
	got, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return bytes.Equal(got, want)
}

// systemCue names the sounds the OS already has: the rung xscapes played
// before it had a voice of its own.
//
// They are kept as the fallback and not as the default for the reason the
// droplet exists at all -- Glass has been every Mac's notification for twenty
// years, so it cannot be xscapes' -- but a borrowed sound that plays beats an
// own sound that cannot be written.
//
// Verified present rather than assumed: a missing file makes afplay exit
// non-zero and the knock is silently lost.
func systemCue() (ask, done string, ok bool) {
	switch runtime.GOOS {
	case "darwin":
		a, okA := firstFile(
			"/System/Library/Sounds/Glass.aiff",
			"/System/Library/Sounds/Tink.aiff",
		)
		d, okD := firstFile(
			"/System/Library/Sounds/Submarine.aiff",
			"/System/Library/Sounds/Purr.aiff",
		)
		return a, d, okA && okD
	case "linux":
		a, okA := firstFile(
			"/usr/share/sounds/freedesktop/stereo/message.oga",
			"/usr/share/sounds/freedesktop/stereo/bell.oga",
		)
		d, okD := firstFile(
			"/usr/share/sounds/freedesktop/stereo/complete.oga",
			"/usr/share/sounds/freedesktop/stereo/message.oga",
		)
		return a, d, okA && okD
	}
	return "", "", false
}

// playerBin is the command that can open a file and make a noise with it.
//
// Both of them handle WAV as well as the platform's own format, which is what
// lets one binary serve both the embedded droplet and the system sounds.
func playerBin() string {
	switch runtime.GOOS {
	case "darwin":
		if have("afplay") {
			return "afplay"
		}
	case "linux":
		if have("paplay") {
			return "paplay"
		}
	}
	return ""
}

// systemCueName is what Describe calls the borrowed rung, so "which sounds am
// I actually hearing" has an answer that names the source.
func systemCueName() string {
	if runtime.GOOS == "darwin" {
		return "macOS system sounds"
	}
	return "freedesktop sounds"
}
