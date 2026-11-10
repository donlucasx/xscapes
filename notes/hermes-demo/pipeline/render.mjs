// render.mjs <take dir> <edit.json> <out dir> [--stills t,t,...] [--from s --to s]
// The recorded pty bytes replayed into xterm.js inside a window, under a camera.
// edit.json: segs (video time -> recording time, piecewise linear, monotonic),
// camera keyframes (video s; col/row in cells, zoom 1 = the whole window),
// captions, tags, end card. Writes frames f00000.png... and timeline.json.
import { chromium } from 'playwright';
import fs from 'fs'; import path from 'path';
const here = path.dirname(new URL(import.meta.url).pathname);
const [take, editPath, out] = process.argv.slice(2).map(p => path.resolve(p));
const arg = (k) => { const i = process.argv.indexOf(k); return i > 0 ? process.argv[i + 1] : null; };
const stills = arg('--stills'); const from = +(arg('--from') ?? 0); const to = arg('--to');
const E = JSON.parse(fs.readFileSync(editPath, 'utf8'));
fs.mkdirSync(out, { recursive: true });
const load = (t, f) => fs.readFileSync(path.join(take, t, f), 'utf8').trim().split('\n').map(JSON.parse);
const TK = {}; for (const s of E.segs) TK[s.take] ??= { tty: load(s.take, 'tty.jsonl'), marks: load(s.take, 'marks.jsonl') };
if (E.stripPrompt) for (const k of Object.keys(TK)) for (const c of TK[k].tty) {
  const b = Buffer.from(c.b, 'base64').toString('latin1').split(E.stripPrompt).join('');
  c.b = Buffer.from(b, 'latin1').toString('base64');
}
const fontsDir = '/Users/lucasgarzoli/Documents/claude/xscapes/notes/trailer/fonts';
const fontCss = fs.readFileSync(path.join(fontsDir, 'fonts-local.css'), 'utf8')
  .replace(/url\(([^)]+\.woff2)\)/g, (_, f) => `url(${path.join(fontsDir, path.basename(f))})`);
const W = 1920, H = 1080, FPS = 30, COLS = 124, ROWS = 44, FS = E.fontPx ?? 26;
const endFrames = fs.readdirSync(path.join(here, 'endsrc')).filter(f => f.endsWith('.png')).sort();

// the timeline in video seconds
const segs = E.segs; const montageEnd = segs[segs.length - 1].v1;
const endAt = montageEnd + E.end.toBlack;
const endVidLen = endFrames.length / FPS;
const urlAt = endAt + E.end.urlAt, urlEnd = urlAt + E.end.url.length / E.end.urlCPS;
const total = urlEnd + E.end.hold + E.end.fade;
function at(v) { // video -> {take, r}
  for (const s of segs) if (v <= s.v1) {
    const x = Math.max(0, v - s.v0), d = s.v1 - s.v0;
    if (s.ramp) { const b = ((s.r1 - s.r0) - d) / (d * d); return { take: s.take, r: s.r0 + x + b * x * x }; } // starts at real speed, speeds up
    return { take: s.take, r: s.r0 + (s.r1 - s.r0) * x / d };
  }
  const s = segs[segs.length - 1]; return { take: s.take, r: s.r1 };
}
const rec = (v) => at(v).r;

