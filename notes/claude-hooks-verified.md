# Claude Code hook payloads — VERIFIED, not inferred

Source: the embedded Zod schema inside the Claude Code binary itself,
`~/.local/share/claude/versions/2.1.251` (Mach-O arm64, 188 MB), read on 2026-08-30.
Extracted at byte offset ~155,752,000. This is the definition the program validates
against, not documentation about it.

**Do not re-derive any of this. If a claim below is wrong, the binary moved — re-extract.**

```
python3 -c "
d=open('/Users/lucasgarzoli/.local/share/claude/versions/2.1.251','rb').read()
i=d.find(b'hook_event_name:N(\"PreToolUse\")')
print(d[i-4000:i+12000].decode('utf8','replace'))"
```

## Every payload carries this base object

```
session_id       string    == $CLAUDE_CODE_SESSION_ID, == transcript basename
transcript_path  string    absolute path to the session .jsonl
cwd              string
prompt_id        string?   UUID; same value until the next user prompt. Absent before first input.
permission_mode  string?
agent_id         string?   PRESENT ONLY when the hook fires INSIDE a subagent.
                           The binary's own note: "Use this field (not agent_type) to
                           distinguish subagent calls from main-thread calls."
agent_type       string?   e.g. "general-purpose", "code-reviewer"
effort           {level}?  low|medium|high|xhigh|max
```

## Per-event fields

| event | fields beyond the base |
|---|---|
| `PreToolUse` | `tool_name`, `tool_input`, `tool_use_id` |
| `PostToolUse` | `tool_name`, `tool_input`, **`tool_response`**, `tool_use_id`, **`duration_ms?`** |
| `PostToolUseFailure` | `tool_name`, `tool_input`, `tool_use_id`, **`error`**, **`is_interrupt?`**, `duration_ms?` |
| `PostToolBatch` | `tool_calls[]` — fires once per batch, after every tool in it resolved |
| `PermissionRequest` | `tool_name`, `tool_input`, `permission_suggestions?` |
| `PermissionDenied` | `tool_name`, `tool_input`, `tool_use_id`, `reason` |
| `Notification` | `message`, `title?`, **`notification_type`** |
| `UserPromptSubmit` | `prompt`, `source?` (user\|sdk\|system\|loop_wakeup\|schedule_wakeup\|poll_event), `session_title?` |
| `SessionStart` | `source` (startup\|resume\|clear\|compact\|fork), `agent_type?`, `model?`, `session_title?`, `seconds_since_last_response?`, `context_tokens?` |
| `SessionEnd` | `reason` (clear\|resume\|logout\|prompt_input_exit\|other) |
| `Stop` | `stop_hook_active`, **`last_assistant_message?`**, `background_tasks[]?`, `session_crons[]?` |
| `StopFailure` | `error`, `error_details?`, `last_assistant_message?` |
| **`SubagentStart`** | **`agent_id`**, **`agent_type`** |
| **`SubagentStop`** | `stop_hook_active`, **`agent_id`**, `agent_transcript_path`, **`agent_type`**, `last_assistant_message?`, `background_tasks[]?` |
| `PreCompact` | `trigger` (manual\|auto), `custom_instructions` |
| `PostCompact` | `trigger`, `compact_summary` |

Also present and unused by us: `UserPromptExpansion`, `PreModelSwitch`, `PostModelSwitch`,
`Setup`, `TeammateIdle`, `TaskCreated`, `TaskCompleted`, `Elicitation`, `ElicitationResult`,
`ConfigChange`, `FileChanged`, `DirectoryAdded`, `MessageDisplay`, `InstructionsLoaded`, `CwdChanged`.

## The three questions this project had been guessing at

**1. Does subagent identity arrive? YES.** `SubagentStart{agent_id, agent_type}` and
`SubagentStop{agent_id, agent_type}` are real, separate hook events. RESUME.md item 3 is
answered: the field is spelled `agent_type`, not `subagent_type`. Kittens can be counted
exactly from a start/stop id set, and each one can be *labelled* with its agent type.
Better still, the base object's `agent_id` means every tool call made *inside* a subagent
is attributable to that subagent.

**2. The 60-second nag needs no timing heuristic at all.** `notification_type` is a required
field with a closed vocabulary, and the idle nag has its own value:

```
permission_prompt  idle_prompt  auth_success  elicitation_dialog  agent_needs_input
agent_completed  elicitation_url_dialog  worker_permission_prompt  push_notification
computer_use_enter  computer_use_exit  quota_auto_resume_{fired,stale,disabled}
```

`idle_prompt` **is** the nag. Drop it by field, not by clock. No state file, no 60-second
window, no memory between hook invocations. Claude Code will even do the filtering for us:
`Notification` supports `"matcher"` on `notification_type` (the binary's `matcherMetadata`
is `{fieldToMatch:"notification_type"}`). We still re-check in Go, because a matcher that
stops matching fails *open* and would put the nag back.

Genuine "needs you": `permission_prompt`, `worker_permission_prompt`, `agent_needs_input`,
`elicitation_dialog`, `elicitation_url_dialog`. Done: `agent_completed`.

*The measured log still earned its keep* — 2,220 exact-60s Stop→Notification pairs are what
told us a filter was needed at all. See `RESUME.md`. The binary only told us how to write it.

**3. Context for the moon does NOT come from a hook.** No hook payload carries context
remaining. The statusline input does, pre-calculated and correct per model:

```
context_window: {
  total_input_tokens, total_output_tokens,
  context_window_size,          // the CURRENT model's window
  current_usage: {...} | null,
  used_percentage: number|null, // 0-100, already divided
  remaining_percentage: number|null
}
```

This matters more than it looks. Summing usage out of the transcript and dividing by a
hardcoded 200000 is *wrong on the machine we are building on*: this session measured
164,450 window tokens on `claude-opus-5`, which is 82% of 200k (the moon would scream)
but 16% of the 1M-context variant actually in use. `used_percentage` has no such failure
mode because Claude Code computes it against the real `context_window_size`.

So the moon is driven by a **statusline adapter**, and it must **chain** to the user's
existing statusline command rather than replace it — Lucas already has one at
`~/.claude/statusline-command.sh`.

## Hook mechanics

- Hook config supports `"async": true` (already used in Lucas's settings) and `"timeout"`.
- `Notification`: exit code 0 → **stdout/stderr not shown**. Other hooks differ: on
  `UserPromptSubmit` and `SessionStart`, exit 0 stdout **is fed to Claude**. So the adapter
  must print nothing, ever, on any event, or it injects text into the agent's context.
- Matcher field per event (from the binary's own switch): `tool_name` for the tool events,
  `notification_type` for Notification, `source` for SessionStart, `reason` for SessionEnd,
  `trigger` for PreCompact/PostCompact, **`agent_type` for SubagentStart/SubagentStop**.

## Environment exported to every child process

`CLAUDE_CODE_SESSION_ID` (== transcript basename, verified equal this session),
`CLAUDE_PID`, `CLAUDE_EFFORT`, `CLAUDECODE`, `CLAUDE_CODE_ENTRYPOINT`.
So a scape pane launched from inside the session can bind to the exact session with no
configuration. A pane launched separately cannot, and needs the `run/current` fallback.
