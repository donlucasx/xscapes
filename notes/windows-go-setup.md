# Windows Go Setup for xscapes

xscapes on Windows needs **Go 1.27+** and **Windows 10 version 1809** or later (for ConPTY support). This guide walks through getting Go installed and verified so you can build and run the port.

---

## 1. Download Go

Open a browser and go to:

```
https://go.dev/dl/
```

Download the **Windows AMD64 MSI installer** — it's the `.msi` file, not the zip. The MSI puts Go on your PATH automatically; the zip does not.

> **Why the MSI?** The zip version leaves you to set `GOROOT`, `GOPATH`, and `%PATH%` by hand. The MSI does all of that in one click.

---

## 2. Install

1. Run the `.msi` file.
2. Accept the default install path (`C:\Go\`).
3. **Make sure "Add Go to PATH" is checked** — it usually is by default, but verify before clicking Finish.
4. When the installer finishes, **close and reopen your terminal** (PowerShell, cmd, or Git Bash). The new PATH won't exist in open windows.

---

## 3. Verify the install

Open a fresh terminal and run:

```powershell
go version
```

You should see something like:

```
go version go1.27.0 windows/amd64
```

If you get `go is not recognized`, your PATH didn't pick up. Fix it:

```powershell
# Check where Go lives
where.exe go

# If it's there but not on PATH, add it manually:
$env:PATH = "C:\Go\bin;$env:PATH"

# To make it permanent, run once as admin:
[Environment]::SetEnvironmentVariable("Path", $env:PATH, "Machine")
```

---

## 4. Check Windows build readiness

xscapes' Windows port needs ConPTY, which ships with Windows 10 1809+. Verify you're good:

```powershell
# Any version string starting with 10.0.1809 or higher is fine
[Environment]::OSVersion.VersionString
```

---

## 5. Build xscapes

Clone or copy the Windows port into a folder, then from that folder:

```powershell
cd C:\Users\celes\Documents\Hermes-Workspace\xscapes-main
go build -o xscapes.exe .
```

If it compiles cleanly you're set. If you get errors about missing packages, run:

```powershell
go mod download
go build -o xscapes.exe .
```

---

## 6. First run

```powershell
.\xscapes.exe install hermes --apply
.\xscapes.exe hermes
```

The first time a hook fires, Hermes asks for consent. Run `hermes hooks doctor` to answer once for all of them.

---

## Troubleshooting

| Problem | Fix |
|---|---|
| `go: command not found` | Restart terminal; re-check PATH above |
| `go version` shows < 1.27 | Download the latest from go.dev/dl/ and run the MSI again |
| Build fails with `syscall` errors | You need Windows 10 1809+; check with `[Environment]::OSVersion` |
| `xscapes.exe` won't start — "blocked by Windows" | Right-click the exe → Properties → check "Unblock" at the bottom |

---

## What changed vs macOS/Linux

The core animation code is identical. The port replaces four Unix mechanisms with Windows equivalents:

- **PTY** → `CreatePseudoConsole` (ConPTY) instead of `/dev/ptmx`
- **Terminal resize** → polls `GetConsoleScreenBufferInfo` every 250ms instead of SIGWINCH
- **Raw input** → `SetConsoleMode` flags instead of `stty raw`
- **Audio cues** → `PlaySoundW` from `winmm.dll` instead of `afplay`/`paplay`

Full details are in [notes/windows-conpty.md](../notes/windows-conpty.md).