const html = `<!doctype html><html><head><meta charset="utf-8">
<link rel="stylesheet" href="file://${here}/node_modules/@xterm/xterm/css/xterm.css">
<style>${fontCss}
html,body{margin:0;width:${W}px;height:${H}px;overflow:hidden;background:#0f0f10}
#stage{position:absolute;inset:0;opacity:var(--k,1);background:radial-gradient(ellipse at 50% 40%,#2a2730 0%,#16151a 55%,#0b0b0c 100%)}
#win{position:absolute;left:0;top:0;transform-origin:0 0;border-radius:14px;overflow:hidden;
 box-shadow:0 30px 90px rgba(0,0,0,calc(.65*var(--k,1))),0 0 0 1px rgba(255,255,255,calc(.09*var(--k,1)))}
#bar{opacity:var(--k,1);height:${Math.round(FS*1.6)}px;background:#2b2a2e;display:flex;align-items:center;position:relative;border-bottom:1px solid #1a191c}
#bar i{width:${Math.round(FS*.52)}px;height:${Math.round(FS*.52)}px;border-radius:50%;margin-left:${Math.round(FS*.42)}px;display:block}
#bar i:nth-child(1){background:#ff5f57;margin-left:${Math.round(FS*.7)}px}#bar i:nth-child(2){background:#febc2e}#bar i:nth-child(3){background:#28c840}
#bar b{position:absolute;left:0;right:0;text-align:center;font:500 ${Math.round(FS*.62)}px -apple-system,"Helvetica Neue",sans-serif;color:#9d9aa3}
#term{background:#0f0f10;padding:${Math.round(FS*.35)}px}

#cap{position:absolute;left:0;right:0;bottom:34px;text-align:center;font:500 ${E.captionPx}px "Geist Mono",monospace;color:#f2f2f2;
 text-shadow:0 2px 14px rgba(0,0,0,.9),0 0 3px rgba(0,0,0,.9);opacity:0;white-space:pre}
#cap span{background:rgba(10,10,12,.55);padding:.25em .7em;border-radius:.35em}
#tag{position:absolute;left:56px;top:44px;font:700 ${E.tagPx}px "Geist Mono",monospace;color:#f2f2f2;opacity:0;background:rgba(10,10,12,.82);padding:.3em .6em .35em;border-radius:.3em;box-shadow:0 0 0 1px rgba(255,255,255,.08)}
#tag small{display:block;font-weight:400;font-size:.5em;color:#bdbdbd;margin-top:.25em}
#endimg{position:absolute;inset:0;width:${W}px;height:${H}px;opacity:0}
#card{position:absolute;inset:0;opacity:0;background:#000}
#card img{position:absolute;inset:0;width:${W}px;height:${H}px}
#card div{position:absolute;left:0;right:0;top:${E.card?.y ?? 420}px;text-align:center;font:700 ${E.card?.px ?? 64}px/1.25 "Geist Mono",monospace;color:#f2f2f2;letter-spacing:-.01em;text-shadow:0 0 28px rgba(0,0,0,.95),0 0 6px rgba(0,0,0,.9)}
#card div small{display:block;font-weight:400;font-size:.62em;color:#d9d9d9;margin-top:.35em}
.h{color:#e8a23c}
#win2{position:absolute;left:0;top:0;transform-origin:0 0;border-radius:14px;overflow:hidden;opacity:0;box-shadow:0 30px 90px rgba(0,0,0,.65),0 0 0 1px rgba(255,255,255,.09)}
#win2 .bar2{height:${Math.round(FS*1.6)}px;background:#2b2a2e;display:flex;align-items:center;position:relative;border-bottom:1px solid #1a191c}
#win2 .bar2 i{width:${Math.round(FS*.52)}px;height:${Math.round(FS*.52)}px;border-radius:50%;margin-left:${Math.round(FS*.42)}px;display:block}
#win2 .bar2 i:nth-child(1){background:#ff5f57;margin-left:${Math.round(FS*.7)}px}#win2 .bar2 i:nth-child(2){background:#febc2e}#win2 .bar2 i:nth-child(3){background:#28c840}
#win2 .bar2 b{position:absolute;left:0;right:0;text-align:center;font:500 ${Math.round(FS*.62)}px -apple-system,"Helvetica Neue",sans-serif;color:#9d9aa3}
#term2{background:#0f0f10;padding:${Math.round(FS*.35)}px}
.lockup{font:700 1em/1.1 "Geist Mono",monospace;letter-spacing:0;color:#eeeeee;white-space:nowrap}
.lockup .blk{background:#eeeeee;color:#111;padding:0 .04em}
#plug{position:absolute;inset:0;background:#000;opacity:0;display:flex;flex-direction:column;align-items:center;justify-content:center;font:400 54px/1.3 "Geist Mono",monospace;color:#d6d6d6;text-align:center}
#plug .ha{font:700 110px/1.15 "Geist Mono",monospace;letter-spacing:.02em;background:linear-gradient(180deg,#ffe27a 0%,#f2b632 55%,#c98a2c 100%);-webkit-background-clip:text;background-clip:text;color:transparent;margin:.12em 0 .1em}
#plug code{font:400 34px "Geist Mono",monospace;color:#c98a2c;margin-top:1.1em;border:1px solid rgba(201,138,44,.55);padding:.35em .8em;border-radius:.3em}
#url{position:absolute;left:0;right:0;top:${E.end.urlY}px;text-align:center;font:400 ${E.end.urlPx}px "Geist Mono",monospace;color:#9a9a9a;white-space:pre;opacity:0}
#black{position:absolute;inset:0;background:#000;opacity:1;pointer-events:none}
</style></head><body><div id="stage"></div>
<div id="win"><div id="bar"><i></i><i></i><i></i><b id="title"></b></div><div id="term" style="position:relative"><div id="cur" style="position:absolute;background:#d8d8d8;display:none;z-index:9"></div></div></div>
<div id="win2"><div class="bar2"><i></i><i></i><i></i><b>xscapes hermes</b></div><div id="term2"></div></div><div id="split" style="position:absolute;inset:0;opacity:0;background:#0f0f10"></div><div id="tag"></div><div id="cap"><span></span></div>
<div id="card"><img id="cardimg"><div id="cardtxt"></div></div><div id="plug"></div><img id="endimg"><div id="url"></div><div id="black"></div>
<script src="file://${here}/node_modules/@xterm/xterm/lib/xterm.js"></script><script src="file://${here}/node_modules/@xterm/addon-webgl/lib/addon-webgl.js"></script></body></html>`;
fs.writeFileSync(path.join(out, 'page.html'), html);

