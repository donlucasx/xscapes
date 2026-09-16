#!/usr/bin/env python3
"""Build the comparison page: every candidate WAV and the shipped pair embedded
as data URIs, so what the browser plays is byte-for-byte what afplay would.
Writes one self-contained HTML file to the path given as argv[1]."""
import base64, json, os, sys, html

HERE = os.path.dirname(os.path.abspath(__file__))
SHIPPED = os.path.join(HERE, "..", "..", "internal", "notify", "sounds")
out = sys.argv[1]

def b64(path):
    with open(path, "rb") as f:
        return "data:audio/wav;base64," + base64.b64encode(f.read()).decode()

import wave, math, struct
def stats(path):
    w = wave.open(path); n = w.getnframes(); sr = w.getframerate()
    x = struct.unpack("<%dh" % n, w.readframes(n))
    rms = math.sqrt(sum((v / 32768) ** 2 for v in x) / n)
    return round(n / sr * 1000), round(20 * math.log10(max(rms, 1e-9)), 1)

rows = []
a_ms, a_rms = stats(os.path.join(SHIPPED, "ask.wav")); d_ms, d_rms = stats(os.path.join(SHIPPED, "done.wav"))
rows.append({"key": "shipped", "name": "Shipped now", "tag": "baseline",
             "what": "The s28 droplet: one sine whose pitch glides down under a smooth envelope. No impact, and it falls where a real drop climbs.",
             "differs": "Ask and done are the same shape an octave apart.",
             "ask": {"src": b64(os.path.join(SHIPPED, "ask.wav")), "ms": a_ms, "rms": a_rms, "file": "~/.config/xscapes/sounds/ask.wav"},
             "done": {"src": b64(os.path.join(SHIPPED, "done.wav")), "ms": d_ms, "rms": d_rms, "file": "~/.config/xscapes/sounds/done.wav"}})
for m in json.load(open(os.path.join(HERE, "manifest.json"))):
    r = {"key": m["key"], "name": m["name"], "tag": "", "what": m["what"], "differs": m["differs"]}
    for k in ("ask", "done"):
        p = os.path.join(HERE, m[k]["file"])
        r[k] = {"src": b64(p), "ms": m[k]["ms"], "rms": m[k]["rms_db"], "file": "~/Documents/claude/xscapes/notes/s34-sound/" + m[k]["file"]}
    rows.append(r)

def row_html(i, r):
    n = i + 1
    return f"""
<section class="cue" id="cue-{r['key']}" data-key="{r['key']}">
  <div class="head">
    <h2><span class="n">{n}</span>{html.escape(r['name'])}{' <span class="tag">shipped</span>' if r['tag'] else ''}</h2>
  </div>
  <p class="what">{html.escape(r['what'])}</p>
  <div class="pair">
    <button class="play ask" id="play-{r['key']}-ask" data-key="{r['key']}" data-kind="ask" aria-label="Play {html.escape(r['name'])} ask">
      <span class="lbl">ask <kbd>{n}</kbd></span>
      <canvas width="600" height="72"></canvas>
      <span class="meta"><span>{r['ask']['ms']} ms</span><span>{r['ask']['rms']} dB</span></span>
    </button>
    <button class="play done" id="play-{r['key']}-done" data-key="{r['key']}" data-kind="done" aria-label="Play {html.escape(r['name'])} done">
      <span class="lbl">done <kbd>&#8679;{n}</kbd></span>
      <canvas width="600" height="72"></canvas>
      <span class="meta"><span>{r['done']['ms']} ms</span><span>{r['done']['rms']} dB</span></span>
    </button>
  </div>
  <p class="differs">{html.escape(r['differs'])}</p>
</section>"""

cmds = "\n".join(f"afplay {r['ask']['file']}\nafplay {r['done']['file']}" for r in rows)
data = json.dumps({r["key"]: {"ask": r["ask"]["src"], "done": r["done"]["src"]} for r in rows})

