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