const browser = await chromium.launch({ channel: 'chrome', args: ['--use-angle=metal', '--enable-gpu', '--ignore-gpu-blocklist', '--enable-unsafe-swiftshader'] });
const page = await browser.newPage({ viewport: { width: W, height: H } });
await page.goto('file://' + path.join(out, 'page.html'));
await page.evaluate(async ({ COLS, ROWS, FS, W2R }) => {
  await document.fonts.load(`400 ${FS}px "JetBrains Mono"`);
  await document.fonts.load(`500 40px "Geist Mono"`); await document.fonts.load(`700 40px "Geist Mono"`);
  const term = new Terminal({ cols: COLS, rows: ROWS, fontSize: FS, fontFamily: '"JetBrains Mono", Menlo, monospace',
    lineHeight: 1.0, cursorBlink: false, cursorStyle: 'block', cursorInactiveStyle: 'block', allowProposedApi: true, scrollback: 0, customGlyphs: true,
    theme: { background: '#0f0f10', foreground: '#d8d8d8', cursor: '#d8d8d8' } });
  term.open(document.getElementById('term'));
  const gl = new WebglAddon.WebglAddon(); term.loadAddon(gl);
  if (W2R) { const t2 = new Terminal({ cols: COLS, rows: W2R, fontSize: FS, fontFamily: '"JetBrains Mono", Menlo, monospace', lineHeight: 1.0, cursorBlink: false, scrollback: 0, customGlyphs: true, allowProposedApi: true,
      theme: { background: '#0f0f10', foreground: '#d8d8d8', cursor: '#0f0f10' } });
    t2.open(document.getElementById('term2')); t2.loadAddon(new WebglAddon.WebglAddon()); window.term2 = t2;
    window.feed2 = (s) => new Promise(r => t2.write(Uint8Array.from(atob(s), c => c.charCodeAt(0)), r)); }
  window.term = term;
  document.querySelector('.xterm-screen').appendChild(document.getElementById('cur'));
  window.feed = (s) => new Promise(r => term.write(Uint8Array.from(atob(s), c => c.charCodeAt(0)), r));
}, { COLS, ROWS, FS, W2R: E.win2?.rows ?? 0 });
const geo = await page.evaluate(() => {
  const w = document.getElementById('win').getBoundingClientRect();
  const s = document.querySelector('.xterm-screen').getBoundingClientRect(); window._r = !!document.querySelector('.xterm-screen canvas');
  return { ww: w.width, wh: w.height, sx: s.left - w.left, sy: s.top - w.top, cw: s.width / 124, ch: s.height / 44 };
});
console.log('geometry', geo);
// zoom 1: the whole window at E.wideFrac of the frame's height, centred
const base = E.wideFrac * H / geo.wh;

