# Publishing the submission page on Commons

The entry picker on commonsmade.com lists only PUBLISHED Commons apps
(`research/commons-submission.md`). This page is that app. It is one static
file; nothing on Commons needs to be designed, only served.

## Regenerate the page

```sh
go run . -site site            # writes site/index.html from the template, and the clip frame pages into site/anim
python3 site/make-gifs.py      # captures and encodes site/anim/*.gif (Chrome, Pillow, gifsicle)
open site/index.html           # look at it first
```

The copy lives in `site/template.html`. The clips are the demo turn in
`site.go` (`demoTurn`, shared with `-wired`) folded through the real reducer, as
a 256-colour terminal shows it, at 10 frames a second. Edit the template, rerun,
done. The page is `index.html` plus `anim/*.gif` (about 5.4 MB together).

## On Commons

Model: **Qwen3 Coder Next on Quick**. Not Expert, not Superspeed.

Prompt, verbatim, then paste the whole of `site/index.html` after it:

```
Create a static single-page web app. The entire app is the HTML document
below. Serve it exactly as written: do not change, reformat, restyle,
summarise or add to any of its content, markup or CSS, and do not add any
framework, script, analytics or extra page. The file is self-contained.
```

Then: publish it (manage page -> publish, visibility public), open
`/hackathons` signed in, tick it under "Your entry", and confirm the text
"Entry submitted." appears. Screenshot that.

The clips are separate files, so the paste alone is not the whole page. They
are served from GitHub Pages (his pick, 2026-09-05): `site/publish.sh` pushes
`site/index.html` and `site/anim/*.gif` as the `gh-pages` branch, which Pages
serves at https://donlucasx.github.io/xscapes/. Run it after every rebuild of
the page so the live copy matches the repo. Before pasting, point the page's
clips at it and paste THAT file:

```sh
sed 's#src="anim/#src="https://donlucasx.github.io/xscapes/anim/#g' site/index.html > /tmp/commons-paste.html
```

The entry is the Commons app; the Pages URL is where a judge who will not use
Commons can still see the page.

## Before submitting

- `research/commons-submission.md`, last section: open the rules signed in and
  check they still match. Two minutes.
- The clips on the page are the current renderer's output (the hue-rim disc,
  the readout, kittens on the sand and in the water). Regenerate them after
  any renderer change; the deck copies three of them (`assets/deck/make-deck.py`).
