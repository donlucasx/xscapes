#!/usr/bin/env python3
"""Rebuild the submission deck from its template and the real frames.

The three scene slides are not mockups: they are frames rendered by the engine
into site/index.html (`go run . -site site`), lifted verbatim. When the renderer
changes, regenerate the page first, then run this, then print the PDF.

    go run . -site site                      # re-renders the five frames
    python3 assets/deck/make-deck.py         # writes assets/deck/index.html
    "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --headless=new \
      --no-pdf-header-footer --virtual-time-budget=8000 \
      --print-to-pdf=assets/deck/xscapes-deck.pdf "file://$PWD/assets/deck/index.html"

Frames used: 0 (fourteen seconds into a turn), 3 (dusk, needs permission),
4 (night, done). The template is deck.tpl.html next to this file; the deck's
rules are assets/brand/guidelines.html.
"""
import os, re, sys

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.dirname(os.path.dirname(HERE))
SITE = os.path.join(ROOT, 'site', 'index.html')
FRAMES = (0, 3, 4)

def main():
    site = open(SITE).read()
    css = re.search(r'(\.p0\{color:[^\n]*)\n', site).group(1)
    pres = re.findall(r'<div class="win"><pre[^>]*>(.*?)</pre></div>', site, flags=re.S)
    if len(pres) < max(FRAMES) + 1:
        sys.exit(f'expected at least {max(FRAMES)+1} frames in site/index.html, found {len(pres)}')
    t = open(os.path.join(HERE, 'deck.tpl.html')).read()
    t = t.replace('@@FRAMECSS@@', css)
    for i in FRAMES:
        t = t.replace(f'@@FRAME{i}@@', pres[i])
    if '@@' in t:
        sys.exit('a placeholder was left unfilled')
    cut = t.index('</style>') + len('</style>')
    out = ('<!doctype html>\n<html lang="en">\n<head>\n<meta charset="utf-8">\n'
           '<meta name="viewport" content="width=device-width,initial-scale=1">\n'
           + t[:cut] + '\n</head>\n<body>' + t[cut:] + '\n</body>\n</html>\n')
    with open(os.path.join(HERE, 'index.html'), 'w') as fh:
        fh.write(out)
    print(f'wrote assets/deck/index.html ({len(out)} bytes) from frames {FRAMES}')

if __name__ == '__main__':
    main()