const ease = (u) => u < 0 ? 0 : u > 1 ? 1 : u * u * (3 - 2 * u);
function cam(v) {
  const K = E.camera; let a = K[0], b = K[0];
  for (let i = 0; i < K.length; i++) { if (K[i].t <= v) a = K[i]; }
  const ia = K.lastIndexOf(a); b = K[ia + 1] && !K[ia + 1].cut ? K[ia + 1] : a;
  if (b === a) return a;
  const u = ease((v - a.t) / (b.t - a.t));
  const lz = Math.log(a.zoom) + (Math.log(b.zoom) - Math.log(a.zoom)) * u;
  return { col: a.col + (b.col - a.col) * u, row: a.row + (b.row - a.row) * u, zoom: Math.exp(lz), free: a.free && b.free };
}
function xf(c) { // camera -> window transform
  const s = base * c.zoom;
  // the point to centre, in window px; zoom 1 centres the whole window
  const px = c.zoom <= 1.0001 ? geo.ww / 2 : geo.sx + c.col * geo.cw;
  const py = c.zoom <= 1.0001 ? geo.wh / 2 : geo.sy + c.row * geo.ch;
  const k = Math.min(1, Math.max(0, (c.zoom - 1) / 0.6)); // blend from window centre into the target
  const cx = geo.ww / 2 + ((geo.sx + c.col * geo.cw) - geo.ww / 2) * k;
  const cy = geo.wh / 2 + ((geo.sy + c.row * geo.ch) - geo.wh / 2) * k;
  let x = W / 2 - cx * s, y = H / 2 - cy * s;
  if (!c.free && geo.ww * s > W) x = Math.min(0, Math.max(W - geo.ww * s, x));
  if (!c.free && geo.wh * s > H) y = Math.min(0, Math.max(H - geo.wh * s, y));
  return { s, x, y };
}
const fade = (v, t0, t1, f = 0.35) => Math.min(ease((v - t0) / f), ease((t1 - v) / f));

