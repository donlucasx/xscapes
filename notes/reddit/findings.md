# Measured findings worth a comment (the record, for /reddit drafts)

Every number here was measured on Lucas's own logs and is written up in
`_FEEDBACK.md` or `CLAUDE.md`; the session that measured it is in brackets. A
draft may quote these as "in my logs"; nothing else may be invented. Say "one
person's use, May to September 2026" when citing them.

## Claude Code hooks (s6, s21, s27, s28)
- 12,809 hook events logged 2026-05-17 onward; 5,436 Notifications; 2,334 of
  them (43%) fire exactly 60 s after Stop with notification_type idle_prompt:
  the "waiting for your input" nudge, not a question.
- The types that mean a human is asked: permission_prompt,
  worker_permission_prompt, agent_needs_input, elicitation_dialog;
  agent_completed is a finish. Allow-list them; a deny-list fails open.
- An installer must MERGE settings.json; overwriting kills the user's own
  permissions block.
- Stop is "done". SubagentStart/SubagentStop carry agent_id and agent_type.
- His own Stop/Notification hooks stacked: 14,792 firings, 324 bursts of 2+
  inside 3 s, the worst burst 21. Debounce anything that makes noise.

## Tool calls and subagents (s21, s27, s28, s19)
- TodoWrite: 0 calls in 60,000+ tool calls through early September.
- 590 of 675 tool errors happened inside subagents, not the main thread.
- In a fan-out a subagent finished every 31 s; between his own turns 14 min.
  Sessions bimodal: 20 of 30 had 0-26 subagent runs, 10 had 64-734.
- Subagent lifetimes: the shortest seen 6.9 s and 15.7 s; the longest 57.3
  min of 639. A UI must hold a subagent on screen for a minimum time or it is
  never seen.
- Between a prompt and its finish the main thread was silent (no tool event)
  for 4.43 h of 35.7 h measured, 12.4%, the worst stretch 126 min: a "quiet
  means idle" rule needs a long timeout (30 min measured 0.00 h of false work).

## The status line (s37)
- The status payload's total_*_tokens are the current window's (the last
  response's usage), not the session's; they reset on compact. The session's
  spend is the sum of the transcript jsonl's usage records (input, cache
  creation, cache read, output); a day of work ran about 23M tokens.
- The payload carries context_window_size.

## Terminals (s26, s41)
- Claude Code emits nothing on a resize; the terminal moves rows, the app
  repaints on the next keystroke. It never uses the alternate screen and
  never addresses absolute rows inside a turn.
- Kimi Code CLI clears the whole screen AND the scrollback (ESC[2J ESC[H
  ESC[3J) on every resize step: 21 steps, 21 clears in a traced session.
- A pty read can split a UTF-8 rune; box-drawing U+2500 is three bytes and
  Claude Code writes it constantly; treat a partial trailing rune like a
  partial escape or you get U+FFFD and a shifted row.

## Other agents (s35, s40, s42)
- Kimi Code CLI has Claude Code's hook events in [[hooks]] TOML tables in
  ~/.kimi-code/config.toml; payload gaps: the prompt can be an array, the
  error an object, sub-agents are named by profile; a Stop with a sub-agent
  open is the sub-agent's; Esc is a quiet interrupt.
- Hermes Agent: a hooks: mapping in ~/.hermes/config.yaml; pre_llm_call and
  post_llm_call fire once per turn (the prompt and the done); consent is the
  user's (hermes hooks doctor).
- Cross-session messaging in Claude Code (2.1.224): a JSON file and a Unix
  socket; two sessions kept one record in one tree by anchored inserts and
  told each other what they touched.

## Waiting, at a glance (the encoding rule, 2026-08-30)
- Encode state in coverage, count or position, never in rate: a glance is the
  whole budget and a screenshot has no motion. Activity mapped to wave speed
  made idle indistinguishable from flat-out.
- One variable per perceptual channel; bind a channel to the real clock only
  if nothing else needs it (the sea is the work, the sky is the world).
