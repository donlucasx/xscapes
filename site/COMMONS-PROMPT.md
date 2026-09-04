# Publishing the submission page on Commons

The entry picker on commonsmade.com lists only PUBLISHED Commons apps
(`research/commons-submission.md`). This page is that app. It is one static
file; nothing on Commons needs to be designed, only served.

## Regenerate the page

```sh
go run . -site site        # reads site/template.html, writes site/index.html
open site/index.html       # look at it first
```

The copy lives in `site/template.html`. The five frames are rendered by the real
reducer from the demo turn in `site.go` (`demoTurn`, shared with `-wired`), as a
256-colour terminal shows them. Edit the template, rerun, done.

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

If the chat rejects a paste of this size, host the same file on GitHub Pages
from the repo (`site/index.html`, Settings -> Pages -> deploy from `main`,
folder `/site`) and give Commons a one-line app that redirects there. The
entry is the Commons app either way.

## Before submitting

- `research/commons-submission.md`, last section: open the rules signed in and
  check they still match. Two minutes.
- The GIFs in `assets/frames` are not on the page on purpose: the frames are
  the real renderer's output, and the GIF is from before the distinct cues.
