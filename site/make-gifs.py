#!/usr/bin/env python3
"""Encode the site's animated clips from the frame pages the engine writes.

    go run . -gifs site/anim          # one HTML page per clip, frames stacked
    python3 site/make-gifs.py         # captures each page once, slices, encodes

Each page holds every frame of a clip under a magenta separator (site/anim/
manifest.json says how many). One headless Chrome screenshot per page, sliced
on the separators, quantised to one 256-colour palette per clip without
dithering (the frames are flat cells), written as GIF, then gifsicle -O3.
Needs Pillow and gifsicle; Chrome at the usual path.
"""
import json, os, subprocess, sys
from PIL import Image

HERE = os.path.dirname(os.path.abspath(__file__))
ANIM = os.path.join(HERE, 'anim')
CHROME = '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'
MAGENTA = (255, 0, 255)

def capture(page, png, height):
    subprocess.run([CHROME, '--headless=new', '--disable-gpu', '--use-mock-keychain',
                    '--password-store=basic', '--hide-scrollbars', '--force-device-scale-factor=1',
                    f'--window-size=900,{height}', f'--screenshot={png}', f'file://{page}'],
                   check=True, capture_output=True)

def slices(im):
    """Frame boxes between magenta separators: (x0, x1) from the separator's run, y ranges between."""
    w, h = im.size
    px = im.load()
    seps = []
    y = 0
    while y < h:
        if px[2, y] == MAGENTA:
            y0 = y
            while y < h and px[2, y] == MAGENTA:
                y += 1
            seps.append((y0, y))
        else:
            y += 1
    if len(seps) < 2:
        sys.exit('no separators found; is the page rendered?')
    x1 = 0
    y0 = seps[0][0]
    while x1 < w and px[x1, y0] == MAGENTA:
        x1 += 1
    boxes = [(0, seps[i][1], x1, seps[i + 1][0]) for i in range(len(seps) - 1)]
    return boxes

def main():
    manifest = json.load(open(os.path.join(ANIM, 'manifest.json')))
    for clip in manifest:
        name, n, fps = clip['name'], clip['frames'], clip['fps']
        page = os.path.join(ANIM, name + '.html')
        png = os.path.join(ANIM, name + '.png')
        capture(page, png, n * 292 + 8)
        im = Image.open(png).convert('RGB')
        boxes = slices(im)
        if len(boxes) != n:
            sys.exit(f'{name}: sliced {len(boxes)} frames, manifest says {n}')
        frames = [im.crop(b) for b in boxes]
        pal = frames[0].quantize(colors=255, method=Image.Quantize.MEDIANCUT, dither=Image.Dither.NONE)
        q = [f.quantize(palette=pal, dither=Image.Dither.NONE) for f in frames]
        raw = os.path.join(ANIM, name + '.raw.gif')
        q[0].save(raw, save_all=True, append_images=q[1:], duration=int(1000 / fps), loop=0,
                  optimize=False, disposal=1)
        out = os.path.join(ANIM, name + '.gif')
        subprocess.run(['gifsicle', '-O3', '--no-warnings', '-o', out, raw], check=True)
        os.remove(raw); os.remove(png)
        print(f'{name}.gif  {n} frames  {frames[0].size[0]}x{frames[0].size[1]}  {os.path.getsize(out)//1024} KB')

if __name__ == '__main__':
    main()
