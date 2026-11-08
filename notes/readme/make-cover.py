#!/usr/bin/env python3
"""README cover: the deck's cover inverted to the brand's light palette, as a GIF.

Uses the same six sea frames the deck cycles (site/anim/cover.html, 0.7 s a frame),
renders each through headless Chrome, and joins them. Run from the repo root:
    /opt/homebrew/bin/python3.14 notes/readme/make-cover.py
Writes assets/readme/cover.gif.
"""
import os, re, subprocess, tempfile
from PIL import Image

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '..'))
CHROME = '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'
W, H = 1600, 640
src = open(os.path.join(ROOT, 'site/anim/cover.html')).read()
pres = re.findall(r'<pre[^>]*>(.*?)</pre>', src, re.S)
assert len(pres) == 6, len(pres)

PAGE = '''<!doctype html><html><head><meta charset="utf-8">
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Geist+Mono:wght@400;700&family=Geist:wght@400;500;600&display=swap">
<style>
html,body{margin:0;width:%dpx;height:%dpx;overflow:hidden;background:#eeeeee}
.sea{position:absolute;inset:0;display:flex;align-items:center;justify-content:center}
.sea pre{margin:0;font:700 16px/1 "Geist Mono",monospace;color:#b8b8b8;white-space:pre;
  -webkit-mask-image:linear-gradient(90deg,transparent,#000 12%%,#000 88%%,transparent)}
.front{position:absolute;inset:0;display:flex;flex-direction:column;align-items:center;justify-content:center}
.lockup{font:700 150px/1.2 "Geist Mono",monospace;color:#121212;letter-spacing:0}
.blk{display:inline-block;width:1ch;line-height:1.2em;text-align:center;vertical-align:top;background:#121212;color:#eeeeee}
.tag{margin-top:34px;font:500 38px/1.2 "Geist",sans-serif;color:#121212;letter-spacing:-0.01em;
  background:#eeeeee;padding:6px 18px}
.sub{margin-top:10px;font:400 22px/1.2 "Geist Mono",monospace;color:#626262;background:#eeeeee;padding:2px 12px}
</style></head><body>
<div class="sea"><pre>%s</pre></div>
<div class="front">
  <div class="lockup"><span class="blk">x</span>scapes</div>
  <div class="tag">Cozy ASCII scenes that show what your agent is doing</div>
  <div class="sub">for Claude Code, Kimi and Hermes</div>
</div></body></html>'''

tmp = tempfile.mkdtemp()
frames = []
for i, p in enumerate(pres):
    html = os.path.join(tmp, f'f{i}.html'); png = os.path.join(tmp, f'f{i}.png')
    open(html, 'w').write(PAGE % (W, H, p))
    prof = os.path.join(tmp, f'prof{i}')
    subprocess.run([CHROME, '--headless=new', f'--user-data-dir={prof}', '--hide-scrollbars',
                    f'--window-size={W},{H}', '--virtual-time-budget=4000',
                    f'--screenshot={png}', 'file://' + html],
                   check=True, capture_output=True)
    frames.append(Image.open(png).convert('RGB'))
out = os.path.join(ROOT, 'assets/readme'); os.makedirs(out, exist_ok=True)
for i, f in enumerate(frames): f.save(os.path.join(tmp, f'check{i}.png'))
pal = [f.quantize(colors=32, method=Image.Quantize.MEDIANCUT) for f in frames]
pal[0].save(os.path.join(out, 'cover.gif'), save_all=True, append_images=pal[1:],
            duration=700, loop=0, optimize=True, disposal=1)
print(tmp, os.path.getsize(os.path.join(out, 'cover.gif')))
