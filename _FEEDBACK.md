# asciiscapes — feedback log

Lucas's feedback, **verbatim**, newest section at the bottom. One section per session.
Never paraphrase. Read the relevant section before editing anything it covers.

## 2026-08-28 — session 0, verbatim

- *"create a new subfolder /asciiscapes/ within ~/documents/claude/ and prepare to start a new
  fresh project w its own memory"*
  ⇒ Scaffold only. Scope deliberately left undefined; do not infer the project from its name.

## 2026-08-29 — session 1, verbatim

- *"considering calling the project iixscapes, or xscapes. no need to rename project yet, we can
  brainstorm a bit, the name will come as we build it."*
  ⇒ Name stays **asciiscapes** for now. New candidates on the list: **iixscapes**, **xscapes**.
    Do NOT rename anything yet. Name gets decided late, from the built thing.
- *"Also, do a deep dive into the site https://commonsmade.com/, the hackathon, their ai platform and
  how can we better approach it"*

## 2026-08-29/31 — sessions 2–4, verbatim

**Companion design**
- *"i think we need to do a thorough exploration of ascii companions in terms of aesthetics, pic
  which ones we like, how they animate, etc, how do we want to approach it"*
- *"i like both 1 and 2 for the visual approach, even braile could be fun. I do want to see a broad
  sheet."*
- *"they look kinda crazy (not good). cant you reproduce any of the references I gave you within our
  constrains? can we expand on our constrains? what limits us?"*
  ⇒ Cause: kaomoji are PROPORTIONAL-font art; and one-char-per-cell is the lowest-resolution medium.
- *"i like the quadrants for the companion, it helps it stand out from the rest"* ⇒ **LOCKED.**
- *"Can we add some eyes to this one, perhaps a slight animation?"*
- *"i liked the cat before much better. also, I dont see this one moving"*
  ⇒ Sitting cat is THE design; side-view walk is not the plan. And: **ship GIFs, not PNGs** — a
    screenshot cannot show motion and an HTML page needs a server.
- *"Im confident you could animate that exact design... when it sits/stands it already reveals
  little paws"* ⇒ pose set, not a walk cycle.
- *"companion should have multiple different actions/behaviors, and they should happen both at
  random, as well as triggered by specific processes or events"*

**Layout and encoding**
- *"i think the user should always be able to see what the agent is working on"*
  ⇒ Killed the popup default. Activity tail written INTO THE SAND.
- *"i like the moon idea for the context, but idk if its clear enough for the user. perhaps we need
  a more explicit visual cue - one that translates across all different ascii scenes"*
- *"should we include an actual percentage meter on the context meters? find a sensible middle point
  between artistic and informative"* ⇒ treatment **C**: silent, appears at 65%, brightens at 85%.
- *"yes, i love that- actual time of the day works. What if its rainy outside tho, are we sync'ing
  time AND weather?"*
- *"i like the waves to showcase how hard its working, and I like keeping time of the day and
  weather as an experiential feature"* ⇒ gave the rule: **the water is the work, the sky is the world.**
- *"i think if something is broken, the companion should communicate it (not the [weather])"*
  ⇒ Fixed a real flaw: weather had been carrying busy AND broken.
- *"can we have the ocean react to how hard its working with waves?"* ⇒ travelling swells, not a
  thickening noise field.
- *"discard weather altogether"* then *"i would not delete all the ideas around weather, clouds or
  even storms, but ok to defer for now"* ⇒ **defer, do not erase.** Parked in `ideas.md`.

**Subagents**
- *"can we do baby kittens?"* ⇒ yes, chosen over distant lights.
- *"Id say, use the large ones when running up to 5 sub agents, the small 5x3 when up to 8..."*
- *"I dont like the agents that are 'in the background' or in the water. They dont read well...
  If a better solution, then dont mix and match, just have all sub agents be the same size depending
  on the amount"* ⇒ **uniform litter size, chosen by count, all on the near layer at full alpha.**
- *"could any of the sub agents be swimming? so we could place them further in the back, perhaps
  poking out of the water and waddling around gently?"*
