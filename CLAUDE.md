# xscapes — project brief for Claude Code

*(Renamed end to end on 2026-09-03: directory, env vars, state path and hook marker. Two names are kept on purpose and are not leftovers -- `internal/envx` still reads `ASCIISCAPES_*` and warns, and `install.go` still RECOGNISES the `# asciiscapes:v1` marker so the hooks it wrote before the rename can be found and removed.)*

> **Session 27 (2026-09-09 afternoon), IN PROGRESS. Two stories landed, HEAD `745c47e`, tree clean,
> green, installed (inode 84169114). ⚠ NOT PUSHED.** He is live-testing the s26 scrollback fixes in
> this window and reports *"so far so good ... no strikethrough"*.
> ⇒ ⭐ **THE ASK BALLOON WAS POINTING AT BARE SAND, at every width, in both facings, for both
> animals.** Measured on the RENDERED frame at 124 columns: the `v` on column 101, the companion's box
> beginning at 107, its nearest eye at 112 — eleven cells of empty beach. The balloon was anchored to
> its own corner and cleared the companion by two columns; **the unmirrored layout it was mirrored from
> OVERLAPS the animal by design, and that overlap is what makes a speech balloon read as speech.**
> Now anchored to the companion's own `HeadCol()` (cat's eyes 2/6, mirrored 9/5; crab's 4/7 and the
> crab does not flip), with the pointer column READ OFF the balloon's rows so `MirrorTail` cannot
> desync it. `layout.BubbleX` is GONE — it was the wrong anchor, not a wrong number.
> ⇒ ⭐⭐ **THE `retainWidth` WIRING GAP IS CLOSED AS REFUTED — DO NOT WIRE IT.** S25 called it "the
> strongest lead yet" and said *wire it and replay*. Replayed: **31 of 975 mirrored rows merge, and 17
> of 1472 on a second independent window** — the current row with a fragment of the pre-resize row
> stitched on (`don't scroll, don't resize, don't exit.files while you`), **which is the corruption's
> own signature.** Faithful to the terminal and wrong for the mirror: the retained cells are stale, and
> a mirrored line longer than the window WRAPS in the main buffer and takes the row under it.
> ⚠ **`reallocBand` does NOT settle it, though it looks like it should**: it deletes rows
> `agentRows+1..rows`, which is the SCAPE's band — the AGENT's rows keep their tails.
> ⇒ The reason now lives at `host.go`'s `newScreen` call, plus `TestReplayTraceRetainDiff` (the
> instrument) and `TestTheMirrorDropsTheCellsTheTerminalRetains` (unit-scale). **The instrument FAILS
> on a window with no width change** — a clean result there looks exactly like a pass, the trap that
> voided the first strikethrough comparison.
> ⏭ **Still open, measured today:** crablets have ONE size (the cat's 6/10-up 4/8-down ladder is not
> ported; past ~14 sitters at 124 columns they silently stop being drawn) · traces at **13 GB**,
> `20260908-102733.bin` (10.2 GB) closed at 14:21 today and spent · **HIS:** Andale Mono in Terminal.
> ⏰ **8 days.** ⏸ Commons kit HELD at his word.

