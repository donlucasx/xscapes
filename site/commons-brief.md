# xscapes: brief for the Commons Space agent

You are building the hackathon ENTRY for xscapes, not xscapes itself. Read all of this before doing anything.

## What xscapes is

xscapes is a cozy ASCII thinking screen for terminal AI agents. It is a Go program that runs ANY terminal agent inside a living landscape in the user's own terminal: Claude Code through its hooks, anything else through `xscapes inside <cmd>`. The shoreline is the first scape, not the product: the tide comes up the beach as the agent works, the last actions are written in the sand, a crab lives on the beach and walks up close when the agent needs you, crablets arrive for subagents and walk into the water when they report back, stars light as the plan gets done, and the moon sinks as the context fills. It installs with one command from a public repo. It runs in Terminal.app or Ghostty. It cannot run in a browser and must not be rebuilt as a web app.

- Source, public, MIT: https://github.com/donlucasx/xscapes
- The page, live: https://donlucasx.github.io/xscapes/
- No demo video. The page's animations are the demonstration: real renderer output as text, not mockups.

## What to build here: an app that serves that page, byte for byte

The entry is the page at https://donlucasx.github.io/xscapes/ . It is ONE self-contained HTML file of about 5.5 MB (465 KB gzipped): every animation is embedded as text, and the only thing it loads is one Google Fonts stylesheet. It is too big to paste here, so the app fetches it.

Build a Worker whose root route proxies that URL:

- On a request for `/` or `/index.html`, fetch `https://donlucasx.github.io/xscapes/` on the server and return its body unchanged, with the header `content-type: text/html; charset=utf-8`. Let the platform handle compression.
- Do not cache it for longer than 10 minutes. The page is republished up to the deadline and the app must show the latest one.
- Any other path returns 404.
- Name the app `xscapes`.

If the platform reserves `/` for a static file and will not let a route answer it, use this instead, as the entire static page:

```html
<!doctype html><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1,viewport-fit=cover"><title>xscapes</title>
<style>html,body{margin:0;height:100%;background:#0b0b0b}iframe{border:0;width:100%;height:100%;display:block}</style>
<iframe src="https://donlucasx.github.io/xscapes/" title="xscapes" allow="fullscreen"></iframe>
```

## Do not

- Do not copy, rewrite, restyle, summarise or improve the page. It is served as is.
- Do not add a framework, a build step, a router, analytics, your own fonts, a favicon, a footer, or a second page.
- Do not download or re-encode anything. The page carries its own animations.
- Do not clone, rebuild, port or recreate the GitHub repo. It is Go and terminal hooks; nothing in it runs in a Space.
- Do not plan. This is a ten-line proxy. There is nothing to design.
- Do not generate a submission guide, a checklist, or any other page.
- Do not ask about the hackathon rules. They are handled.

You are done when the app's URL renders the same as https://donlucasx.github.io/xscapes/ : the splash reads "xscapes" over "a thinking screen for terminal agents", the headline under it reads "Cozy ASCII scenes that react to your agent while it works", and the terminal window under the headline is animating. Do not offer improvements.

## After the build (for Lucas, not for the agent)

DONE 2026-09-16: https://xscapes.vibe.commonsmade.com is published (public, source shared) and submitted; the hackathon home reads "xscapes · ENTRY SUBMITTED." What the flow actually was, for the record:

1. Code, fresh chat on Default · Quick, paste this file's message, Send, then Start building. The platform first deploys its own "first shot" template; the agent deletes it and writes the proxy (it dropped `[assets]` so the Worker answers `/`). ~5 credits.
2. Publish (top right of the preview). Visibility cannot be changed BEFORE the first publish; it reads Public after. Share source code: on.
3. The app dock then offers "Enter into hackathon" (or: hackathon home, Your entry, Submit, pick xscapes).
4. Verified from outside: the entry serves the gh-pages bytes plus one platform-injected `/__commons/analytics.js` line; `/index.html` the same; other paths 404; `cache-control: public, max-age=600`. So a republish of gh-pages (`go run . -site site && sh site/publish.sh`) reaches the entry within 10 minutes with no Commons step. Entries close Sep 17, 23:59 UTC; the build is changeable until then, and the entry keeps serving the LIVE page after.
