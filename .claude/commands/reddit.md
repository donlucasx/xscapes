---
description: Reddit digest for xscapes - fresh threads ranked, comment drafts to paste by hand (never posts)
---

Produce today's Reddit digest for Lucas. READ AND DRAFT ONLY. Never post, vote,
message or log in on his behalf: comments posted by a script on a human account
are what Reddit's filter and the subs' rules ban, and a ban would follow xscapes.
The plan and the ranking are in `~/Desktop/xscapes-qt/reddit-plan-2026-09-21.md`;
the numbers a draft may cite are in `notes/reddit/findings.md` (nothing else may
be invented).

1. The Chrome extension refuses reddit.com; use the Playwright plugin. Navigate
   to https://www.reddit.com/r/ClaudeCode/new/ then, in ONE browser_evaluate,
   fetch in-page (`fetch(url, {credentials:'include'})`) the newest 25 of
   `/r/<sub>/new.json?limit=25` for: ClaudeCode, ClaudeAI, AI_Agents, aiagents,
   AgentsOfAI, and `$ARGUMENTS` if given. Save with `filename:
   .playwright-mcp/digest.json` (the only writable root), then parse it with
   python from the repo root.
2. Keep threads under 30 hours old. Rank by fit with `notes/reddit/findings.md`
   (hooks, notifications, subagents, status line, tokens, terminals, Kimi,
   Hermes, waiting, agents in parallel, "how do you know it is done") times
   freshness, and prefer threads with fewer than 15 comments. Take the top 8;
   fetch each body and its top 4 comments (`/comments/<id>.json?limit=6&depth=1`)
   the same way.
3. For each, write a comment draft of 3 to 8 sentences in Lucas's voice: plain,
   first person, specific, a number from the findings where one applies, one
   question back where natural. No product name and no link unless the thread
   asks "what did you build"; then name xscapes once and put the GitHub URL in
   the draft's last line, marked optional. No em dashes, no lists in a comment,
   no "great question".
4. Write `~/Desktop/xscapes-qt/reddit-drafts-<YYYY-MM-DD>.md`: for each thread
   the sub, flair, age, upvotes, comment count, the URL
   (`https://www.reddit.com/comments/<id>`), a two-line summary, then the draft,
   with the note at the top that he rewrites every draft in his own words before
   posting (r/ClaudeCode and r/LocalLLaMA remove AI-only text; r/aiagents allows
   AI-assisted, not AI-only).
5. Close the browser page. Report in chat: the count, the three best threads
   with why, and the file path. Do not append to the feedback log unless he
   rules on something.