- *"its ok as long as layers are clear when they move behind or infront of each other"*
  ⇒ occlusion rims.
- *"also should probably test some animations w everything in motion to ensure it reads correctly"*
  ⇒ `cat-everything.gif`. Judge the whole scene, never one element alone.

**Naming**
- *"considering calling the project iixscapes, or xscapes"* — still undecided; name comes late.

## 2026-08-31 — session 5, verbatim

- *"resume work on asciiscapes"*
  ⇒ Whole session ran off the ▶ NEXT list with no further steer. Nothing new from
    Lucas to log. Recording it so the next session knows the session's direction
    was inferred from RESUME.md, not given — and that the three defects now at the
    top of ▶ NEXT (sand on the water, tail colliding with the cat, done vs
    needs_input identical) are MY findings from `assets/frames/wired.png`, not his.
    He has not seen the wired scene yet. Show it to him before building on it.

- *"looking at the wired.png example. Maybe we should have our companion on the right side of the
  frame, so the code doesnt cover him, and align sub agents from right to left. What do you think?
  and populate them intentionally in areas where the code wont touch- unless its too many agents,
  then its ok for some to be covered by the code. Assess this- would it present other issues if
  placed opposite on the screen?"*
  ⇒ Companion moves RIGHT; kittens fill RIGHT-TO-LEFT; placement is intentional around the sand
    text, and overlap is ACCEPTABLE once the litter is large. Asked for an assessment first, not
    an implementation.

## 2026-08-31 — session 6, verbatim

**Mirroring the composition**
- *"aligned, lets mirror the composition- shall we mock it up first?"* ⇒ mock BEFORE building.
- *"also, are we accounting for the user experiencing the terminal on different aspect ratios?
  are our scenes reactive? do they adjust to the widght/height/aspect ratio of the window?"*
  ⇒ Exposed that `-live` never resized at all. See ▶ terminal fixes.
- *"if we mirror the composition, would we risk the companion being lost if the width of the
  terminal screen is too narrow?"* ⇒ Measured: NO, the opposite. Mirrored the cat tracks the
  edge and stays whole to 14 cols; pinned-left it clipped at 16.
- *"lets do both"* / *"explain in simpler terms what these decisions mean"* / *"ok, lets do both"*
  ⇒ Reserved beach on short panes + tail degrades by dropping whole pieces. BOTH SHIPPED.
  ⇒ And: when asked to explain, drop the jargon and SHOW rendered before/after. He asked once.

**Seeing it**
- *"where can i see it"* ⇒ built `wired-turn.gif`. A screenshot cannot show motion.
- *"im on the app, can I see them here"* ⇒ published an artifact. He is often on the app, not
  the terminal. Offer a viewable link, not just a local path.
- *"need to test it on the terminal now. Can I do so within an existing session like this one?
  or recommend to start a new one?"*
- *"run it"* ⇒ I could not: `-live` needs a real TTY and the Bash tool has none. Say so.

**Terminal reality — three bugs he found by running it**
- *"[screenshots] ran the demo, this is how it looks when I open it (it doesnt fill the frame,
  no color?) ... when shrank too much it breaks ... when I stretch it again it glitches"*
  ⇒ ONE bug: `stty size` reads from stdin, which exec.Command gives /dev/null. Always 80x24.
- *"looks much better. Still glitches a bit when resizing the screen (during the dragging) but
  settles correctly. I noticed the 'shine' effect on the ocean is on the right side, but it
  should be on the left, under the moon- correct? Regarding Colour: whats the best approach to
  optimize the visual experience? can we override the terminal color settings when running
  this? I have different terminal windows open working w different colors (mostly variations of
  solids). How do we ensure everyone is having the same visual experience?"*
  ⇒ He was right about the shine: `glitter()` hardcoded 0.72 while the moon moved to 0.28.
- *"if I understand correctly, the terminal.app is limited to 256 colors, right? but is it
  limited to greyscale? can we try an alternative approach to 256 using colors?"*
  ⇒ **He was right to push and I had been wrong.** 256 is not greyscale. Led to the glyph
    chroma boost. DO NOT let "the backgrounds cannot be coloured" become "256 cannot do colour".
