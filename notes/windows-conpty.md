# Windows: what was measured, and what is different

*(Port built and verified on Windows 11, 2026-11-11. Everything below is a
measurement from this machine, not a reading of the documentation.)*

The scape is macOS-and-Linux shaped in one place only: the host that runs the
agent in the top rows of the window. Everything above it — the protocol, the
reducer, the scenes, the canvas — is portable Go and needed no change. What
follows is the five pieces that did.

## 1. The pty is a pseudoconsole (ConPTY)

`internal/host/pty_windows.go`. Windows has no `/dev/ptmx`; since Windows 10
1809 it has `CreatePseudoConsole`, which hands back an HPCON that a child is
attached to by **process attribute**, with the parent talking to it through two
`CreatePipe` pipes.

Three things had to be right, and every one of them fails silently:

| detail | wrong value | what you see |
|---|---|---|
| `UpdateProcThreadAttribute(PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE, …)` | `&hpc` instead of `hpc` | the child launches fine and is attached to the **parent's** console: its output goes to the real screen and the pipes stay silent forever. No error anywhere. |
| `STARTF_USESTDHANDLES` with three NULL handles | omitted | the same failure, worn differently: the child inherits the parent's std handles |
| reading the pipe | `os.NewFile(...).Read` | a `Read` that never returns: a `CreatePipe` handle is synchronous and Go's poller cannot drive it. `syscall.ReadFile` on the raw handle works. |

`exec.Cmd` cannot create this process: `syscall.StartProcess` builds its own
attribute list and there is no exported way to add a third attribute. So
`startAgent` calls `CreateProcessW` itself and hands the pid back as
`cmd.Process` (`os.FindProcess`), after which `cmd.Wait()` and the whole of
`Host.Run`'s exit path work unchanged.

The command line is quoted to `CommandLineToArgvW` rules, and a `.cmd`/`.bat`
target is not executable by `CreateProcess` at all — it needs `cmd /c`.

## 2. There is no SIGWINCH and no ioctl

`ttysize_windows.go` reads `GetConsoleScreenBufferInfo` (the *window*, never the
buffer: a scrolled console's buffer is wider than the screen) and polls it every
250 ms to poke the same channel the unix build fills from `SIGWINCH`. The live
loop asks for the size every frame regardless — that is what makes a drag
coalesce.

## 3. Raw mode is two console modes

`rawmode_windows.go`. `ENABLE_ECHO_INPUT`/`ENABLE_LINE_INPUT` off so the agent
draws its own input; `ENABLE_PROCESSED_INPUT` off so Ctrl-C arrives as `0x03`
rather than killing the host (inside an agent it means "interrupt this turn");
`ENABLE_VIRTUAL_TERMINAL_INPUT` on so arrows arrive as VT sequences. Output gets
`ENABLE_VIRTUAL_TERMINAL_PROCESSING`, without which a Windows console prints
every escape the scape emits as text.

A non-console handle (a pipe, a redirect, CI) is not an error: there is nothing
to switch, and the host still passes bytes through.

## 4. The event socket is the spool

Go's `net` has no AF_UNIX on Windows, so `sendSock` fails by construction and
every event takes the JSON-lines spool the bus already follows — the same
two-path design as the other platforms with the fast path absent. The cost is
`spoolPoll`: 250 ms between an event and the scene noticing it, against a 12 fps
scape. `internal/event/sock_{unix,windows}.go`.

## 5. Sound is winmm

`internal/notify/sound_windows.go`. No `afplay`/`paplay`, and PowerShell costs
~200 ms of process start per knock — the one place latency is the whole point.
`PlaySoundW(path, NULL, SND_FILENAME|SND_ASYNC)` starts the embedded WAV and
returns. The fallback rung is `C:\Windows\Media` instead of `/System/Library`.

## Where Hermes Agent keeps its config

Not `~/.hermes` on Windows. Hermes's own rule is `$HERMES_HOME`, else the
platform default `%LOCALAPPDATA%\hermes`, and the desktop app runs the agent
with a **profile** that is a whole home of its own under `profiles/<name>`. On
this machine that is `%LOCALAPPDATA%\hermes\profiles\def\config.yaml`, and hooks
written anywhere else would never fire. `install hermes` resolves it that way
(`hermesConfigPath` in `install_agents.go`): `$HERMES_HOME`, else
`$HERMES_PROFILE`, else the only profile on the machine, else the platform
default.

The agent splits a hook command with its own Windows-safe tokeniser
(`split_command_line`, `posix=False`, one layer of matching quotes stripped) and
runs it with `shell=False`, so a path with a space is quoted in the command
string itself and the path's backslashes survive.

## What the test suite says

`go test ./internal/... .` passes on Windows. Four tests are skipped, each with
its reason in the test: three measure what **Terminal.app and Ghostty** do to
the band on a resize and how they answer the `ESC[6n` cursor query — ConPTY
anchors rows differently and answers differently, so the Windows host needs its
own measurements before it can be held to those — and two trace tests need a
POSIX directory mode to make a path unwritable.

The port's own proof is `internal/host/conpty_windows_test.go`: a child gets a
pseudoconsole, sees the width the host set (`mode con` reports 60), its output
comes back through `master`, and bytes written to `master` reach its stdin.
