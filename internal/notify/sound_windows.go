//go:build windows

package notify

import (
	"fmt"
	"syscall"
	"unsafe"
)

// PLAYING A CUE ON WINDOWS.
//
// There is no afplay and no paplay, and shelling out to PowerShell costs about
// 200 ms of process start -- per knock, which is the one place latency is the
// whole point. So the file is handed to winmm's PlaySound directly: one call,
// asynchronous, returns immediately, and the process only has to stay alive
// for the length of the cue (which the caller already guarantees).
//
// The player is named "winmm" everywhere a command line would be named, so
// `xscapes notify` and Describe read the same as on the other platforms.

const (
	sndAsync     = 0x0001
	sndFilename  = 0x00020000
	sndNoDefault = 0x0002
)

var procPlaySound = syscall.NewLazyDLL("winmm.dll").NewProc("PlaySoundW")

// playerName is what Describe calls this rung's player.
const playerName = "winmm"

func playSound(path string) error {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	r, _, e := procPlaySound.Call(uintptr(unsafe.Pointer(p)), 0, sndAsync|sndFilename|sndNoDefault)
	if r == 0 {
		return fmt.Errorf("PlaySound %s: %v", path, e)
	}
	return nil
}

// directPlay is this platform's way of playing a file with no player binary.
func directPlay() func(string) error { return playSound }

// playerBin is empty: nothing external is run.
func playerBin() string { return playerName }

// systemCue names the sounds Windows already has, used only when the embedded
// cues cannot be written to disk. The three are the closest the OS gets to a
// question, a finish and a failure.
func systemCue() (ask, done, worried string, ok bool) {
	dir := `C:\Windows\Media\`
	a, okA := firstFile(dir+"Windows Notify System Generic.wav", dir+"chimes.wav", dir+"ding.wav")
	d, okD := firstFile(dir+"Windows Notify Calendar.wav", dir+"Windows Background.wav", dir+"ding.wav")
	w, okW := firstFile(dir+"Windows Critical Stop.wav", dir+"Windows Error.wav", dir+"chord.wav")
	if !okW {
		w = d
	}
	return a, d, w, okA && okD
}

func systemCueName() string { return "Windows system sounds" }