- *"i like glyph chroma 2.6x best."* ⇒ **LOCKED at 2.6.**

**Companion**
- *"can we take a look at the companion. Are there any more face details we can add to make it
  look more like a cat? also try it on more solid colors? perhaps taking inspiration on claudes
  mascot"*
- *"love these, im leaning towards cream, slate or charcoal. Terracota and Ginger look great but
  they are too similar to claude's."* ⇒ **Claude's own colours are OUT.**
- *"I like this face variation for the cat, can we try additional details? any way we can get
  some whiskers on the cat? or any other feline characteristics? expand the character study"*
- *"im still leaning towards the original one, or the one with a nose we did last session. can we
  attempt a few more subtle variations? also try again the chest bib and toe tips. Let me see
  each variation on its own and then all together as well, keep same colors and try some new
  ones"* ⇒ **SUBTLE. And show each feature ISOLATED before combining.**
- *"fav colors so far: cream, fog, slate, sage, mauve and charcoal. as far as features, I like
  the nose and the toes. Can we try an alternative approach to the whiskers and inner ears?"*
- *"ok, colorwise lets keep cream, slate, sage, mauve and charcoal in the mix. I think
  cream/sage/slate win because they stand out best. regarding features: whiskers dont quite read
  yet, can we test the'strokes' closer to the body instead of having a gap? and also prob closer
  to each other as well. Inner ears dont quite read- the inner ear detail should be within the
  existing ears, not outside. Lets see a couple variations for each, with and without the toes"*
- *"whiskers are better, they shouyld be flush, but all 4 whiskers should land around the nose,
  not the eyes. maybe a mix of flush and long. for the inner ears, the inner shadow works best."*
  ⇒ **Inner ears SETTLED: inner shadow.**
- *"whiskers are not it yet. they all need to be connected to the body, and they should be closer
  to the nose, not eye level. Maybe bottom whiskers are longer than the top ones, or viceversa,
  try a couple more alts"*
  ⇒ Root cause found on the 4th attempt: the head is NOT centred in its sprite, so any FIXED
    cell offset connects on one side and gaps on the other. Whiskers now find the fur.
    **Lesson: three rounds were spent moving a number when the placement model was wrong.**

- *"lets /wrap and revisit this on a fresh session. Will look at the companion-study.png and get
  back to u"* ⇒ Companion is PAUSED awaiting his pick. Nothing is defaulted.

## 2026-08-31 — session 7, verbatim

- *"resume work on asciiscapes"*
- Asked for the companion pick (coat / whiskers / toes / or defer). His four answers, verbatim:
  - Coat: *"pull up the latest companion study"*
  - Whiskers: *"need to revise this one"*
  - Toes: *"tbd, let me see with and without still"*
  - Defer: chose **"Not decided — build on"** (leave companion paused; work the done vs
    needs_input distinct-cues gap).
  ⇒ Companion stays PAUSED, still nothing defaulted. He wants the CURRENT study in front of him
    (it already renders with/without toes), expects another whisker revision round after he
    looks, and meanwhile the build continues on the locked-brief gap.

- *"is this the file we are both looking at? I can give you feedback based on this, or lmk if
  theres a more recent"* ⇒ confirmed same file (the full-length recapture).

- *"whiskers aint working- there should be a total of 4, 2 on each side, and they should be
  connected to the head (not the tail), leveled around the nose (not the eyes, not the chin).
  We love the ear shadows and toes. Lets lock in the whiskers and then decide wether we are
  adding all 3 details or a combination"*
  ⇒ The whisker SPEC, crisp: exactly 4, 2 per side, attached to the HEAD's fur, vertically
    on the nose line. Right-side whiskers must not touch/read as the tail. Ear shadows and
    toes are LOVED (not yet locked as included). Order of decisions: whiskers first, then
    which of the 3 details ship together.