> **Session 22 (2026-09-07 evening), WRAPPED. THE CRAB IS ON THE BEACH, and the hairlines had a third
> cause nobody had found.** ⚠ **NOTHING IS COMMITTED.** Five stories sit in the working tree, all green
> (`go test ./...`, `go vet`) and installed at `~/.local/bin/xscapes`. Land them as five commits, not
> one. `internal/companion` is OURS now — the parallel session released it.
>
> **HERO SHIPPED AND IS THE DEFAULT.** His pick from four silhouettes over two rounds; salmon LOCKED at
> **cube index 210** (an exact cube entry, so it survives the glyph path's 2.6x saturate unchanged).
> Five states, crablets on the sand, crablets in the water (**a crab does not swim — it walks in and the
> eyestalks carry on above the surface**, same 10x8 box as `KittenSwim` so the lane arithmetic needed no
> new numbers), and an exit queue that goes UNDER rather than swimming off. `xscapes companion
> [cat|crab]` switches it, `XSCAPES_COMPANION=` overrides for one run, `/companion` shells out to it.
> ⇒ **Join cost nothing**: the sprite data became a FIELD on the existing type, so `NewCrab()` returns
> what `NewCat()` returns and **all 25 call sites are untouched**. No interface anywhere.
> ⇒ Measured, so nobody re-derives it: the crab's ink is **12x7 cells to the cat's 9x7**, at 64-66 of 84
> cells against the cat's 60 — same weight, arranged sideways. **The three extra columns were always
> free**: `cat.Size()` returns the BOX, not the ink, and the cat has never used its last three.
>
> **Session 26 (2026-09-09), WRAPPED. ⭐⭐ BOTH SCROLLBACK DEFECTS ARE CLOSED, AND BOTH WERE OURS.**
> HEAD `da4ea99`, tree clean, suite + vet green, installed. Page still LIVE (HTTP 200).
> ⭐⭐ **THE STRIKETHROUGH WAS THE FIRST LINE OF `screen.feed`:** `r := []rune(s.pending + in)`.
> `pending` held a trailing partial ESCAPE and never a trailing partial RUNE. The agent's output comes
> off a pty read that splits wherever it likes, and Claude Code's rule is **U+2500, three bytes at a
> time** -- split one and `[]rune` makes U+FFFD, **one cell becomes THREE**, every cell after it shifts
> two columns and the overflow wraps onto the row below. `"──── done ────"` fed with a cut after one
> byte becomes `"���─── done ────"`.
> ⇒ **Why seven sessions missed it: the TERMINAL never saw any of it.** It gets the exact bytes and
> draws them right, which is why his no-xscapes session is clean and the LIVE screen is clean. Only the
> MODEL was wrong, and `Host.mirror` writes the model's rows into SCROLLBACK -- the one place he ever
> saw it. Every screenshot in seven sessions was scrolled-back content and nobody asked why.
> ⚠ **When a defect only ever appears in one view, suspect the thing that renders that view.**
> ⇒ **HIS TEST closed it**: bare Claude at the same window height stayed clean, which eliminated the
> agent AND the height in one move and left only something we do to bytes nobody else touches.
> ⚠ **FOUR HYPOTHESES DIED FIRST — do not raise them again.** `Rebind`'s `ESC[NS` (a bare SU keeps
> ZERO rows) · an off-by-one between region and pty (34↔34, 21↔21, 29↔29, 33↔33) · a shared cursor-save
> slot (**1,054,053 of 1,054,059** `ESC7`s are ours; the agent never uses DECSC) · our band paint
> disturbing the agent (**966 paints** replayed, every one left its rows and cursor byte-identical).
> **What found it was feeding the same bytes in different CHUNK SIZES and getting different screens.**
> ⇒ **THE DUPLICATION IS ALSO CLOSED** (`731768c`): `resizeScrolling` fed the mirror on every shrink and
> the agent repaints after a resize, so the rows came back and were kept again. **A scroll means
> content moved on; a resize means only the viewport changed.** Reverses
> `TestAShrinkKeepsTheRowsTheTerminalDestroys` deliberately; the guarantee survives because content
> that really leaves is SCROLLED and `scrollUp` still keeps it. ⚠ Residue: an agent that does NOT
> repaint after a resize loses those rows from scrollback.
> ⇒ **A latent PANIC guarded** (`719572e`): `ESC[?47l` swapped buffers with no nil check, so a model
> seeing a LEAVE before an ENTER nils `cells` and the next newline panics in `scrollUp`.
> ⇒ ⭐ **HERO PACES** (`b92a9cf`), built from his own study. One step per MAIN-THREAD tool event -- a
> COUNT and a POSITION, both of which survive a screenshot. Subagent events do NOT move him (one event,
> two channels, is what the encoding rule forbids). Inward on a triangle: **107 → 101 → 107** at 124
> columns, sand pulled back to 100. needs-you/done/worried hold still.
> ⚠ **I said his "pace the other way" note was nowhere; it is in `notes/charstudy/crab/pace.go`
> verbatim.** Wrong about the record twice in two days -- SEARCH `notes/` before saying something is
> unrecorded.
> ⏸ **HIS, and both are held at his word:** the Commons kit regeneration (*"wait on the commons kit"*)
> and *"lets focus on open engineering before we do anything hackathon related"*. ⏰ **8 days.**
> ⇒ **PUSHED 2026-09-09 at his word** ("lets switch the login to donlucasx"): `origin/main` is
> `2f4ce38`, verified with `git ls-remote`, nothing unpushed.
> ⚠ **`gh auth`'s active account is HOST-GLOBAL and it does NOT STAY PUT.** It flipped back to
> `eclipsevalidators` between two of this session's own commands, because his Validators session
> reclaims it. ⇒ **Switch immediately before every push, do not assume a previous switch held:**
> `gh auth switch --user donlucasx && git push origin main`. The failure is a misleading 403 —
> *"Permission to donlucasx/xscapes.git denied to eclipsevalidators"* — which reads like a repo
> permission problem and is an identity problem. Three GitHub accounts here, same rule as the three
> Vercel ones.
> ⇒ **HE RESTARTED 14:21:52 AND IS RUNNING ALL OF IT** — scape PID 87332, resumed BY ID, running
> inode **84141583 == installed**. First live look: *"i dont see anything strikedthrough so far."*
> ⚠ **Thin, and say so.** 559 mirrored rows in the new trace, 7 with U+FFFD and **all 7 his own quoted
> examples**; the old trace gave 174/0, 63/**2**, 67/0 across three windows. Consistent with the fix,
> NOT proof — the unit test is the strong evidence. ⚠ My first version of that comparison was VOID:
> three 40 MB windows containing ZERO complete mirror blocks, reporting clean. **A clean result from a
> window with nothing in it looks exactly like a pass.**
> ⏭ **He is starting a FRESH session to keep testing.** Work in it, resize, scroll back, and look.
> ⚠ Traces at **13 GB**, 102 GiB free. `20260908-102733.bin` (9.5 GB) is CLOSED and spent — safe to
> delete. **KEEP `20260908-005450.bin`**: it holds the 07:11 occurrence, the only full recording of
> one, and it is what any future re-check compares against.

> **Session 25 (2026-09-08 morning), WRAPPED. THE KIMI AUDIT KILLED MY SCROLLBACK MECHANISM.**
> HEAD `957f9f0`, pushed, tree clean but for `notes/kimi-audit-brief.md`. Page still LIVE (HTTP 200).
> ⭐ **THE MECHANISM I REPORTED IN S24 IS REFUTED.** I claimed the agent's post-resize repaint writes
> column-addressed segments with NO per-row erase, so cells between them keep stale content. Counted:
> **exactly 33 `ESC[2K ESC[1B` pairs, contiguous, at offsets 9843..10106, right after `ESC[H` — and 33
> is the AGENT ROW COUNT.** It blanks every one of its rows before drawing a glyph. **Why I was wrong:
> I read a 3 KB window starting at offset 10050, which is INSIDE the erase run.** A window is not the
> structure. ⚠ And "300 KB of the occurrence" is **~2.6 SECONDS** of trace at 6.6 MB/min — we aimed at
> the resize; the corrupted rows are written when they are MIRRORED, which is a different moment.
> ⇒ ⭐⭐ **`retainWidth` IS A WIRING GAP AND IT IS IN THE MIRROR PATH — the strongest lead yet.**
> `Rules.RetainsWidth` is measured, documented and honoured by `reallocBand`; `screen.retainWidth`
> documents the same rule; **`host.go:272` calls `newScreen(cols, rows)` and never connects them.** It
> is true only in tests. So the model that decides what gets mirrored into SCROLLBACK runs with
> Terminal.app's measured width rule OFF, while `reallocBand` runs with it ON — and this resize was a
> WIDTH change. ⚠ **NOT claimed as the cause.** Wire it and replay.
> ⇒ **`ESC[8S` exactly undoes Terminal.app's grow-push, net zero** — s24's open question is closed:
> our scroll-up does NOT move rows under the repaint.
> ⚠ **Kimi was wrong twice** (both checked): Andale Mono IS installed (`/System/Library/Fonts/
> Supplemental/`); and "the corrupted strings are absent from the .bin" refutes nothing, because the
> CLEAN text is absent too — it is written as column-addressed segments.
> ⇒ **FONT ADVICE CORRECTED, I overstated it first**: 87.6% vs 100.0% holds across hhea and usWin, but
> Menlo's metrics predict 3.7px unfilled where his screen measures 5.2px, so Terminal.app adds its own
> leading. **Andale Mono should SHRINK the gap ~3.5x, not close it.**
> ⇒ **`xscapes companion <name>` now reaches a RUNNING scape** (`957f9f0`): `companionPref()` was read
> once in `newFrames`; `frame()` re-reads every 500 ms. ⚠ My first test asserted ccw/MoonX would move
> on a swap and **failed on the truth** — `Size()` returns the BOX and both animals are 12x7.
> ⚠⚠ **RETRACTED THE SAME MORNING, AND IT NEARLY COST HIM LIVE WORK.** This banner said "three
> restarts failed, his restarts made new empty sessions" and told him to close two windows. **Checked
> by cwd: ttys002 is `~/Documents/claude/tyastie` (mid-run, four-agent audit) and ttys005 is
> `~/Documents/claude/Validators`.** They are `xscapes claude` because he runs the scape in every
> project — the product working, not orphans. I saw two processes with the right name and never ran
> `lsof -a -d cwd`. **Nothing here supports a failed-restart story: PID 10648 has run unbroken since
> 00:54:50, so this window was never restarted at all.** The real defect was the one that got fixed:
> the old binary reads the companion once at startup. Still resume by ID as cheap insurance —
> `XSCAPES_TRACE=1 xscapes claude -- --resume 2c247007-955d-479f-9266-7d5d4f8d6db1`.
> ⚠ Twice I said "restart" and then installed a new binary minutes later. **Install FIRST, then ask.**
> ⏰ **9 days. The Commons paste kit at `~/Desktop/xscapes-commons/` is STALE — all three files still
> say cat.** Regenerate before he submits.

> **Session 24 (2026-09-08, early hours). ⭐ THE ENTRY IS ON THE CRAB AND IT IS LIVE.**
> https://donlucasx.github.io/xscapes/ rebuilt and republished, verified with real fetches: the page,
> **all fifteen clips**, the deck (HTML + PDF) and the README. HEAD `7f68db6`, pushed, tree clean, green.
> ⇒ The clips render **`companion.New(companion.DefaultName)`**, not `NewCat()` and not
> `companionPref()` — the SHIPPED default, so the page renders the same on any checkout and shows what
> a fresh install gets. Confirmed in pixels, not assumed: **(253,134,134) at 3954 px** in `g-worried`
> (cube 210, two counts off from the GIF quantiser) and **one contiguous salmon run twelve cells wide**,
> the crab's ink exactly. `g-sea`, `g-sand`, `g-stars` are byte-identical, which is what proves the
> change is only the companion.
> ⇒ **The copy describes the animal on screen.** A crab has no ears and no tail: "ears back, tail flat"
> is now "claws down, stalks short", and the worried pose reads *the widest thing on the beach becomes
> the smallest*. Kittens are **crablets**, and they do not swim off — a crab walks in and the eyestalks
> carry on above the surface. "Choose your xscape" now says the companion IS a choice today, with the
> command, instead of promising one.
> ⚠ **Three stale facts a reader could have caught first**: the README said the readout "stays quiet
> until 65%" when `ReadoutFrom` has been **0.40** since 09-05 · `xscapes companion` was undocumented ·
> the deck's title slide still said "Claude Code runs inside a shoreline" (part of 0B, fixed because
> the cat was in the same sentence). Also 12 declared clip sizes were wrong BEFORE tonight (the four
> state clips are 675 px and were declared 672).
> ⇒ **`go test` no longer writes into his real `~/.config`.** Run from inside a traced session, every
> test that built a Host inherited `XSCAPES_TRACE=1` and opened a REAL trace — sixteen files in one
> run — while three replay tests failed on `open 1`. `TestMain` clears it for the package. Same trap as
> `term.NoSplitCells`, different variable.
> ⏳ **A trace is RUNNING** (`~/.config/xscapes/traces/20260908-005450.bin`, scape up 00:54:50, window
> 119x51). ~6.6 MB/min. **When the strikethrough appears: STOP** — no scroll, no resize, no exit.
> ⏸ **HIS, open:** test **Andale Mono** in Terminal (its block glyph fills **100.0%** of the line box
> against Menlo's **87.6%**, read from the font files) — if it holds live it kills every remaining
> hairline AND gives back the split-cell gradient, better than the trade he took · Commons publish +
> submit, **closes 09-17**.

> **Session 23 (2026-09-07 night), WRAPPED. SESSIONS 22 AND 23 ARE LANDED — eight commits, HEAD
> `62f2664`, tree clean, all green, installed.** ~~UNPUSHED~~ **PUSHED 2026-09-08 with his go-ahead.**
> ⚠ **ONE QUESTION IS OPEN AND IT IS HIS**: the scrollback repro, or rebuilding the entry around the
> crab? **Ten days to Commons and the live page, its five clips and the deck all still show the CAT.**
> The corruption damages the agent's SCROLLBACK, not the scape — a judge never sees it. Recommendation
> on the table: **the entry first**, trace running in the background.
> ⭐ **THE SCROLLBACK BECAME TRACTABLE TONIGHT, on two measurements.** (i) The corrupted window's scape
> started **20:28:20** against a **21:04:44** screenshot — **36 minutes**, so "it needs a long-lived
> window" is REFUTED (it was my own claim, an hour old). (ii) **The RESIZE is on the record for the
> first time**: the same window is titled **107x51** at 21:04 and **119x51** at 21:18, with an
> occurrence after. ⇒ `XSCAPES_TRACE=1 xscapes claude -- --continue`, resize, work normally, and it
> fires inside the hour. **When it fires: STOP** — read the window back and pair it with the bytes.
> ⚠ Still not proof of CAUSE: one window, one grow.
> ⚠ **The fix does not reach a running scape.** Settled by inode, not argument: PID 66211 holds
> `83591812`, `~/.local/bin/xscapes` is `83606834`. The new-inode install is what keeps a running scape
> alive and it is also what keeps the fix out of it. **Restart, or you re-photograph the old picture.**
>
> **THE DISC'S OUTLINE IS CLOSED, and the fix session 22 proposed was wrong.** Per-cell half-block ORIENTATION cannot work: Menlo's block ink runs **4.6 →
> 29.4 of a 30px row**, so a cell has two achievable partitions and BOTH leave the background at the
> bottom — `▄` errs 0.6px there, `▀` errs 4.6px at the top. Making the sliver match what is below
> means `▀` everywhere, which is U+2580, which is the glyph `LowerHalf` exists to avoid. **No glyph
> avoids the leak; it is the terminal.**
> ⇒ **What is wrong is the RUN, not the pixel.** `notes/rulecount` counts rules and their widths at
> Terminal.app's geometry, calibrated against his own crop (17px sky · 12px rim · ONE px of
> 0.43·rim+0.57·sky · body). At his 128x27 the frame carries **4 rules, all on the disc**: two **five
> cells wide** at the flat top and bottom CAPS, two one cell wide at the shoulders. A split cell whose
> neighbour's edge is in the same cell on the same side buys **no roundness** — it only draws 70px of
> line. ⇒ **`Shore.FlatCaps`** (default on, gated on `NoSplitCells`) gives those cells a whole cell and
> leaves the shoulders split. **Longest run 5 → 1 cell, rule cells 12920 → 4922**, silhouette width
> profile IDENTICAL. ⚠ `TestTheMoonIsRoundAtEveryHeight` had been running at GHOSTTY's flags; it now
> runs both, and `TestNoRuleRunsAcrossTheDiscsCap` was checked to FAIL without the fix.
> ⇒ **Verified against his 9.04.45 PM moon crop**, which is the OLD binary (his scape started 20:28,
> the fix installed 20:42). Every rule in that crop sits at **+29px of a cell row** — the model
> confirmed at four independent places — and the runs measure **one 3.0-cell and four ~1-cell**. The
> instrument at his 107 cols x 22 rows says **6 rules with two 3-wide → 4 rules, all 1-wide**. Same
> picture. ⚠ **What remains is four 14px dashes at the shoulders, and that is not "pixel perfect".**
> ⇒ **ZERO rules is reachable and now MEASURED, and it is HIS ruling**: collapse EVERY edge cell by
> area coverage and the whole sweep goes to **0 rules** — but the disc is a **rectangle at 3 of 13
> heights (18, 22, 23 rows)**, and ~22 rows is about what his window gives. That is the "Pixel perfect,
> no exceptions" option from the 09-06 menu, no longer an estimate. The shipped flat-cap fix is round
> at **13 of 13**.
> ⭐ **THE "IT NEEDS A LONG-LIVED WINDOW" THEORY IS REFUTED, and it was mine, written an hour earlier
> in this same file.** His 9.04.44 PM screenshot shows the corruption in a window whose scape started
> at **20:28:20** — **36 minutes old**. The s22 reading (22-hour window: 210 rows · fresh window: none)
> was true but the wrong variable: age is not what earns it. ⇒ **The trace is NOT blocked.** Launch
> traced and work normally; it showed up in half an hour. ⚠ Still UNMEASURED, and it is the one thing
> that separates the hypotheses: whether that window had been RESIZED. Ask, do not assume — this defect
> has produced three confident wrong answers already.

> **⭐ THE HAIRLINES: a THIRD path, and it is why two previous fixes survived.** `canvas.resolve` has a
> branch that takes a cell with **no half state and no ramp** — a plain single-colour background — and
> **splits it anyway** to place a band edge implied by its neighbours. It was never gated by
> `NoSplitCells`. The ramp path got gated, the disc was made to paint one tone, and then this put the
> split back FROM THE OUTSIDE. Now gated. The disc's INTERIOR also collapses to one tone where both
> halves are inside it; the silhouette still splits, so roundness is untouched. **Interior rules on the
> disc: 156 → 0** over 7 heights x 48 half-hours. ~~⚠ **STILL OPEN: the disc's OUTLINE**~~ — **CLOSED
> in session 23, but NOT by the fix proposed here, which was wrong. See the session 23 block above.**
>
> **The right-edge strip is FIXED.** `297dd7a` was the measurement, not the fix; there was no DL
> anywhere in `internal/host`. `reallocBand` now runs on a WIDTH change only.
>
> **⭐ THE SCROLLBACK FORK IS SETTLED: BAD BYTES.** AppleScript returns plain text with no attributes,
> and the captured row holds real **U+2500** characters — so it is not a strikethrough SGR and not a
> Terminal.app rendering artifact. The **210-row sample** is at
> `~/Documents/Screenshots/corrupt-hist-515.txt`. ⚠ It reproduces on LONG-LIVED, REPEATEDLY RESIZED
> windows: the 22-hour window had 210 corrupted rows, a fresh one had none. **Trace that window next.**
> Tracing is now `XSCAPES_TRACE=1` (picks its own path, makes the directory, and **never fails
> silently** — it did, and it cost a whole occurrence). ⚠ A trace grows **~7 MB/min**.
>
> ⚠ **THE ENTRY SHOWS THE WRONG COMPANION.** He made the crab default knowing the live page, its five
> clips and the deck all show the CAT. **Commons closes 09-17.**
>
> ⚠ **Two lessons paid for tonight**: the `!` prefix is a CLAUDE-PROMPT convention and is wrong at a
> shell prompt — every `!` command given to him was mis-instructed · and **`ps eww` cannot read another
> process's environment on this machine**; a positive control read empty too, so a claim built on it
> had no evidence. He pushed back and was right.

> **Session 21 (2026-09-07), WRAPPED. The entry stopped being the thing that could kill this.**
> ⚠ **`internal/companion` BELONGS TO THE PARALLEL SESSION** — he is implementing the crab ("Hero")
> there and will open a fresh session here when it lands. Do not touch it. Brief and pick in
> `_FEEDBACK.md` §"Session 21 (parallel)". The unmade question that sizes it is his: does Hero
> **replace** the cat or **join** it as a choice? No companion interface exists; `NewCat()` is
> concrete at 25 production sites + 13 in tests.
>
> **Shipped:** **v0.3.0 tagged** — `go install …@latest` was serving **v0.2.1, 115 commits behind**,
> and that is the command on the README and the live page · **the worry bar on his ruling
> "main-thread errors only"** (`reduce.go` raises `Worried` only when `e.Agent == ""`): over his real
> 330h, **worried 44.6% → 16.0%, working 49.6% → 78.1%**, because 590 of 675 errors fire inside a
> subagent — the error still drives the sea and still lands on the sand, only the FACE is quiet ·
> **`xscapes -h` serves help** (it printed a flag error; `dispatch`'s own `case "-h"` could never run,
> it returns early on any `-` argument) · **the live page and clips rebuilt and republished**, having
> served the star-on-the-disc bug he reported for a day and a half · the Commons brief rescoped and
> the demo-video promise removed.
>
> **His rulings, closed:** the worried eye **stays amber** (red is darker than the ground — WCAG 5.03
> → 1.24 at night) · the disc keeps its **flat caps** · the **shoreline stays split** · **no demo
> video**, taken knowingly against the ~30% of the rubric that rides on the waiting experience.
>
> ⚠ **The 37% worry figure below is stale — it is 44.6% before the fix, 16.0% after.**
> ⏰ **Commons closes 2026-09-17.** The paste kit is current at `~/Desktop/xscapes-commons/`; the
> **publish** step is the blocker, not the submit click.

> **Session 19 (2026-09-06, ~17:50–18:45), his three reports — the machine CRASHED mid-session; recovered
> 19:00.** (Not the 09-05 "Session 19" further down; that was a parallel session.) He ran long live
> sessions and reported: (a) *"the Sun seems to break sometimes, and fix itself"*, (b) unintentional
> thin lines on 256-colour Terminal.app — *"It should be pixel perfect"*, (c) scrollback text
> *"corrupted, striked through"*. All three are measured and logged in `_FEEDBACK.md` §Session 19.
>
> **His two rulings, verbatim.** The sun: **"Sun wanes as a crescent"** — SHIPPED and installed in
> `8d84eeb`; no unlit face is painted by day, the moon keeps its shaded face at night, and
> `SunShadow = "slate"` keeps the old look for the study pages. The thin lines: **"Leave it, I'll try
> line spacing first"** — *ship nothing here yet*; **he is testing Terminal.app's line-spacing setting
> and reporting back**. The block glyph is ~25.4px in a 30px cell, so the gap is leading and a terminal
> setting may fix it for free: **Terminal → Settings → Profiles → Text → Font "Change…" → Line Spacing.**
> If that fails he picks from a pre-costed menu of three (in `_FEEDBACK.md`); do not re-derive it.
>
> ⚠ **The 09-05 "U+2584 is bottom-exact" measurement is REFUTED.** Its ink runs 17 → **29.4** of a 30px
> row, half a logical pixel short. The last device pixel row falls back to the cell's background, so a
> 1px rule reads across every two-colour cell — 5 in his sky, 4 inside the disc. ▄ reduces the hairline
> from 5px to 1px; **it does not remove it**, and within half blocks on Terminal.app it cannot be.
> `notes/lineprobe` renders a frame at that geometry so it is visible in a browser.
>
> **(c) the scrollback is OPEN and is the next work.** It is NOT a strikethrough attribute — the marks
> sit in their own cells and do not cross the letterforms. It is a cell-level merge. The fork nobody
> has settled: **bad bytes or bad drawing.** ⚠ `/tmp/apple.bin` and eight other traces were DESTROYED
> by the 09-06 18:51:32 reboot; capture any new trace outside `/tmp`.

> **Session 18 (2026-09-05 into 09-06), WRAPPED. The submission page is LIVE and is the entry.**
> https://donlucasx.github.io/xscapes/ — `site/publish.sh` force-pushes `site/index.html` + `anim/*` as an
> orphan `gh-pages` branch; **rerun it after every page rebuild.**
>
> **Shipped to the product** (installed, `fd6a43c`), from his two rulings on the features test: **a kitten
> stays at least 60 s after its subagent starts** (`reduce.KittenDwell`; an end inside the dwell is
> remembered and the swim-off waits), because a subagent that finished in seconds was invisible and he had
> reported seeing none twice; and an **exit queue** (`exitSpans`), because the dwell makes a fan-out leave at
> one instant and the painter drew them all on one another. **The companion is 12 columns by 7 rows** in the
> table and the bullet below; the "3–4 lines" figure was stale.
>
> **Shipped to the page**, in the order he asked: the brand guidelines say ink, not gold · the copy
> **RESCOPED** — xscapes is for **any agent that runs in a terminal**, not Claude Code alone (`xscapes inside
> <cmd>` has always done this and was never on the page), a ***scape* is any of the ASCII landscapes and the
> shoreline is the FIRST one**, not the product, and "four states" is gone because the real feature is that
> the scape keeps your own time of day · the logline he picked from four, *"Cozy ASCII scenes that react to
> your agent while it works"*, with the negative opener dropped at his word · the legend promoted, reordered
> most-useful-first and retitled **"How to read it"**, each row with a **dissected clip** of the one part of
> the scene that carries it · **"Choose your xscape"**: the rainy window at dusk with the frog, the aquarium
> at night with the owl · ASCII rules on every section header · everything **truecolor** · and **the hero is
> one whole session that loops**, 160 frames over 16 seconds — the sea rising and settling, kittens arriving
> and swimming off, stars lighting, the moon sinking with the context, the companion asking and finishing,
> then a **compact and a fresh list, which is what closes the loop**, the day turning underneath, the agent's
> own transcript scrolling above.
>
> ⚠ Three traps this session set: the stored GitHub token has **no `workflow` scope**, so Pages is a branch
> and not an Action · `.gitignore` excludes `site/anim/*.png` with a **name-specific** exception, so a renamed
> still silently stops being committed · **the moon and the constellation are washed out at midday by
> design**, so any clip of them must run at night.
>
> **Commons needs only him now**: the rules were re-read signed in and the entry picker takes only a
> published Commons build, so the entry is this page served verbatim. Brief in `site/commons-brief.md`, paste
> kit on his Desktop, **Default · Quick, not Expert**. Closes **09-17**. Details in `RESUME.md`.
>
> **Session 19 (2026-09-05), WRAPPED. No code changed; three things wait on him.** A live features
> test run INSIDE `xscapes claude` at his ask, plus two research streams. **The scrollback corruption
> REPRODUCED a third time and was UNTRACED again** (his screenshot of this session's own scrollback:
> table rules struck through prose rows, `sc-ne` inside a doubled header = the s14 #2 merge signature).
> Confirmed ours rather than inferred — host PID 29679, `TERM_PROGRAM=Apple_Terminal`, so `-history`
> was ON and those rows were written by our model through DECSET 47. NOT claimed as ours: the
> duplicated prose blocks (s14 #6 — Claude Code re-renders; a plain terminal keeps both). **Order for
> the next one, and it is not optional: read the live window back FIRST** (`1049l` discards most
> mirrored rows, s14 probe 3, so a restart destroys the evidence), then
> `XSCAPES_TRACE=/tmp/apple2.bin xscapes claude -- --continue` — **the `--` is REQUIRED**, verified
> with `-print`. Two measurements that are HIS ruling: **kittens are the SECOND report** (*"i did not
> see any multi agents at work"*; not a channel bug — his subagents lived 6.9 s and 15.7 s against a
> 6 s exit, so nothing was on screen by the time he looked — a minimum dwell, or short subagents stay
> invisible), and **the companion figure below is stale** (the table says 3–4 lines; `CatBody` is
> 24x28 px at `W/2, H/4`, so the shipped cat is **12 cols x 7 rows** — the line stays until he rules).
> Research, all his pick, nothing chosen: five scenes (hearth/cabin · rainy window · café table ·
> aquarium · snow cabin; rain is a MODIFIER, not a scape, and the planned Campsite is the weaker half
> of its archetype) and five companions (owl · rabbit · frog · hedgehog · otter; bird/fox/wisp turn
> out to be words in this file only, never drawn). Details in `RESUME.md` ▶ NEXT 1–4, `_FEEDBACK.md` s19.
>
> **Session 16 (2026-09-05), WRAPPED.** Live tests PASSED in both terminals. Shipped + installed:
> kitten swim-off · the context READOUT from 40% used (his ruling; it had never been in the live
> scene) · ▄ split cells on Terminal.app (⚠ **re-measured 09-06: the hairline went 5px → 1px, it is
> NOT gone** — see the session 19 block above) · the disc-tip sky half through the ramp ·
> **the disc's edge is the HUE RIM** (his pick from "The Moon, Four Ways"; quad edge, shadowless sun,
> night halo stay as study switches) · the outstanding-todo ring REMOVED (his ruling) · swimmers
> never share columns. Measured: Terminal.app's alt screen RETAINS rows at their widest on a width
> change (`notes/width-audit.md`). **The site and the deck are rebuilt on the brand** with five
> TRUECOLOR clips from the real reducer (`go run . -site site && python3 site/make-gifs.py`, then
> `assets/deck/make-deck.py`), a cover over the real sea's glyphs and a one-page pitch; his look is
> next, then Commons (closes 09-17). **Brand LOCKED** by the parallel s17 session (b&w on the page and
> deck by his ruling; `assets/brand/`). Details in `RESUME.md`.
>
> **Session 15 (2026-09-04).** His picks: gradients **cube-path ONLY**; Terminal
> automation **this session only** (tty rule); the page stays unpublished until the
> pending items are in. **Built, measured, and INSTALLED** (*"Install it"*): the sky and the sea are
> painted as one PATH through the 256 palette (`term.Ramp`, `canvas.SetBGRamp`) instead
> of rounding each row — hard edges sky 84→65, sea 120→99, half-hours with a step ≥30
> 29→6, the largest step unchanged at 33 (the cube's green step; unavoidable). Page:
> artifact "Sky and Sea Repainted". Pushed; `~/.local/bin/xscapes` rebuilt. His first live
> look (124x52): the moon was a RECTANGLE — pre-existing, the disc's radius is 1.92 rows there
> and lost its tips — **FIXED** with half-row sampling (`canvas.SetBGHalves`); the companion's
> right margin now grows with the width (5 columns at 124, was 2) — **FIXED**. NEXT: live in it.
> **Afternoon, his first run in Ghostty**: the sun differed (truecolor painted the raw tan blend)
> and a resize broke the layout (Ghostty's alternate screen anchors content to the TOP on a grow
> and moves the cursor with its row on a shrink; the host's tick encoded Terminal.app only). Both
> SHIPPED and installed on *"standardize the experience"*: **the cube on every terminal**
> (`DetectProfile`), and **`host.Rules` per TERM_PROGRAM** (`RulesFor`; Terminal.app byte-identical,
> everything else xterm-like). Committed `b9d65e7`, pushed. His last note: the companion's EYES are
> holes to the sea by day (by design, eyes sit in gaps of the body bitmap) — his pick. Details in `RESUME.md`.
>
> **Session 14, the evening.** He lived in it and came back with six reports
> (`_FEEDBACK.md`): three FIXED (swimmers above the waterline, balloon text
> capped, the model's ESC ( B); the scroll-back glitch NOT reproduced in the
> buffer; the transcript-start garble PARTLY diagnosed (Ink re-rendering while
> his 747 permission warnings scroll the band; one model divergence fixed, the
> proper diff still to do); and **the gradient review MEASURED, nothing changed,
> his decision pending** (`notes/gradientaudit`, artifact "Sky and Sea by the
> Hour"). ⚠ An automation incident typed `/exit` into HIS window: no Terminal
> driving without his OK, and never `front window` (hub memory).
>
> **Where we left off — 2026-09-03 (session 14).** The submission page exists
> (`site/`, one static file, `xscapes -site site`) and is NOT yet published or
> submitted — his hands. The scrollback plan below was AUDITED at his ask and
> REPLACED: the alternate screen has no history (measured), but DECSET 47
> switches buffers without clearing, so the terminal's OWN scrollback can be fed
> while the band stays up; and every shrink was misplacing Claude's input box
> (the terminal leaves the cursor, the host restored it into a band that no
> longer held its row) — Report 1, reproduced and FIXED. His ruling: *"Go:
> shrink fix, then mirroring."* Plan and measurements in
> `notes/scrollback-audit.md`. **Both SHIPPED the same evening**: `RebindShrinkAlt`
> and `Host.History` (rows leaving the band mirrored into the main buffer
> through DECSET 47, replayed on exit; `-history`, default on in Terminal.app).
> The screen model is production now (`internal/host/screen.go`). NEXT: live in
> it for a day.
>
> **Where we left off — 2026-09-03 (session 13), HEAD `9452bc6`, pushed, tree clean.**
> **Session 13: the resize damage was OURS, twice, and he had to report it four
> times.** Both fixed and mutation-proven — the clear was painting rows in
> Claude's own background (an erase fills with the CURRENT background), and the
> host never undid the terminal's downward push on a grow.
>
> **The measured fact that settles all of it:** Terminal.app's ALTERNATE screen
> anchors CONTENT to the BOTTOM edge both ways, and the CURSOR moves with
> NEITHER. **The rule worth more than the fixes:** an instrument can answer a
> question you did not ask — four did today, each producing a confident wrong
> answer to him. An external audit (agent Kimi, at his suggestion) caught the
> worst one before any measurement did.
>
> ⚠ ~~**SCROLLBACK IS A REQUIREMENT WITH NO CHEAP ANSWER.**~~ **Superseded
> 2026-09-03 (s14): there IS a cheap answer, measured — mirror into the main
> buffer through DECSET 47 (see above).** The rest of this paragraph was true
> as far as it went: the main screen reflows on width, tmux stacking leaves a
> seam, and the alternate screen has no history of its own.
>
> **Session 12 was the rename.**
> Finished end to end: directory
> `~/Documents/claude/xscapes/`, `XSCAPES_*` env vars, `~/.config/xscapes/`, and
> the `# xscapes:v1` marker on the twelve installed hooks. `ASCIISCAPES_*` still
> works and warns (`internal/envx`); `install.go` still RECOGNISES the old marker
> so pre-rename hooks can be found and removed. ⚠ Two running scapes went deaf in
> the move and need `xscapes claude` again. Nothing else about the project changed
> — session 11's findings below still stand, as do the three things waiting on him.
>
> **Session 11:**
> **Live: https://github.com/donlucasx/xscapes** (public, MIT). Milestone 1 is
> COMPLETE, the hooks are installed and firing, and the agent runs INSIDE the
> scape on the alternate screen (`xscapes claude`; `-beside` is the old tmux
> layout, `inside <cmd>` hosts anything, `-alt=false` reverses the alt screen).
>
> **⭐ SESSION 11 FOUND THAT NOTHING MARKED DONE ACTUALLY WAS.** Seven defects in
> shipped features, none visible in any study, all fixed: the sea and sky had NO
> COLOUR on Terminal.app for most of a working day · the 256 sky was the wrong
> HUE, not just banded · an electric night that shipped AND was pushed · a
> grey-fringed sun · a blob moon · a resize leaving sky in the agent's transcript
> · kittens losing an eye to their neighbour's seam. Two channels were measured
> for the first time and both were broken: the sea's dynamic range was collapsed
> into a fifth of itself (the floors CLAMPED instead of lifting), and the
> companion's alarm is on **37% of active time**.
>
> **The durable part is the instruments**, and the rule they teach: build the
> instrument before trusting the picture, and measure the RENDERED frame rather
> than what the source says it should do.
> `xscapes tune` folds real recordings through the reducer offline (`-sweep` for
> settings) · `internal/host/screen_test.go` is a small terminal that replays the
> host's real bytes through eleven resize modes · `xscapes shades` asks the
> smoothing question in HIS terminal at HIS window size · `-day` shows five
> panels an hour including the glyph cells 256 changed.
>
> **⭐ TARGET IS TERMINAL.APP, his ruling 2026-09-02.** Cube-exact colour is the
> general rule; `term.Index256Keeping` preserves hue for backgrounds and leaves
> GREYS ALONE, which is the guard that stops it forcing an electric night.
>
> ⚠ **Three things wait on him**: submit the entry (no record it is done, plan
> said ~Sep 1) · the worry trigger (a locked channel, so raising the bar is his
> call) · the banding decision at full window size.
> ⚠ ~~**One limitation that is not a bug**: a resize scrambles the AGENT's own
> text. The scape is provably right.~~ **WRONG, and corrected 2026-09-03 (s13).
> Both halves of that damage were the HOST'S**, found after he reported it a
> third time. (1) The resize clear emitted `ESC[2K` with no SGR of its own, and
> an erase fills with the CURRENT background -- so it painted rows in whatever
> colour Claude was mid-draw, and those rows scrolled into scrollback as a wall
> of black. (2) `drop` is a MAIN-screen correction and was applied on the
> ALTERNATE screen too, where nothing slides on a shrink, so the clear walked up
> into the transcript and at a big enough shrink ate the input box.
>
> **Both were invisible to the instrument that exonerated the host**: `screen`
> discarded SGR ("colour is not what these tests are about") so a row filled
> with colour read as blank, and every resize test ran `AltScreen: false` while
> production runs on the alternate screen. The lesson is not about resizing --
> it is that a harness which no-ops a platform behaviour, or never enters the
> mode production runs in, will keep returning a clean bill of health.
>
> It remains TRUE that Claude Code emits nothing on a resize, so anything the
> TERMINAL moves stays moved until a keystroke. That part was never the bug.

**Name: `xscapes`** (decided 2026-09-01 with the repo; asciiscapes and iixscapes are out; carried through everything 2026-09-03). A cozy ASCII "thinking screen" for terminal AI agents. While Claude Code (or any agent) works, a small living scene runs WITH it — a shoreline whose sea rises with the work, and a companion animal — and nudges the user, visually and with a sound, when the agent finishes or needs input.

Owner: Lucas (cinematographer + solo dev, LA). Solo build, AI-assisted. Start from this file; do not re-derive decisions already made here.

## Hackathon context

*(Facts below verified 2026-08-30 against the live site and Lucas's logged-in screenshots. Do not re-derive.)*

- Commons "Make Waiting for AI Fun" hackathon (`commonsmade.com/hackathons`). **Aug 27 – Sep 17, 2026**, three weeks. One entry per builder, editable until the deadline. Submit an early version by ~Sep 1 and iterate.
- Prizes: **1st $20,000 · 2nd $8,000 · 3rd $4,000 · 4th–19th $500 each · Vault $20,000 + 80% of all revenue generated.** $60,000 total. Treat payout as a bonus; the product is worth building regardless.

### Judging criteria — five weighted. Build against this.

| weight | criterion | the question asked |
|---|---|---|
| **30%** | Waiting Experience | Does it genuinely make waiting better? |
| **25%** | Originality | Is the idea new or surprising? |
| **20%** | Fit | Does it feel native to using AI agents? |
| **15%** | Repeatability | Would users enjoy it again? |
| **10%** | Execution | Does the prototype clearly prove the idea? |

Judges **to be announced**. Their framing line, verbatim: *"The biggest opportunity may be building the company that owns the waiting layer for AI."*

**What the rubric implies, and it should drive every tradeoff:**
- **Execution is only 10%.** Polish is nearly worthless. A rough scene that lands the idea beats a smooth one that doesn't. Never trade scope for finish.
- **Originality + Fit = 45%, and both favour the terminal build.** "Native to using AI agents" is precisely what a scene living in tmux beside Claude Code *is*; a browser tab is a worse answer. On a vibe-coding platform, near every rival entry will be a web app — a Go binary in a terminal is the only one of its kind in the pile. This is the moat.
- **Lead the submission with the layer, not the landscape.** The event protocol + pluggable adapters *is* "the waiting layer for AI" they say they want to fund. Most entries will be one clever screen. Frame xscapes as the protocol with a reference implementation, and the cozy scene as what it looks like.
- 30% Waiting Experience is the notification doing its job: the nudge has to genuinely beat a terminal bell.

### Building on Commons' platform — resolved

**No rule requires it.** What exists is a **token leaderboard** incentivising platform use; it is marketing, and it appears nowhere in the rubric. **Build the TUI here.** Use Commons only for genuinely web-shaped side pieces — the landing/submission page and a browser scape gallery for judges who will not install a binary — which feeds the leaderboard for free.

- Commons free tier: **150 AI credits/day**, **600/month**, resets 8/31/2026. Wallet **$0 and stays $0** — paying to climb a leaderboard that isn't scored buys nothing. (Daily-vs-monthly interaction is contradictory; `BALANCES`/`USAGE`/`MODELS` panels hold the real per-model rates and are still unexpanded.)
- Model for the web pieces: **Qwen3 Coder Next on Quick · $**, fed specs written here so it transcribes rather than designs. **Never Expert · $$$** (it pays premium to plan work already planned). Skip Superspeed · $$ (that buys latency, not quality). DeepSeek V4 Pro only if something truly needs reasoning.
- Code mode is chat-based (`/chat/new`); no sandbox or terminal observed. GitHub `donlucasx` is connected to the account.

### Demo video
Highest-leverage deliverable: 45–60s real screen recording, no narration — prompt in Claude Code → scene reacts → companion knocks → Enter → back in the agent. **Record in Terminal.app** — that follows from his 2026-09-02 ruling, and it reverses what this line said before. The scape is now built for the 256-colour cube, so a truecolor terminal would show a picture no user gets.

## Locked decisions

### The encoding rule (decided 2026-08-30)

**The water is the work. The sky is the world.** Sea state always means the
agent; sky, light and time always mean reality. Nothing crosses. Every variable
gets its own perceptual channel, and a channel may be bound to the real world
only if nothing else needs it.

| variable | channel | state |
|---|---|---|
| agent busy, how hard | **swells**: how many are travelling, how tall, whitecaps above half | done |
| something is broken | **the companion**, not the weather: ears back, hunched, tail flat, amber eyes; persists until it clears | done |
| context remaining | **moon** phase *and* altitude; numeric readout of what is LEFT under the moon from **40% used** (his ruling 2026-09-05), warm &ldquo;NN% left&rdquo; from 85%. **The disc's edge is the HUE RIM** (the outer ring one tone darker in its own hue; his pick 2026-09-05 from "The Moon, Four Ways" over the quad edge, the shadowless sun and the night halo, which stay as study switches) | done &mdash; the readout was decided in s6 and marked done, but was never in the live scene until 2026-09-05 (`drawReadout`, `ReadoutFrom`) |
| time of day | **sky colour**, real wall clock | done |
| weather | **deferred, not rejected** &mdash; no rain, clouds, fog or sync in v1; the thinking is parked in `ideas.md` | deferred 2026-08-30 |
| needs you | **bubble**, rare: needs_input, error, done. Nothing else | done — distinct cues shipped 2026-08-31: ask = warm SOLID box + alert pose; done = cool DOTTED knock + content `^ ^` pose, bounded by DoneHold |
| companion identity | **coat + face**: cream/slate/sage/mauve/charcoal, nose, toes, inner-shadow ears, whiskers | done — **cream, as shipped (his pick 2026-09-05)**; the other coats stay as built options |
| what it is doing now | **text written in the sand**, newest brightest, older fading as the tide takes them | done — anchored to the waterline, degrades by dropping whole pieces when narrow. **The lower beach falls away to black (`DefaultSandFade` = 1.0, locked 2026-09-01)**: contrast on the newest line 132→204 at midday, 148→204 at night, and equal at every hour, so legibility stops depending on the clock. Ink is sampled from the PAINTED background per row, never the palette's nominal sand. |
| todos completed | **star count** | done 2026-09-02 &mdash; a constellation in the upper sky, `*` for each finished todo. ~~`&#8728;` for each outstanding one, so it reads *n of N*~~ &mdash; **the ring is GONE (his ruling 2026-09-05: "discard the ring altogether, it's not clear what it means")**; the sky says *n*. Position is fixed by index and seed so a star lights where it always was. Held at a visibility floor like the moon: a completed todo is a fact about the AGENT and `StarVis` is 0 at noon. ⚠ **TodoWrite has been called ZERO times in the whole recorded history** &mdash; 13,682 tool events &mdash; so today it only lights from `xscapes emit todo` or the demo cycle. |
| subagents | **kittens** | done — `agent_id`/`agent_type`, counted live; **a kitten stays at least 60 s after its start** before it swims off (`reduce.KittenDwell`; his ruling 2026-09-05, after twice not seeing subagents that lived 7–16 s) |

Rejected and why: session-elapsed as its own variable (the real clock covers it,
and a session-relative sky lies about the world); weather carrying activity (it
was carrying two masters, which is the collision this rule exists to prevent);
tide-as-time-since-input, horizon-glow, driftwood-count (all fine, none
load-bearing &mdash; eight variables is already at the edge of what reads without
a legend).

**Encode in coverage, count or position &mdash; never in rate.** Activity was first
mapped to wave *speed* and idle was indistinguishable from flat-out, because a
glance is the entire budget and a screenshot has no motion at all.

### Six slots every scape must fill
light, sky, motion (the WATER, not the air), surface, accumulator, companion. Any scape providing all six
works with the whole encoding system for free. The rainy window currently fails
on companion &mdash; fix by putting the cat on the inside sill.

### Layout &mdash; ⚠ SUPERSEDED 2026-09-01, agent goes INSIDE the scape

**Everything in this section describes the side-by-side design, which he has
ruled out**: *"The entire Claude experience should happen within the xscape,
not next to it."* It still ships and still works, so it is kept as the record
of what is running today, but it is no longer the target. The replacement is
mocked in `-overlay` and not built. Read `RESUME.md` ▶ NEXT before touching it.

### Layout (the shipped, superseded design) &mdash; supersedes the popup default below
The scape is a full pane and carries the activity tail written into the sand.
Not a popup: a popup covers the session, and the user should always be able to
see what the agent is doing.

**The composition is MIRRORED (locked 2026-08-31).** Companion on the RIGHT,
litter growing leftward, sand written from the left margin, moon at 0.28. The
cat sprite is flipped rather than moved, because its tail sweeps from the right
hip and would otherwise be pinned to the frame edge. Measured, this is also more
robust in a narrow pane than the old left-anchored layout, not less: the cat
stays whole to 14 columns instead of clipping at 16, and the sand survives to 30
columns instead of 34. `-mirror=false` still renders the old layout.

**The waterline reserves a beach**: never fewer than five rows of sand, which is
identical to the old flat 80% at 24 rows and above and only bites on a short
pane, where 80% used to leave a single row.


**Lifecycle**
- Session-long, not task-long: launch once when the agent session starts, exit when it ends. Tasks modulate the scene (weather rises while working, settles to a resting state while waiting on the user). Day/night cycle spans the session.
- Persistence is tiny: companion identity/memory + a per-repo seed (keyed by git root) so the same repo always gets the same landscape shape. Nothing else accumulates. No decay.

**Placement**
- Lives alongside the agent via tmux: popup (`display-popup`) on think by default, split pane as an option. Zellij floating pane later.
- Lucas uses Terminal.app with no tmux. `xscapes claude` must bootstrap tmux itself (claude in main pane, scape beside). Only dependency: tmux via brew. No-tmux fallback: second Terminal window via `osascript`. Never seize the agent's TTY.

**Stack**
- Go + bubbletea/lipgloss, single static binary. No Node, no Python.
- Design target 80×24; must look fine at 40×12.
- Glyphs: Unicode blocks + braille for water/fire, ASCII fallback. **The 256-colour cube on EVERY terminal** (decided 2026-09-04 when Ghostty's truecolor picture differed: the scene is designed for the cube, and indices 16-255 are the same everywhere). `XSCAPES_COLOR=truecolor` opts into the untuned raw palette.
- **256 is not greyscale** — it is 216 colours plus 24 greys. A dark palette
  collapses to grey because the colour cube has almost no resolution below luma
  25 (4 entries, all pure blue) against 108 above 150. So the darkness lives in
  the BACKGROUNDS and the colour lives in the GLYPHS, which are the bright part
  of the frame. `term.GlyphBoost` (2.6, `XSCAPES_CHROMA` to override) lifts
  glyph chroma toward a target of 100 before quantising; near-neutrals below
  chroma 30 are left alone so the companion does not turn orange.
- **Never emit an ANSI index below 16.** Those sixteen are the only colours a
  terminal profile can repaint; 16-255 are fixed by the xterm standard, which is
  what makes the scene look the same on every profile. Proven exhaustively.
- Renderer is a real layer/alpha model from day one: three depth layers per scene (far α≈0.3 slow parallax sparse glyphs, mid α≈0.6, near α=1.0 dense fastest). Alpha = fg color blended toward bg/lower layer; quantize to palette on 256-color. Companion always in near layer. Weather modulates far-layer alpha (fog = one number).
- Frame pacing and clean redraws matter more than scene count; judges will screen-record it.

**Agent integration**
- Two layers: a tiny event protocol + thin adapters. Engine listens on a Unix socket with a JSON-lines file fallback. Adapters translate each agent into the protocol.
- Events: `session_start`, `prompt`, `tool` (with kind: read|write|edit|search|shell|web|subagent|todo|mcp), `error`, `test_pass`, `test_fail`, `compact`, `needs_input`, `done`, `session_end`.
- Adapter 1: Claude Code hooks (SessionStart/End, UserPromptSubmit, PreToolUse/PostToolUse, Notification, Stop, SubagentStop, PreCompact). Lucas already has a Stop/Notification hook that beeps — reuse it as the first adapter and replace the beep.
- Adapter 2: generic "watch this process" fallback (busy = alive + output activity; done = prompt back). Test targets: Claude Code, Kimi, Hermes. Verify what hooks Kimi/Hermes expose before writing adapters.
- Coarse output parsing is acceptable for agents without hooks.

**Scapes (three for v1, each with exactly one toy)**
- Shore: waves (layered sine + foam), stars, moon, sand. Toy: skip a stone.
- Campsite: campfire (Doom-fire algorithm) foreground, night sky, tent. Toy: log on the fire.
- Rainy window: rain streaks, blurred city lights, lightning. Companion-less scape — weather delivers the notification. Toy: wipe fog off the glass.
- Which survive is TBD after seeing them rendered.

**Activity mapping** *(supersedes the per-tool weather taxonomy, discarded 2026-08-30)*
- **Weather is deferred, not rejected.** No rain, clouds, fog, storms or
  real-weather sync in v1; the ideas are parked in `ideas.md` for after the
  vocabulary settles. The sky is time of day, stars and the moon; nothing else.
  This removes a network dependency, a location permission, and the risk of
  atmospheric motion being misread as agent activity.
- **Tool events do not each get their own visual.** There is no read=rain,
  shell=thunder taxonomy to memorise. Every tool event feeds ONE aggregate
  activity level, which drives the swells: how many are travelling and how tall.
- **Identity lives in the sand, not in the scene.** The activity tail names the
  tool and the file. The sea says how much; the sand says what. That split is
  why no per-tool vocabulary is needed.
- `error` / `test_fail` put the companion into its worried pose, which persists
  until it clears. `needs_input` and `done` raise the bubble.
- Growth: files touched leave driftwood on the sand, placed by path hash.
  Completed todos light stars. Subagents appear as kittens.

**Companion**
- 12 columns by 7 rows as shipped (`CatBody` 24x28 px drawn at half width, quarter height; the brief said "3–4 lines" until 2026-09-05, corrected on his ruling, and a new companion is designed against 12x7), resident not subject. States: resting, working (small idle motion), needs-you (walks to the edge nearest the agent pane, shows `!`).
- Shortlist: cat and bird. Fox and wisp as alternates. Decide after rendering real frames.
- Companion is global (same across repos); has a name.

**Interactivity (minimal — it must never become a game)**
- v1 ships exactly three interactions: pet the companion, plant a tree, one physics toy per scape.
- One keypress, zero commitment, never competes with the agent for attention.

**Notification**
- Distinct cues for `done` vs `needs_input`. Companion delivers it where present; weather (lightning/ember pop/foam) where not.
- One on-theme sound per scape (bell, bird, thunder) via `afplay`/`paplay`. Silent during work; ambient audio is off by default and optional.
- Keypress on notification focuses the agent pane (tmux `select-pane`). Fallback when pane hidden: tmux `display-message` + OS notification.
- One-line status on tap ("in auth/ for 3 min").

**Shipping**
- GitHub release binaries + brew tap. MIT, public from day one.
- State in `~/.config/xscapes/`.

## Open questions
1. ~~Name~~ **ANSWERED 2026-09-01: `xscapes`.**
2. Companion: cat or bird (decide from rendered frames).
3. Which of the three scapes survive.
4. ~~Must the build happen on Commons' platform?~~ **Answered 2026-08-30: no.** Token leaderboard only, absent from the rubric.
5. ~~Does the Commons "Code" tab have a sandbox/terminal?~~ **Chat-only as far as tested.** GitHub account is linked; repo import untested.
6. Per-model credit rates on Commons, and how 150/day reconciles with 600/month.

## Milestone 1 (target: ~Sep 1, submittable) — status 2026-09-01
1. ✅ Go module, 80×24 canvas, layer/alpha renderer with truecolor→256 fallback. (No bubbletea; stdlib only.)
2. ✅ Shore scape, three layers, working/resting states.
3. ✅ Event protocol (socket + file), `xscapes emit <event>` CLI for testing.
4. ✅ Claude Code hook adapter + `xscapes install claude`.
5. ✅ `xscapes claude` launcher — bootstraps tmux, joins an existing session,
   osascript fallback, `-print` dry run. Verified in a real tmux.
6. ◑ needs_input/done cues ✅ (distinct: warm solid ask box vs cool dotted knock, and
   two sounds via `internal/notify`, edge-detected so a 60s nag rings once).
   **Keypress-focuses-agent-pane is NOT built** — the live loop reads no keys.
7. ✅ Cat companion — five states (resting, working, needs-you, done, worried).
8. ✅ README + MIT LICENSE written 2026-09-01; every command in it verified by running it.
   Published 2026-09-01 at github.com/donlucasx/xscapes; clone-and-build verified from the public repo.

## Milestone 2
Campsite + rainy window, weather mapping, plant-a-tree, per-repo seed, generic process adapter, Kimi/Hermes tests, bird companion, demo video.

## Working style
- Direct, honest tradeoffs over optimism. Verbatim-ready commands.
- Lucas will push back when advice doesn't match what he sees; take that seriously.
- Keep it lean. If a feature makes it feel like a game, cut it.
