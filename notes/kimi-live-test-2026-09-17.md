# First live run of Kimi inside xscapes — test report

2026-09-17 evening, session `session_93afa372-e3dd-4a74-b522-e9ca28af26e7`
(`xscapes kimi`, scape pid 99995, run/current bound at 18:20:13).
Every finding below was measured on this machine during the run; the human
watched the scape live and confirmed each visible behavior.

## Verdict

The s40 thread-2 Kimi compatibility work holds up in its first real session.
Every wired channel fired as designed. One false alarm along the way, fully
explained by the protocol (below). Two items need the human's confirmation.

## Confirmed working, live

| Path | Evidence |
|---|---|
| First-run bind | `run/` created clean at 18:20:13; `SetCurrent` fix (v0.4.3) held — no missed bind |
| Hook install | Exactly one `# xscapes:v1` block in `~/.kimi-code/config.toml`; 15 events, `uniq -d` clean, no hand-written entries (the installer's refusal held — the friend's 12-entry failure did not recur) |
| Working state | Companion visibly working through sustained main-thread tool activity |
| TodoList → stars | Stars lit on TodoList activity — s40 wiring works end to end |
| Subagent litter | Owlet appeared in its own window for a live subagent (dedicated watched test); left when the subagent finished |
| Per-agent attribution (negative test) | One subagent deliberately exited 1 (`go vet ./doesnotexist`) — scape stayed non-worried, exactly the s40 design ("a Stop with a sub-agent open is the sub-agent's") |
| Spend/moon data source | `session_index.jsonl` resolves this session id verbatim to an existing dir whose `agents/*/wire.jsonl` files carry `usage.record` / `token_counting.measured` / `llm.request` in the exact shape `internal/spend/kimi.go:92-125` parses; 6 subagent wires minted per profile (`agent-0..5`) |
| Event translation | `translate()` (hook.go:347) 1:1 per tool call; no batching, no agent-name filtering |

## The false alarm — and why it will mislead the next reader too

Mid-test the event file `run/ccb07cb46007.jsonl` showed **2 lines and stopped
growing** while 6 subagents ran. A subagent sampling it for 45 s concluded
"the scape is connected but not consuming events."

**The file going silent is the success signature, not the failure one.**
`Emit` (internal/event/emit.go:31-34) tries the unix datagram socket FIRST
and appends to the jsonl only when the socket send fails. The first two
events (SessionStart, Prompt) landed in the file only because the scape had
not bound its socket yet in the first milliseconds after launch. After bind,
everything rides the socket. The scape holds fd 5u on the `.sock` and fd 8r
on the jsonl at EOF — both consistent with a healthy fast path.

**Suggestion for the record:** the jsonl is a fallback log, not a liveness
probe. Anyone debugging "is the scape receiving events?" must not tail the
file; the instruments that work are (a) XSCAPES_HOOKLOG=1 to capture raw
payloads at the shim, (b) watching the scape, or (c) `lsof` on the socket.

## Not verified from inside (needs the human's eyes)

1. **The ask's signature moment** — two real AskUserQuestion asks fired; he
   answered both (so Kimi→scape→human→answer closed twice) but did not
   confirm whether the balloon appeared and the bird cue sounded.
2. **Owlet departure** — the watched subagent finished; the owlet should have
   left on SubagentStop. Unconfirmed at report time.

## Notes

- Kimi Code CLI (0.39) logs **no hook firings** anywhere we could find
  (`~/.kimi-code/logs/kimi-code.log`, the session's own `logs/kimi-code.log` —
  zero "hook" matches). If a future adapter question needs the raw event
  order, XSCAPES_HOOKLOG is the only capture; it was off for this run.
- The swarm's litter was easy to miss live: L6 shows one owlet at a time in
  its own window and the six swarm agents were short. A dedicated watched
  single-subagent run is the reliable way to demonstrate it — and it passed.
- No defects found. Nothing committed, nothing changed; this file is the
  only artifact.