- *"none of your whiskers hit the mark. Here is a very rough draft I made, where I drew the
  whiskers in red (for you to see them). You see how they are attached to the body around the
  nose level, and the top are longer than the bottom? re generate a couple versions based on
  this guide"* (+ a marked-up screenshot of the NONE portrait)
  ⇒ HE DREW THE ANSWER. Measured off his ink, in cells: both strokes per side hug the NOSE
    LINE (upper at the nose row's vertical middle, lower at the nose cell's BOTTOM EDGE --
    half a cell down, not a full row); both start flush at the fur edge; top ~2 cells, bottom
    ~1.3; and his top-right stroke runs STRAIGHT ACROSS the raised tail -- the whisker passes
    the tail (drawn by skipping solid cells), it does not stop at it. The four length
    variants from the previous rounds are all dead.

- *"whiskers: they are all too long, and for all the alts, the bottom whisker is sitting to
  low. Review my reference once again"*
  ⇒ Re-measured his ink as fractions of the NOSE CELL: top stroke at 56%, bottom at 89% --
    the bottom is INSIDE the nose cell, a third of a cell below the top. The overline on the
    next row rendered at ~112%: half a cell lower than his ink. No two line glyphs on
    adjacent rows can sit a third of a cell apart; BRAILLE can (dot row 3 ~62%, dot row 4
    ~87%, same cell). Whiskers go braille, both lines inside the nose row. Reaches shrink:
    3-cell variants dead; snug (1/1) / guide (2/1) / taper (2/1, tip fading).

