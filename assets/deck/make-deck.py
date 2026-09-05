#!/usr/bin/env python3
"""Rebuild the submission deck from its template and the site's animated clips.

The three scene slides are not mockups: they are the clips the engine renders
into site/anim (`go run . -site site` writes the frame pages, then
`python3 site/make-gifs.py` encodes them). This copies the clips next to the
deck and writes the deck; the PDF print shows each clip's first frame.

    go run . -site site && python3 site/make-gifs.py
    python3 assets/deck/make-deck.py         # writes assets/deck/index.html
    "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --headless=new \
      --no-pdf-header-footer --virtual-time-budget=8000 \
      --print-to-pdf=assets/deck/xscapes-deck.pdf "file://$PWD/assets/deck/index.html"

Clips used: hero, ask, done. The template is deck.tpl.html next to this file;
the deck's rules are assets/brand/guidelines.html.
"""
import os, shutil, sys

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.dirname(os.path.dirname(HERE))
CLIPS = ('hero', 'ask', 'done')

def main():
    anim = os.path.join(HERE, 'anim')
    os.makedirs(anim, exist_ok=True)
    for name in CLIPS:
        src = os.path.join(ROOT, 'site', 'anim', name + '.gif')
        if not os.path.exists(src):
            sys.exit(f'missing {src}: run go run . -site site && python3 site/make-gifs.py first')
        shutil.copyfile(src, os.path.join(anim, name + '.gif'))
    t = open(os.path.join(HERE, 'deck.tpl.html')).read()
    cover = os.path.join(ROOT, 'site', 'anim', 'cover.html')
    if not os.path.exists(cover):
        sys.exit(f'missing {cover}: run go run . -site site first')
    t = t.replace('@@COVER@@', open(cover).read())
    if '@@' in t:
        sys.exit('a placeholder was left unfilled')
    cut = t.index('</style>') + len('</style>')
    out = ('<!doctype html>\n<html lang="en">\n<head>\n<meta charset="utf-8">\n'
           '<meta name="viewport" content="width=device-width,initial-scale=1">\n'
           + t[:cut] + '\n</head>\n<body>' + t[cut:] + '\n</body>\n</html>\n')
    with open(os.path.join(HERE, 'index.html'), 'w') as fh:
        fh.write(out)
    print(f'wrote assets/deck/index.html ({len(out)} bytes) with clips {CLIPS}')

if __name__ == '__main__':
    main()
