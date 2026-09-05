#!/usr/bin/env python3
"""Regenerate the xscapes brand files from Geist Mono 700.

The mark is one terminal cell with the x drawn in reverse video. This script
takes the real glyph outlines from Geist Mono 700, lays them on a 600 x 1200
cell (0.6 em by 1.2 em, the same cell the guidelines page uses), and writes
the lockups, the marks and the avatars as SVG with paths, so no viewer needs
the font.

    python3 assets/brand/make-brand.py            # writes into assets/brand/

Needs fontTools (pip install fonttools) and network access for the font, or a
local GeistMono-700.ttf next to this file.
"""
import os, sys, urllib.request
from fontTools.ttLib import TTFont
from fontTools.pens.svgPathPen import SVGPathPen
from fontTools.pens.transformPen import TransformPen

HERE = os.path.dirname(os.path.abspath(__file__))
FONT = os.path.join(HERE, 'GeistMono-700.ttf')
FONT_URL = 'https://fonts.gstatic.com/s/geistmono/v6/or3yQ6H-1_WfwkMZI_qYPLs1a-t7PU0AbeHaL55T.ttf'

INK, GROUND, LIGHT = '#eeeeee', '#121212', '#eeeeee'   # 255, 233, 255
GOLD, GOLD_LIGHT = '#d7af5f', '#af8700'                # 179, 136 (the cell on light)

def main():
    if not os.path.exists(FONT):
        urllib.request.urlretrieve(FONT_URL, FONT)
    f = TTFont(FONT)
    if 'fvar' in f:
        from fontTools.varLib import instancer
        f = instancer.instantiateVariableFont(f, {'wght': 700})
    upm = f['head'].unitsPerEm
    asc, desc = f['hhea'].ascent, f['hhea'].descent
    gs, cmap = f.getGlyphSet(), f.getBestCmap()
    adv = f['hmtx'][cmap[ord('x')]][0]
    cell_h = round(1.2 * upm)
    baseline = (cell_h - (asc - desc)) / 2 + asc      # where a terminal puts it

    def glyph(ch):
        pen = SVGPathPen(gs)
        gs[cmap[ord(ch)]].draw(TransformPen(pen, (1, 0, 0, -1, 0, 0)))
        return pen.getCommands()
    paths = {c: glyph(c) for c in 'xscapes'}

    def lockup(ink, ground, cell, bg=None, word='scapes'):
        w, h = adv * (1 + len(word)), cell_h
        out = [f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {w} {h}" width="{w}" height="{h}">']
        if bg: out.append(f'<rect width="{w}" height="{h}" fill="{bg}"/>')
        out.append(f'<rect width="{adv}" height="{h}" fill="{cell}"/>')
        out.append(f'<path transform="translate(0 {baseline})" d="{paths["x"]}" fill="{ground}"/>')
        for i, c in enumerate(word):
            out.append(f'<path transform="translate({adv * (i + 1)} {baseline})" d="{paths[c]}" fill="{ink}"/>')
        out.append('</svg>')
        return '\n'.join(out)

    def mark(ground, cell, bg=None):
        return lockup(ground, ground, cell, bg=bg, word='')

    def avatar(ground, cell, size=512):
        s = size * 0.74 / cell_h
        x, y = (size - adv * s) / 2, (size - cell_h * s) / 2
        return (f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {size} {size}" width="{size}" height="{size}">'
                f'<rect width="{size}" height="{size}" fill="{ground}"/>'
                f'<g transform="translate({x:.2f} {y:.2f}) scale({s:.6f})"><rect width="{adv}" height="{cell_h}" fill="{cell}"/>'
                f'<path transform="translate(0 {baseline})" d="{paths["x"]}" fill="{ground}"/></g></svg>')

    files = {
        'lockup-dark.svg':       lockup(INK, GROUND, INK),              # transparent, for ground 233
        'lockup-light.svg':      lockup(GROUND, LIGHT, GROUND),         # transparent, for ground 255
        'lockup-dark-bg.svg':    lockup(INK, GROUND, INK, bg=GROUND),
        'lockup-light-bg.svg':   lockup(GROUND, LIGHT, GROUND, bg=LIGHT),
        'lockup-gold-dark.svg':  lockup(INK, GROUND, GOLD),             # Reverse, the alternate
        'lockup-gold-light.svg': lockup(GROUND, LIGHT, GOLD_LIGHT),
        'mark-dark.svg':         mark(GROUND, INK),
        'mark-light.svg':        mark(LIGHT, GROUND),
        'mark-gold-dark.svg':    mark(GROUND, GOLD),
        'mark-gold-light.svg':   mark(LIGHT, GOLD_LIGHT),
        'avatar.svg':            avatar(GROUND, INK),
        'avatar-gold.svg':       avatar(GROUND, GOLD),
    }
    for name, svg in files.items():
        with open(os.path.join(HERE, name), 'w') as fh:
            fh.write(svg + '\n')
    print(f'cell {adv}x{cell_h}, baseline {baseline}, {len(files)} files written to {HERE}')

if __name__ == '__main__':
    sys.exit(main())