page = f"""<title>The Cue, Six Ways</title>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&display=swap">
<style>
/* The submission page's own tokens (site/template.html), so the candidates
   are auditioned inside the product's look rather than a generic one. Dark on
   purpose and single-theme, like the site; everything painted explicitly. */
:root{{--g:#121212;--s:#1a1a1a;--ln:#2e2e2e;--ln2:#242424;--d:#8a8a8a;--dd:#bcbcbc;--i:#eeeeee;
--acc:#87afff;--warm:#d7af5f;
--mono:"JetBrains Mono",ui-monospace,"SF Mono",Menlo,monospace;
--lh:1.55rem;--measure:66ch;--deck:17px;--body:15px;--meta:13px}}
*{{box-sizing:border-box}}
html{{color-scheme:dark}}
body{{margin:0;background:var(--g);color:var(--i);font-family:var(--mono);font-size:var(--body);line-height:var(--lh);
padding-block:calc(var(--lh)*2) calc(var(--lh)*3);padding-inline:16px}}
main{{max-width:var(--measure);margin:0 auto;display:flex;flex-direction:column;gap:var(--lh)}}
h1{{font-size:var(--deck);font-weight:700;margin:0;letter-spacing:0;text-wrap:balance}}
h1 .cur{{display:inline-block;width:.6em;height:1em;background:var(--i);vertical-align:-.15em;margin-left:.1em;animation:blink 1.1s steps(1) infinite}}
@keyframes blink{{50%{{opacity:0}}}}
@media (prefers-reduced-motion:reduce){{h1 .cur{{animation:none}}}}
p{{margin:0;color:var(--dd)}}
p.lede{{color:var(--dd)}}
.hint{{color:var(--d);font-size:var(--meta)}}
kbd{{font-family:inherit;font-size:11px;color:var(--d);border:1px solid var(--ln);border-radius:3px;padding:0 .35em;line-height:1.4;vertical-align:.05em}}
hr{{border:0;border-top:1px solid var(--ln2);margin:0}}
.cue{{display:flex;flex-direction:column;gap:calc(var(--lh)*.5);padding-block:calc(var(--lh)*.5) var(--lh);border-top:1px solid var(--ln2)}}
.cue.playing{{border-top-color:var(--ln)}}
.cue h2{{font-size:14px;font-weight:700;text-transform:uppercase;letter-spacing:.06em;margin:0;color:var(--i)}}
.cue h2 .n{{display:inline-block;min-width:2ch;color:var(--d);font-weight:500}}
.cue h2 .tag{{font-weight:500;color:var(--warm);text-transform:none;letter-spacing:0;margin-left:1ch;font-size:var(--meta)}}
.what{{color:var(--dd)}}
.differs{{color:var(--d);font-size:var(--meta)}}
.pair{{display:grid;grid-template-columns:1fr 1fr;gap:12px}}
@media (max-width:520px){{.pair{{grid-template-columns:1fr}}}}
.play{{appearance:none;font:inherit;color:inherit;text-align:left;cursor:pointer;background:var(--s);border:1px solid var(--ln);border-radius:4px;
padding:8px 10px 6px;display:flex;flex-direction:column;gap:4px;min-width:0;transition:border-color .12s}}
.play:hover{{border-color:var(--d)}}
.play:focus-visible{{outline:2px solid var(--acc);outline-offset:2px}}
.play.on{{border-color:var(--dd)}}
.play .lbl{{font-size:var(--meta);color:var(--d);display:flex;justify-content:space-between;align-items:baseline;text-transform:uppercase;letter-spacing:.06em}}
.play.ask .lbl{{color:var(--warm)}}
.play.done .lbl{{color:var(--acc)}}
.play canvas{{width:100%;height:36px;display:block;max-width:100%}}
.play .meta{{font-size:11px;color:var(--d);display:flex;justify-content:space-between;font-variant-numeric:tabular-nums}}
pre{{margin:0;background:var(--s);border:1px solid var(--ln2);border-radius:4px;padding:10px 12px;font-size:12px;line-height:1.5;color:var(--dd);overflow-x:auto}}
.foot{{color:var(--d);font-size:var(--meta)}}
</style>
<main>
  <h1>The cue, six ways<span class="cur"></span></h1>
  <p class="lede">The sound xscapes makes when the agent asks, and when it finishes. Six candidates against the one that ships. Every player here holds the exact bytes afplay would play, so this is the sound, not a preview of it.</p>
  <p class="hint">Click a strip to play it. Keys <kbd>1</kbd>&ndash;<kbd>7</kbd> play an ask, <kbd>&#8679;</kbd>+key plays the done. The drawn strip is the amplitude, on one time scale for every cue, so a longer sound is a longer strip. Ask is the warm one, done the cool one, as on screen.</p>
  <p class="hint">What the shipped drop lacks: a real drop is a tick and then a ring that climbs, because the bubble the impact pinched off sinks away from the surface. Rows 2 to 5 are built from that. Rows 6 and 7 are not water at all, in case natural did not mean wet.</p>
{''.join(row_html(i, r) for i, r in enumerate(rows))}
  <hr>
  <p class="foot">To hear them through the product's own path, in Terminal.app at the volume the scape uses:</p>
  <pre>{cmds}</pre>
  <p class="foot">Made by notes/s34-sound/make.py, stdlib only, 16-bit mono 44.1 kHz, every file at the shipped cues' peak. The bird is written at half level: a whistle at a drop's peak is 8 dB louder.</p>
</main>
<script>
(function(){{
  var SRC = {data};
  var audio = {{}};
  var PXMS = 1.0; // canvas px per ms: 600 px = 600 ms, the longest cue here
  function decode(uri){{
    var b = atob(uri.split(',')[1]); var n = b.length; var u = new Uint8Array(n);
    for (var i = 0; i < n; i++) u[i] = b.charCodeAt(i);
    var dv = new DataView(u.buffer); var sr = dv.getUint32(24, true);
    // find the data chunk rather than assuming a 44-byte header
    var off = 12, len = 0;
    while (off + 8 <= n) {{ var id = String.fromCharCode(u[off],u[off+1],u[off+2],u[off+3]); var sz = dv.getUint32(off+4, true);
      if (id === 'data') {{ off += 8; len = sz; break; }} off += 8 + sz + (sz & 1); }}
    var count = Math.floor(len / 2); var out = new Float32Array(count);
    for (var j = 0; j < count; j++) out[j] = dv.getInt16(off + 2*j, true) / 32768;
    return {{sr: sr, x: out}};
  }}
  function draw(canvas, pcm, color){{
    var ctx = canvas.getContext('2d'); var W = canvas.width, H = canvas.height, mid = H/2;
    ctx.clearRect(0,0,W,H);
    ctx.fillStyle = '#2e2e2e'; ctx.fillRect(0, mid, W, 1);
    var perPx = pcm.sr / 1000 / PXMS; // samples per canvas px
    ctx.fillStyle = color;
    for (var px = 0; px < W; px++) {{
      var a = Math.floor(px*perPx), b = Math.floor((px+1)*perPx); if (a >= pcm.x.length) break;
      var m = 0; for (var i = a; i < b && i < pcm.x.length; i++) {{ var v = Math.abs(pcm.x[i]); if (v > m) m = v; }}
      var h = Math.max(1, m * (H - 4)); ctx.fillRect(px, mid - h/2, 1, h);
    }}
  }}
  var colors = {{ask: '#d7af5f', done: '#87afff'}};
  var buttons = document.querySelectorAll('.play');
  Array.prototype.forEach.call(buttons, function(btn){{
    var key = btn.getAttribute('data-key'), kind = btn.getAttribute('data-kind');
    var uri = SRC[key][kind];
    try {{ draw(btn.querySelector('canvas'), decode(uri), colors[kind]); }} catch (e) {{}}
    btn.addEventListener('click', function(){{ play(key, kind); }});
  }});
  function play(key, kind){{
    var id = key + '-' + kind; var a = audio[id];
    if (!a) {{ a = new Audio(SRC[key][kind]); audio[id] = a; }}
    var btn = document.getElementById('play-' + key + '-' + kind); var sec = document.getElementById('cue-' + key);
    Array.prototype.forEach.call(document.querySelectorAll('.play.on'), function(b){{ b.classList.remove('on'); }});
    Array.prototype.forEach.call(document.querySelectorAll('.cue.playing'), function(s){{ s.classList.remove('playing'); }});
    btn.classList.add('on'); sec.classList.add('playing');
    a.onended = function(){{ btn.classList.remove('on'); sec.classList.remove('playing'); }};
    try {{ a.currentTime = 0; a.play(); }} catch (e) {{}}
  }}
  var order = Array.prototype.map.call(document.querySelectorAll('.cue'), function(s){{ return s.getAttribute('data-key'); }});
  document.addEventListener('keydown', function(ev){{
    if (ev.metaKey || ev.ctrlKey || ev.altKey) return;
    var digits = '1234567', shifted = '!@#$%^&';
    var i = digits.indexOf(ev.key); var kind = 'ask';
    if (i < 0) {{ i = shifted.indexOf(ev.key); kind = 'done'; }}
    if (i < 0 && ev.shiftKey) {{ i = digits.indexOf(ev.code.replace('Digit','')); kind = 'done'; }}
    if (i < 0 || i >= order.length) return;
    ev.preventDefault(); play(order[i], kind);
  }});
}})();
</script>
"""
with open(out, "w") as f:
    f.write(page)
print(out, os.path.getsize(out), "bytes")
non_ascii = sum(1 for ch in page if ord(ch) > 127)
print("non-ascii chars:", non_ascii)