- *"you had it going fine w the lines before, issue was placement and length. I gave you a
  very easy guideline to follow, use it"*
  ⇒ Braille REJECTED. Solid lines were right; fix ONLY length and placement, per his draft:
    top '─' on the nose row, 2 cells; bottom '‾' tucked on the row below, 1 cell. And found
    while reverting: the STUDY portraits alone drew the cat one row lower than every other
    surface (`c.H-1-chh` vs the live scene's `c.H-2-chh`), which parked the nose row ON the
    waterline -- wave glyphs continued the top whisker ("too long") and the bottom whisker
    floated amid the waves ("too low"). Portraits now match the live composition. Lesson:
    when he rejects a render twice, audit the CONTEXT the thing is judged in, not just the
    thing.

- *"closer, lets try the upper whiskers a tad shorter, and the bottom nudged up a hair"*
  ⇒ Direction CONFIRMED, fine-tuning now. Top: outer cell becomes a HALF dash ('╶'/'╴'),
    2 cells -> ~1.5. Bottom: '‾' (overline, ~12% down its cell) -> '⎺' (scan line 1, at the
    very top edge) -- a couple pixels up, literally a hair.

- *"fit both whiskers within the nose block of the cat, currently the bottom whiskey is
  floating in the air. you see what I mean? try another variant where the lower whiskers are
  attached to the lower narrower block on the face of the cat. try variants with all the
  whiskers attached to the body, and keep the current one as a variant too"*
  ⇒ He is RIGHT and the cause is exact: both pairs anchor to the NOSE ROW's fur span, but
    the row below is two cells narrower, so the bottom pair hangs in open air. Fixes built
    as four variants: **current** (unchanged, for comparison) · **tucked** (bottom anchored
    to ITS OWN row -- the narrower block) · **double** and **double long** ('═', one glyph
    carrying two parallel strokes, so BOTH whiskers sit inside the nose block and both are
    attached). Lesson: anchor a detail to the row it is DRAWN on, not to a neighbouring row.

- *"switching models, do you still have full context?"* ⇒ Yes; a model switch keeps the
  session's conversation.

- *"let's save the progress on the character study and defer the decision for later. Let's keep
  working and keep the characters as is for now."*
  ⇒ Companion study PARKED. Nothing defaulted, `NewCat()` unchanged. Do NOT re-ask.

- *"create the repo on the donlucasx gh account, call it xscapes"*
  ⇒ **THE NAME IS DECIDED: `xscapes`** (from the candidate list asciiscapes / iixscapes /
    xscapes; the brief always said the name would be picked late, from the built thing).
    Public repo under `donlucasx`. The Go module path MUST match the repo URL or
    `go install` is broken for everyone, so module → `github.com/donlucasx/xscapes`, which
    also makes the built binary `xscapes`. Working directory, CLAUDE.md and RESUME.md still
    say "asciiscapes" -- ⏸ ASK before renaming those; he asked for a repo, not a migration.

- *"go ahead"* (to `install claude --apply`, which writes his live `~/.claude/settings.json`)
  ⇒ INSTALLED 2026-09-01. His 4 existing hooks survived byte-for-byte, secret guard
    included; statusLine untouched; 12 xscapes hooks added; backup written to
    `~/.config/asciiscapes/backups/`. **First real hook events in the project's life**, and
    they arrived within seconds across THREE live sessions at once, with no restart needed.
    Verified end to end in tmux: his own commands appeared in the sand ("shell echo",
    "shell grep"), whitecaps on the sea, cat in the working pose.
    ⚠ **DOUBLE SOUND**: his own afplay beep hooks are still on Notification / Stop /
    PermissionRequest alongside the xscapes knocks, so both fire. The brief always said
    "replace the beep" -- but they are HIS hooks and they still work when no scape is
    running, so this is his call, never a silent removal.

- *"why is it split screen?"* (with a screenshot of the launcher's two panes)
  ⇒ I MISREAD IT as the horizontal horizon seam and spent a round measuring that (real finding,
    kept: a 124-luma step in one row at the horizon, and the sky paints 1.7 glyphs/row against
    the sea's 34.9). He meant the VERTICAL split. **Lesson: when a one-line question has two
    readings, ask or show both — do not pick one and measure it for a round.**

- *"I meant the vertical split. The entire Claude experience should happen within the xscape,
  not next to it"*
  ⇒ **THE SIDE-BY-SIDE LAYOUT IS OUT.** This supersedes the locked "scape is a full pane
    alongside the agent via tmux" decision, and it satisfies his older *"the user should always
    be able to see what the agent is working on"* BETTER, not worse. Mocked with `-overlay`
    before building: 83.5% of a real 145x43 Claude pane is blank.

- *"yes, should be within the scene, and extending the window's height should allow the user to
  preview more of the work that claude is doing, perhaps the taller the window the more sand
  below where we can see what claude is working on?"*
  ⇒ **Extra window height goes to BEACH, not sky.** `Shore.SkyRows/SandRows` added: sky held to
    a band, beach absorbs the rest — 4 lines of work history at 24 rows, 9 at 43, 16 at 60.
    Also fixes the empty-sky problem, since the emptiest region was taking 42% of every new row.
    ⚠ `reduce.TailLen` is still a hard 4 — it must become a function of the beach's rows.
    ⚠ Building the embedded agent = xscapes hosts Claude in a PTY and composites (a terminal
    emulator). Sizing Claude's viewport to (height − beach) solves the text collision by
    construction. NOT STARTED, awaiting his go-ahead.

- *"shall we have the sand turn into a gradient as it goes down so the text is clearer / most
  legible at the bottom? sand fades into black bkg"* then *"proceed with your recommendation"*
  ⇒ **SHIPPED at full strength, `DefaultSandFade = 1.0`.** Measured on the newest line (lowest,
    matters most): contrast 132→204 midday, 148→204 night, and EQUAL at every hour — a black
    bottom row is black at noon and midnight, which the flat beach never was. Full rather than
    0.8 because the agent is moving inside the scene and a true-black bottom merges with its
    own background instead of leaving a seam. Middle values are the WORST place to sit: the ink
    flips direction as the beach crosses mid-luma.
    ⚠ It exposed a defect it did not cause: `drawSand` chose ink from the palette's NOMINAL sand
    while writing onto the PAINTED background; once the beach could darken those diverged and a
    half fade at midday LOST contrast (132→129). Ink now samples the real background per row.
    ⚠ And a test that could not fail: `internal/scape/sandfade_test.go` MIRRORED drawSand's rule
    instead of running it, so it kept passing through all of this. Deleted; replaced by
    `sand_ink_test.go`, which renders a real frame and reads the pixels back.
    Verified on 256: the fade quantises to a clean monotonic descent, every index >= 16, and it
    ADDS depth — without it the bottom half of a tall beach was one flat colour repeated.

## 2026-09-01 — session 8, verbatim

- *"resume work on asciiscapes"*
  ⇒ Reported state, then put the blocking decision to him: build the embedded terminal or not.

- *"I thought we were designing the claude experience WITHIN the xscape all along. How do you
  recommend we proceed? explain the difference between the options provided"*
  ⇒ He is right that the DESIGN has been "within" since the ruling on 09-01, and the layout code
    for it was already built (`Shore.SkyRows/SandRows`, the beach growing with height, the sand
    fading to black). What did not exist was anything that HOSTS the agent — `xscapes claude`
    still opened it in a tmux pane beside the scape. The two options were two ways to make it
    real, and the difference is whether the scape shows THROUGH the agent (a real emulator,
    every cell ours, days of work, a parser bug corrupts Claude's UI) or AROUND it (a pty band,
    the agent's bytes untouched, 1-2 days).

- *"run it"* (the probe, before committing to either)
  ⇒ Measured, in `notes/claude-terminal-emissions.md`. Claude Code never uses the alternate
    screen and emits NO absolute row addressing during a turn — only relative moves, column
    addressing and erase-line. Two exceptions: `ESC[r` once at startup (swallow it) and `ESC[H`
    on resize (origin mode redirects it into the band). And the constraint that shaped the
    layout: scrollback survives only when the band is anchored at ROW 1, so nothing can be
    painted above the agent.

- *"yes"* (build the pass-through band)
  ⇒ **SHIPPED.** `xscapes inside [command]`. Verified in a real terminal with real Claude Code:
    a turn ran inside the band, 99 of 100 scrolled lines stayed reachable in the scrollback,
    resizes landed exactly on the computed split (40 rows → 27/13, 52 → 35/17, 22 → 14/8), and
    the exit hands the terminal back with the agent's screen intact.

## 2026-09-01 — session 9, verbatim

- *"do some research on ascii waiting screens for the terminal. are there any"*
  ⇒ Surveyed and written to `research/prior-art.md`. Yes, dozens, in four categories:
    idle screensavers (cmatrix, pipes.sh, asciiquarium, termsaver, ascsaver), per-command
    spinners (terminal-animations/tan), fake-activity generators (genact, hollywood), and
    agent-aware indicators — the only one that matters. That last category is NEW, all
    within the last year, and growing: pi-animations (26 animations for the pi agent,
    3 phases, inline + widget), claude-code-mascot-statusline (9 hook-driven states),
    tweakcc, and a whole `terminal-pet` GitHub topic (buddy at 102 stars). Four open
    upstream issues asking for exactly this (claude-code #66284/#29200/#35249,
    opencode #24937) — Anthropic has not shipped it.
    **The four gaps that keep xscapes distinct**: nobody runs the agent INSIDE the scene;
    nobody encodes the amount of work into what the scene shows (everyone keys off a
    3-to-9-value state enum); everything is a pet rather than a place; everything is
    host-locked TypeScript. ⚠ The risk that came out of it: the terminal-pet topic is
    crowded, and a skimming reader will file xscapes there on sight.

- *"ok save this in a subfolder /research/ and /wrap"*
  ⇒ `research/prior-art.md`, then this wrap.

## 2026-09-02 — session 10, verbatim

- *"1. it should"* (should `xscapes claude` mean the inside layout)
  ⇒ **SHIPPED.** `claude` hosts the agent inside the scape; `-beside` keeps the tmux layout;
    `inside <cmd>` hosts anything. `-print` works on all three.

- *"still opens side by side"* / *"what am I doing wrong"* / *"issue persists"*
  ⇒ Three of mine in a row, each found by him: `~/go/bin` is not on his PATH so my install went
    nowhere (his only binary is `~/.local/bin/xscapes`); `@latest` resolved to a tag predating the
    work (fixed by tagging v0.2.0/v0.2.1); and DECSTBM homes the cursor, so re-pinning the band on
    resize dropped his typing onto row 1.

- *"as we get deeper in the session the scape breaks"* → then the resize forensics
  ⇒ Root cause was NOT a rendering bug. Claude Code emits **zero bytes on resize** and places its
    input purely by relative moves; growing a window makes the terminal pull scrollback back in,
    pushing the agent's UI out of its band. Controls: no-resize works, plain Claude + resize works,
    hosted + resize fails. ⇒ **the alternate screen**, his call after the pros/cons.

- *"beach now reading like a beach, but its lost its shape/silhouette"* / *"we should be able to see
  the water receding"* / *"it should be sand, not black"* / *"now we have a lot of beach and not
  enough sea. Did u have to eat into the sea to extend the beach?"*
  ⇒ Four separate corrections to the same region, all mine to fix. Ended at: one flat sand tone that
    varies day to night, a ragged waterline with 2.9-7.5 rows of relief, the writing band INSIDE the
    beach's share rather than added to it, and no black seam.

- *"what if we use the moon as a sun. Same object, same variable. only thing we gotta change is the
  color (like the rest of the scape does)"*
  ⇒ **SHIPPED, and it is the only version that does not break the encoding** — context is carried by
    phase AND altitude, so a second body would need a second encoding for one variable.

- *"lets make it a tad taller so all the xscape layers can shine"*
  ⇒ Scape takes 9/20 of the window, was 2/5.

- ⭐ *"at this point I want to optimize the experience for terminal.app which should be the most
  used?"*
  ⇒ **RULING, and it reverses my standing recommendation.** I had been telling him to install a
    truecolor terminal. He is targeting Terminal.app instead. That makes the cube-exact colour
    approach — used so far only for the sand and the sun — the GENERAL rule for the whole scape,
    not a special case. See ▶ NEXT.

## 2026-09-02 -- session 11, verbatim

- *"resume work on asciiscapes"*
  ⇒ Picked up ▶ NEXT #1, the cube-exact colour work that follows from his Terminal.app ruling.
    No new steer from him this session; everything below is mine and is waiting on his eye.

  **The note I started from was wrong about the size of the problem, and I only found that by
  measuring instead of trusting it.** It said the sea's depth gradient collapsed onto one teal.
  True, and beside the point: across 48 half-hours of the day `SeaFar` landed on the GREYSCALE
  ramp at 40 of them and `SkyTop` at 36. The sea and the sky had no colour at all for most of a
  working day on his terminal. Daylight is 0/25 now.

  **Why nobody saw it, and this is the reusable part**: every HTML study in the repo renders true
  RGB. The previews have always shown a blue sea his terminal never painted. A preview that cannot
  produce the target's colours is not a preview of the target.

  **The moon test caught me.** A first pass put the dusk zenith at luma 116 and the sun's contrast
  fell to +37.9 against a floor of 40 -- the disc is painted INTO the sky, so every point of luma
  spent up there comes off the context readout. Fixed by taking the twilight zeniths dark again and
  lifting daylight `MoonVis` to 0.85; floor is +61.3 against +64.4 before.

  ⚠ **For him to judge: daylight is brighter and the sea is more turquoise.** Forced, not chosen --
  the cube keeps no blue below luma 60 except the electric pure-blue column already rejected.

- *"ok looking good. Some of the sky gradients in 256 dont look as good/smooth though, can we look
  for opportunities to make the gradients more subtle even in 256? truecolor looks great. Where
  could users experience the truecolor version?"*
  ⇒ The premise was smoothing and the real defect was HUE. Independent per-channel rounding turns a
    mid blue into a lavender -- rgb(48,112,170) becomes rgb(95,95,175) -- six rows of sky a day.
    `Index256Keeping` preserves any channel ordering the source states clearly, backgrounds only.

  **Three smoothing ideas, two rejected on sight after measuring well.** Shade blocks bought 11 ->
  14 tones and look like stipple. The 1x2 dither measured worse and its premise (the eye averages
  half a cell) is false at terminal size. Only the U+2580 split survived: 5% closer to the ramp,
  no texture.

  ⚠ **Two of my own instruments lied this session and I quoted one before catching it.** A regex
  sweep of the sky-ramp curve corrupted the source and I read four rows of numbers off a build
  failure. And the violet test recomputed the ramp instead of reading the render, so it could not
  see the fix at all -- it now goes through `ResolveAt`, and is proven red without the fix.
