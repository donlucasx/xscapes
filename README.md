# asciiscapes

Waiting for an agent is dead time. You fire off a prompt, the terminal goes
quiet, you tab away, and then you forget to come back. asciiscapes gives that
dead time a face: a small living shoreline that runs beside your agent, rises
while it works, and knocks when it wants you.

![one turn, driven by real events](assets/frames/wired.png)

It is two things. Underneath there is an **event protocol** with pluggable
adapters, which is the part that generalises to any agent. On top there is a
**reference scape**, which is what that protocol looks like when you give it a
sea and a cat.

```
asciiscapes claude
```

That is the whole thing. Claude Code in the left pane, the scape in the right.

## Install

```sh
git clone https://github.com/donlucasx/asciiscapes && cd asciiscapes
go build -o asciiscapes .

./asciiscapes install claude          # prints a plan, writes nothing
./asciiscapes install claude --apply  # writes the hooks, after a backup
```

The only dependency is tmux, for the side by side layout:

```sh
brew install tmux
```

Without tmux it opens a second Terminal window instead. Uninstall restores your
settings byte for byte.

## What you are looking at

Everything on screen means one thing, and no two things share a channel.

| what | how it reads |
|---|---|
| how hard the agent is working | **the sea**: how many swells are travelling, how tall, whitecaps above half |
| what it is doing right now | **writing in the sand**, newest brightest, older lines taken by the tide |
| something is broken | **the companion**: ears back, tail flat, amber eyes, and it stays until you come back |
| it needs you | **a solid balloon** in a warm colour, plus a bright chime |
| it finished | **a dotted balloon** in a cool colour, plus a low sonar note |
| subagents | **kittens**, one per agent, some of them swimming |
| context left | **the moon**: phase and height, with a readout that stays quiet until 65% |
| time of day | **the sky**, from your actual clock |

The rule underneath is that **the water is the work and the sky is the world**.
Sea state always means the agent. Sky, light and time always mean reality.
Nothing crosses. It is the reason a glance tells you anything at all: you never
have to ask which of two meanings a change is carrying.

The two knocks are deliberately different in shape as well as colour, so they
survive a screenshot and a colourblind reading, and they carry different sounds,
so they survive you looking at another pane. "I finished" and "I am blocked on
you" are not the same message and should never look the same.

## The protocol

Adapters translate an agent into events. The engine folds events into a scene.
Anything that can emit these can drive a scape.

```
session_start  prompt  tool_start  tool_end  error  test_pass  test_fail
sub_start  sub_end  needs_input  done  compact  context  todo  session_end
```

Send one by hand:

```sh
asciiscapes emit tool_end -tool Read -target internal/auth/handler.go -detail "142 lines"
asciiscapes emit needs_input -text "allow Bash?"
```

Events reach a running scape over a unix socket in `~/.config/asciiscapes/run/`,
and spool to a file when nothing is listening, so a scape started late still
picks up the session.

**Adapter 1: Claude Code**, via hooks. `asciiscapes install claude` writes them.
The payload schema was read out of the Claude Code binary rather than guessed;
see `notes/claude-hooks-verified.md`.

**Adapter 2** is specified and not built yet: for agents with no hooks, watch
the process instead, where alive plus output means busy and the prompt coming
back means done.

Writing one means emitting the events above. Nothing in the engine knows what
Claude Code is.

## Everything else

```sh
asciiscapes -live              # the scape in this terminal, Ctrl-C to quit
asciiscapes -info              # colour profile, size, which sound player
asciiscapes notify             # hear both knocks
asciiscapes replay session.jsonl   # feed a recorded session back through the engine
ASCIISCAPES_SILENT=1 …         # mute
```

`asciiscapes claude -print` shows you the exact commands it would run and
changes nothing.

## Notes for anyone reading the code

It is Go, standard library only, one static binary. No Node, no Python, no
framework.

Three things in here were harder than they look and are commented where they
live:

- **256 colours are not greyscale.** A dark palette collapses to grey because
  the xterm cube has almost no resolution below luma 25. The fix is to keep the
  darkness in the backgrounds and push the colour into the glyphs, which are the
  bright part of the frame. `ASCIISCAPES_CHROMA` tunes it.
- **Activity is encoded in coverage, count and position, never in rate.** A
  glance is the whole budget and a screenshot has no motion at all. The first
  version mapped activity to wave speed and idle looked identical to flat out.
- **The renderer is a real three layer alpha compositor**, so occlusion between
  the companion, the kittens and the sea is decided once instead of per sprite.

Run the tests with `go test ./...`. The interesting ones assert things that
looked fine on screen and were not: that whiskers touch fur on their own row,
that a sixty second permission nag rings once rather than once a minute, and
that a killed agent eventually settles instead of leaving the cat working
forever.

## Status

Early. The scape runs, the Claude Code adapter works, the sounds work, and the
launcher works. Stars for completed todos are specified and not built. The
companion's final coat and markings are still being chosen.

## License

MIT.