let fed = 0, cur = null;
async function seek(a) {
  if (a.take !== cur) { cur = a.take; fed = 0; await page.evaluate(() => window.term.reset()); }
  const tty = TK[cur].tty, r = a.r;
  const chunks = [];
  while (fed < tty.length && tty[fed].t <= r) chunks.push(tty[fed++].b);
  for (const c of chunks) await page.evaluate((s) => window.feed(s), c);
}
function vid(tk, r) { // recording -> video seconds (for sounds)
  for (const s of segs) if (s.take === tk && r >= s.r0 && r < s.r1) return { v: s.v0 + (r - s.r0) * (s.v1 - s.v0) / (s.r1 - s.r0), speed: (s.r1 - s.r0) / (s.v1 - s.v0) };
  return null;
}
const keys = [], cues = [];
for (const tk of Object.keys(TK)) for (const m of TK[tk].marks) {
  if (m.kind === 'key') { const x = vid(tk, m.t); if (x && x.speed < 1.6) keys.push(+x.v.toFixed(3)); }
  if (m.kind === 'event' && (m.ev === 'needs_input' || m.ev === 'done')) {
    const x = vid(tk, m.at); if (x) cues.push({ kind: m.ev === 'done' ? 'done' : 'ask', v: +x.v.toFixed(3), r: m.at });
  }
}
for (let i = 0; i < E.end.url.length; i++) keys.push(+(urlAt + i / E.end.urlCPS).toFixed(3));
const timeline = { total, montageEnd, endAt, urlAt, urlEnd, keys, cues, scapes: E.scapes ?? [] };
const SP = E.split;
if (SP) await page.evaluate((SP) => {
  const d = document.getElementById('split');
  SP.panels.forEach((p, i) => {
    const x = SP.x0 + i * (SP.w + SP.gap);
    d.insertAdjacentHTML('beforeend', `<img id="sp${i}" style="position:absolute;left:${x}px;top:${SP.y}px;width:${SP.w}px;height:${SP.h}px;border-radius:10px;box-shadow:0 0 0 1px rgba(255,255,255,.09)"><div style="position:absolute;left:${x}px;width:${SP.w}px;top:${SP.y + SP.h + 22}px;text-align:center;font:700 ${SP.labelPx}px 'Geist Mono',monospace;color:#f2f2f2">${p.name}<span style="font-weight:400;color:#a9a9a9"> &middot; ${p.with}</span></div>`);
  });
}, SP);
const CD = E.card, W2 = E.win2, PL = E.plug;
const TK2 = W2 ? load(W2.take, 'tty.jsonl') : [];
let fed2 = 0;
async function seek2(r) { while (fed2 < TK2.length && TK2[fed2].t <= r) await page.evaluate((s) => window.feed2(s), TK2[fed2++].b); }
const w2h = W2 ? await page.evaluate(() => document.getElementById('win2').getBoundingClientRect().height) : 0;
const typingAt = (v) => keys.some(k => k <= v && v - k < 0.45);
const N = Math.round(total * FPS);
const want = stills ? stills.split(',').map(Number) : null;
const frames = want ? want.map(v => Math.round(v * FPS)) : [...Array(N).keys()].filter(i => i / FPS >= from && (!to || i / FPS < +to));
for (const i of frames) {
  const v = i / FPS;
  if (v < montageEnd + 0.01) await seek(at(v));
  const c = cam(Math.min(v, montageEnd)); const t = xf(c);
  const cap = E.captions.find(k => v >= k.t0 - 0.35 && v <= k.t1 + 0.35);
  const tag = E.tags.find(k => v >= k.t0 - 0.35 && v <= k.t1 + 0.35);
  const ef = Math.min(endFrames.length - 1, Math.max(0, Math.floor((v - endAt) * FPS)));
  const typed = Math.max(0, Math.min(E.end.url.length, Math.floor((v - urlAt) * E.end.urlCPS + 1e-6)));
  const blk = v < E.fromBlack ? 1 - ease(v / E.fromBlack)
    : v < endAt ? ease((v - montageEnd) / E.end.toBlack)
    : v > total - E.end.fade ? ease((v - (total - E.end.fade)) / E.end.fade) : 0;
  const cursorOn = typingAt(v) || Math.floor(v / 0.53) % 2 === 0;
  let w2 = null;
  if (W2 && v >= W2.v0 - 0.01 && v < W2.v1) {
    await seek2(W2.r0 + (v - W2.v0));
    const u = Math.max(0, Math.min(1, (v - W2.resizeAt) / W2.resizeDur)), e2 = ease(u);
    w2 = { h: w2h + (geo.wh - w2h) * e2, o: 1 - ease((u - 0.45) / 0.55) };
  }
  await page.evaluate((p) => {
    const w = document.getElementById('win');
    w.style.transform = `translate(${p.t.x}px,${p.t.y}px) scale(${p.t.s})`;
    document.documentElement.style.setProperty('--k', p.k);
    const w2e = document.getElementById('win2');
    if (p.w2) { w2e.style.transform = `translate(${p.t.x}px,${p.t.y}px) scale(${p.t.s})`; w2e.style.height = p.w2.h + 'px'; w2e.style.opacity = p.w2.o; } else w2e.style.opacity = 0;
    const pl = document.getElementById('plug'); pl.style.opacity = p.plO; if (p.plO > 0 && pl.dataset.h !== p.plH) { pl.innerHTML = p.plH; pl.dataset.h = p.plH; }
    document.getElementById('title').textContent = p.title;
    const cap = document.getElementById('cap');
    cap.style.opacity = p.cap ? p.capO : 0; if (p.cap) cap.firstChild.innerHTML = p.cap;
    const cd = document.getElementById('card'); cd.style.opacity = p.cardO;
    if (p.cardO > 0) { const ci = document.getElementById('cardimg'); if (ci.dataset.f !== p.cardF) { ci.src = p.cardF; ci.dataset.f = p.cardF; }
      const ct = document.getElementById('cardtxt'); if (ct.dataset.h !== p.cardH) { ct.innerHTML = p.cardH; ct.dataset.h = p.cardH; } ct.style.transform = `scale(${p.cardS})`; document.getElementById('cardimg').style.opacity = p.cardIO;
      const bg = ct.querySelector('.big'), sm = ct.querySelector('small'); if (bg) bg.style.opacity = p.cardBO; if (sm) sm.style.opacity = p.cardSO; }
    const tag = document.getElementById('tag');
    tag.style.opacity = p.tag ? p.tagO : 0; if (p.tag) tag.innerHTML = p.tag;
    const img = document.getElementById('endimg');
    if (p.end) { if (img.dataset.f !== p.ef) { img.src = p.ef; img.dataset.f = p.ef; } img.style.opacity = 1; }
    else img.style.opacity = 0;
    const u = document.getElementById('url'); u.style.opacity = p.end ? 1 : 0;
    u.textContent = p.url + (p.curOn ? '█' : ' ');
    const sp = document.getElementById('split');
    sp.style.opacity = p.splitO;
    if (p.splitO > 0) p.splitF.forEach((f, i) => { const im = document.getElementById('sp' + i); if (im.dataset.f !== f) { im.src = f; im.dataset.f = f; } });
    document.getElementById('black').style.opacity = p.blk;
    const cu = document.getElementById('cur');
    if (p.ownCursor) {
      const scr = document.querySelector('.xterm-screen');
      const cw = scr.clientWidth / 124, ch = scr.clientHeight / 44, b = window.term.buffer.active;
      cu.style.left = (b.cursorX * cw) + 'px'; cu.style.top = (b.cursorY * ch + ch * 0.1) + 'px';
      cu.style.width = cw + 'px'; cu.style.height = (ch * 0.8) + 'px'; cu.style.display = p.cursorOn ? 'block' : 'none';
    } else cu.style.display = 'none';
    if (window._cur !== p.cursorOn) { window._cur = p.cursorOn; window.term.options.theme = { background: '#0f0f10', foreground: '#d8d8d8', cursor: p.cursorOn ? '#d8d8d8' : '#0f0f10', cursorAccent: '#0f0f10' }; }
  }, { t, w2, plO: PL && v >= PL.v0 && v < endAt ? ease((v - PL.v0) / 0.5) : 0, plH: PL ? PL.html : '', k: E.chrome ? ease((v - E.chrome[0]) / (E.chrome[1] - E.chrome[0])) : 1, title: E.title, cap: cap ? (cap.html ?? cap.text) : null, capO: cap ? fade(v, cap.t0, cap.t1) : 0,
       cardO: CD && v >= CD.v0 && v < CD.v1 ? 1 : 0,
       cardIO: CD ? 1 - ease((v - CD.outA) / CD.fadeA) : 1, cardBO: CD ? ease((v - CD.v0 - CD.textIn) / 0.6) * (1 - ease((v - CD.outA) / CD.fadeA)) : 0,
       cardSO: CD ? ease((v - CD.v0 - CD.textIn - 0.25) / 0.6) * (1 - ease((v - CD.outB) / CD.fadeB)) : 0,
       cardF: CD ? 'file://' + path.join(here, CD.dir, 'f' + String(Math.min(CD.frames - 1, Math.max(0, Math.floor((v - CD.v0) * FPS)))).padStart(5, '0') + '.png') : '',
       cardH: CD ? CD.html : '', cardTO: CD ? ease((v - CD.v0 - CD.textIn) / 0.6) : 0, cardS: CD ? 0.97 + 0.03 * ease((v - CD.v0 - CD.textIn) / 1.6) : 1,
       tag: tag?.html, tagO: tag ? fade(v, tag.t0, tag.t1) : 0,
       end: v >= endAt, ef: 'file://' + path.join(here, 'endsrc', endFrames[ef]),
       url: E.end.url.slice(0, typed), curOn: v >= urlAt - 0.4 && (v < urlEnd + 0.3 || Math.floor(v / 0.5) % 2 === 0) && v >= endAt + 0.5,
       blk: v >= endAt && v < endAt + 0.0001 ? 1 : blk, cursorOn,
       splitO: SP && v >= SP.v0 && v < SP.v1 ? Math.min(ease((v - SP.v0) / 0.25), ease((SP.v1 - v) / 0.25)) : 0,
       splitF: SP ? SP.panels.map(pn => 'file://' + path.join(here, pn.dir, 'p' + String(Math.min(pn.frames - 1, Math.max(0, Math.floor((v - SP.v0) * FPS)))).padStart(5, '0') + '.png')) : [], ownCursor: v < (E.ownCursorUntil ?? 0) });
  await page.screenshot({ path: path.join(out, want ? `still-${v.toFixed(2)}.png` : `f${String(i).padStart(5, '0')}.png`) });
  if (i % 60 === 0) process.stdout.write(`${v.toFixed(1)} `);
}
fs.writeFileSync(path.join(out, 'timeline.json'), JSON.stringify(timeline, null, 1));
await browser.close(); console.log('\nframes', frames.length, 'total', total.toFixed(2));
