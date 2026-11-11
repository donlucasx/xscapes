//go:build !windows

package notify

import "runtime"

// systemCue names the sounds the OS already has: the rung xscapes played
// before it had a voice of its own.
//
// They are kept as the fallback and not as the default for the reason our
// own cues exist at all -- Glass has been every Mac's notification for twenty
// years, so it cannot be xscapes' -- but a borrowed sound that plays beats an
// own sound that cannot be written.
//
// Verified present rather than assumed: a missing file makes afplay exit
// non-zero and the knock is silently lost. The worried knock borrows the OS's
// own error sound where there is one and falls back to the done sound where
// there is not, so this rung is never the reason a finish goes silent.
func systemCue() (ask, done, worried string, ok bool) {
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
		w, okW := firstFile(
			"/System/Library/Sounds/Basso.aiff",
			"/System/Library/Sounds/Sosumi.aiff",
		)
		if !okW {
			w = d
		}
		return a, d, w, okA && okD
	case "linux":
		a, okA := firstFile(
			"/usr/share/sounds/freedesktop/stereo/message.oga",
			"/usr/share/sounds/freedesktop/stereo/bell.oga",
		)
		d, okD := firstFile(
			"/usr/share/sounds/freedesktop/stereo/complete.oga",
			"/usr/share/sounds/freedesktop/stereo/message.oga",
		)
		w, okW := firstFile(
			"/usr/share/sounds/freedesktop/stereo/dialog-error.oga",
			"/usr/share/sounds/freedesktop/stereo/bell.oga",
		)
		if !okW {
			w = d
		}
		return a, d, w, okA && okD
	}
	return "", "", "", false
}

// playerBin is the command that can open a file and make a noise with it.
//
// Both of them handle WAV as well as the platform's own format, which is what
// lets one binary serve both the embedded cues and the system sounds.
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
