// render.mjs -- the trailer's stage. The browser is the compositor: every
// video frame is the page's own markup under styles the script sets for that
// instant, screenshotted at 1920x1080. Glyphs are vector, so any zoom is
// crisp. Deterministic: no CSS animations, every state is a function of t.
//
//   node render.mjs opener <framesDir> <outDir> [--stills]        the title card alone
//   node render.mjs film   <framesDir> <outDir> [--stills t,t,..]  the whole film
//
// film writes <outDir>/film.json: the resolved timeline in VIDEO seconds
// (where the montage starts, every cue, the cut, the end), which audio.py
// scores from.
import { chromium } from 'playwright';
import fs from 'node:fs';
import path from 'node:path';

const here = path.dirname(new URL(import.meta.url).pathname);
const [mode, framesDir, outDir, ...rest] = process.argv.slice(2);
if (!['opener', 'film', 'clip'].includes(mode) || !framesDir || !outDir) {
  console.error('usage: node render.mjs opener|film <framesDir> <outDir> [--stills [t,t,..]]  |  clip <framesDir> <outDir> --name <n>');
  process.exit(2);
}
const stillsAt = rest.includes('--stills') ? rest[rest.indexOf('--stills') + 1] : null;
const stillsOnly = rest.includes('--stills');
fs.mkdirSync(outDir, { recursive: true });

const opener = JSON.parse(fs.readFileSync(path.join(here, 'opener.json'), 'utf8'));
const shots = JSON.parse(fs.readFileSync(path.join(here, 'shots.json'), 'utf8'));
const fontsCSS = fs.readFileSync(path.join(here, 'fonts', 'fonts-local.css'), 'utf8')
  .replace(/url\(([^)]+\.woff2)\)/g, (_, f) => `url(${path.join(here, 'fonts', f)})`);
const unwrap = (js, name) => JSON.parse(js.slice(js.indexOf(name + '=') + name.length + 1).replace(/;\s*$/, ''));
const cover = unwrap(fs.readFileSync(path.join(framesDir, 'cover.js'), 'utf8'), 'window.TRAILER.cover');

const { W, H } = opener;
const esc = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;');

// ---- the opener: arrivals ---------------------------------------------------
function rng(seed) {
  let s = seed >>> 0 || 1;
  return () => { s = (s * 1664525 + 1013904223) >>> 0; return s / 4294967296; };
}
const key = (x, y) => y * cover.cols + x;
const arrival = new Map();
{
  // the stars: every sky cell, in a shuffled order, one at a time, the
  // first ones slow and the rest quickening (an ease-in on the count)
  const sky = new Map();
  for (const f of cover.frames) for (const [x, y] of f) if (y < cover.horizon) sky.set(key(x, y), [x, y]);
  const cells = [...sky.values()];
  const r = rng(opener.stars.seed);
  for (let i = cells.length - 1; i > 0; i--) { const j = Math.floor(r() * (i + 1)); [cells[i], cells[j]] = [cells[j], cells[i]]; }
  cells.forEach(([x, y], i) => {
    const u = Math.pow(i / Math.max(1, cells.length - 1), 1 / opener.stars.ease);
    arrival.set(key(x, y), opener.stars.start + u * opener.stars.dur);
  });
  // the wipe: the field below the horizon in bands, one band per step of
  // the sea's own cycle, each band swept left to right inside its step
  const bw = Math.ceil(cover.cols / opener.wipe.bands);
  const bandEvery = opener.wipe.band || opener.sea.step; // v4: quicker than the sea's own step
  const r2 = rng(opener.wipe.seed);
  for (const f of cover.frames) for (const [x, y] of f) {
    if (y < cover.horizon) continue;
    const k = key(x, y);
    if (arrival.has(k)) continue;
    const band = Math.floor(x / bw);
    arrival.set(k, opener.wipe.start + band * bandEvery + ((x - band * bw) / bw) * opener.wipe.sweep + r2() * opener.wipe.jitter);
  }
}
const seaPx = opener.sea.px;
const seaW = Math.round(cover.cols * seaPx * 0.6);
const seaH = cover.rows * seaPx;

function seaLayerHTML(id, fi, coloured) {
  const rows = Array.from({ length: cover.rows }, () => []);
  for (const [x, y, ch, hex] of cover.frames[fi]) rows[y].push([x, ch, hex]);
  const lines = rows.map((cells, y) => {
    cells.sort((a, b) => a[0] - b[0]);
    let out = '', at = 0;
    for (const [x, ch, hex] of cells) {
      out += ' '.repeat(x - at);
      out += coloured
        ? `<span class="c" data-k="${key(x, y)}" style="color:#${hex}" data-c="#${hex}">${esc(ch)}</span>`
        : `<span class="c" data-k="${key(x, y)}">${esc(ch)}</span>`;
      at = x + 1;
    }
    return out;
  });
  return `<pre class="${id}" data-f="${fi}">${lines.join('\n')}</pre>`;
}
const seaLayers = cover.frames.map((_, fi) => seaLayerHTML('grey', fi, false) + seaLayerHTML('col', fi, true)).join('');

const openerCSS = `
#opener{position:absolute;inset:0;background:#121212;overflow:hidden}
#sea{position:absolute;left:${Math.round((W - seaW) / 2)}px;top:${Math.round((H - seaH) / 2)}px;width:${seaW}px;height:${seaH}px}
#sea pre{position:absolute;left:0;top:0;margin:0;font:700 ${seaPx}px/1 "${opener.sea.font}",monospace;letter-spacing:0;white-space:pre;display:none}
#sea pre.on{display:block}
#sea pre.grey{color:${opener.sea.grey}}
#sea .c{visibility:hidden}
#sea .c.in{visibility:visible}
#title,#endcard{position:absolute;inset:0;display:flex;flex-direction:column;align-items:center;justify-content:center}
.lockup{font:700 ${opener.cell.px}px/1.2 "Geist Mono",monospace;letter-spacing:0;white-space:nowrap;color:#eeeeee;visibility:hidden}
.lockup.on{visibility:visible}
.blk{display:inline-block;width:1ch;line-height:1.2em;text-align:center;vertical-align:top;background:#eeeeee;color:#121212;box-sizing:border-box}
.blk.hollow{background:transparent;color:#eeeeee;box-shadow:inset 0 0 0 ${Math.max(2, Math.round(opener.cell.px / 40))}px #eeeeee}
.blk.nox{color:transparent!important}
.w span{visibility:hidden}
.w span.on{visibility:visible}
.line{font:400 ${opener.line.px}px/1.2 "Geist Mono",monospace;color:#bcbcbc;margin-top:${Math.round(opener.line.px * 0.4)}px;letter-spacing:.02em;opacity:0}
`;
const openerHTML = `<div id="opener">
  <div id="sea">${seaLayers}</div>
  <div id="title">
    <div class="lockup" id="lockup"><span class="blk nox" id="blk">x</span><span class="w" id="w">${'scapes'.split('').map(ch => `<span>${ch}</span>`).join('')}</span></div>
    <div class="line" id="line">${esc(opener.line.text)}</div>
  </div>
  <div id="openerBlack" style="position:absolute;inset:0;background:#121212;opacity:0"></div>
</div>`;
// setOpener(t): the opener's whole state at its own second t.
const openerJS = `
const O = ${JSON.stringify(opener)};
const HORIZON = ${cover.horizon}, COLS = ${cover.cols};
const ARR = new Map(${JSON.stringify([...arrival.entries()])});
const clamp = (u) => Math.max(0, Math.min(1, u));
const smooth = (u) => { u = clamp(u); return u * u * (3 - 2 * u); };
const pres = [...document.querySelectorAll('#sea pre')];
const cellsOf = new Map(pres.map((p) => [p, [...p.querySelectorAll('.c')]]));
const hex2 = (h) => [parseInt(h.slice(1, 3), 16), parseInt(h.slice(3, 5), 16), parseInt(h.slice(5, 7), 16)];
function setOpener(t) {
  const idx = t < O.wipe.start ? 0 : Math.floor((t - O.wipe.start) / O.sea.step) % 6;
  for (const p of pres) p.classList.toggle('on', +p.dataset.f === idx);
  for (const p of pres) {
    if (+p.dataset.f !== idx) continue;
    const coloured = p.classList.contains('col');
    for (const c of cellsOf.get(p)) {
      const a = ARR.get(+c.dataset.k);
      const inNow = a !== undefined && t >= a;
      c.classList.toggle('in', inNow);
      if (!coloured) continue;
      const y = Math.floor(+c.dataset.k / COLS);
      const dur = y < HORIZON ? O.stars.shine : O.wipe.shine;
      let s = inNow ? 1 - (t - a) / dur : 0;
      if (s > 0) {
        s = Math.pow(clamp(s), 1.4);
        const [r, g, b] = hex2(c.dataset.c);
        const L = O.shine.lift * s;
        c.style.color = 'rgb(' + Math.round(r + (255 - r) * L) + ',' + Math.round(g + (250 - g) * L) + ',' + Math.round(b + (236 - b) * L) + ')';
        c.style.textShadow = '0 0 ' + (O.shine.halo * s).toFixed(1) + 'px rgba(255,248,230,' + s.toFixed(2) + '),0 0 ' + (O.shine.glow * s).toFixed(1) + 'px rgba(255,236,200,' + (0.75 * s).toFixed(2) + ')';
      } else if (c.style.textShadow) {
        c.style.color = c.dataset.c;
        c.style.textShadow = '';
      }
    }
  }
  const drain = smooth((t - O.drain.start) / O.drain.dur);
  for (const p of pres) if (p.classList.contains('col')) p.style.opacity = 1 - drain;
  const lockup = document.getElementById('lockup');
  const blk = document.getElementById('blk');
  lockup.classList.toggle('on', t >= O.cell.start);
  const phase = ((t - O.cell.start) % O.cell.period) / O.cell.period;
  const on = t >= O.cell.settle || phase < 0.5;
  blk.classList.toggle('hollow', !on);
  blk.classList.toggle('nox', t < O.x.at);
  let at = O.letters.start;
  const spans = document.querySelectorAll('#w span');
  O.letters.gaps.forEach((g, i) => { at += g; spans[i].classList.toggle('on', t >= at); });
  const line = document.getElementById('line');
  const v = smooth((t - O.line.start) / O.line.dur);
  line.style.opacity = v;
  line.style.filter = 'blur(' + ((1 - v) * 10).toFixed(2) + 'px)';
  // the opener's own fade to black (v4: black, THEN the terminal types)
  document.getElementById('openerBlack').style.opacity = smooth((t - O.out.start) / O.out.dur);
}
`;

// ---- the film ----------------------------------------------------------------
let stage, timeline;
if (mode === 'clip') {
  // A montage alone, wide, every frame: for a GIF. The frames file carries
  // its own palette (window.TRAILER.<name> = {cols, rows, fps, frames, css}).
  const name = rest[rest.indexOf('--name') + 1];
  const clip = unwrap(fs.readFileSync(path.join(framesDir, name + '.js'), 'utf8'), 'window.TRAILER.' + name);
  const S = H / clip.rows;
  fs.writeFileSync(path.join(outDir, 'clip-frames.js'), 'window.CLIP=' + JSON.stringify(clip.frames) + ';');
  stage = `<!doctype html><meta charset="utf-8"><title>clip</title>
<style>${clip.css}
html,body{margin:0;width:${W}px;height:${H}px;overflow:hidden;background:#121212}
#stage{position:relative;width:${W}px;height:${H}px;overflow:hidden;display:flex;justify-content:center}
#win pre{margin:0;font-family:Menlo,"SF Mono","DejaVu Sans Mono",ui-monospace,monospace;font-size:${S}px!important;line-height:1;letter-spacing:0;white-space:pre;display:block}
</style>
<div id="stage"><div id="win"></div></div>
<script src="clip-frames.js"></script>
<script>
let shown = -1;
window.setFrame = function (t) {
  const k = Math.max(0, Math.min(window.CLIP.length - 1, Math.floor(t * ${clip.fps} + 1e-6)));
  if (k !== shown) { document.getElementById('win').innerHTML = window.CLIP[k]; shown = k; }
};
</script>`;
  timeline = { end: clip.frames.length / clip.fps, fps: clip.fps };
} else if (mode === 'opener') {
  stage = `<!doctype html><meta charset="utf-8"><title>opener</title>
<style>${fontsCSS}
html,body{margin:0;width:${W}px;height:${H}px;overflow:hidden;background:#121212}
#stage{position:relative;width:${W}px;height:${H}px;overflow:hidden}
${openerCSS}
#black{position:absolute;inset:0;background:#121212;opacity:0}
</style>
<div id="stage">${openerHTML}<div id="black"></div></div>
<script>${openerJS}
window.setFrame = function (t) {
  setOpener(t);
};
</script>`;
  timeline = { end: opener.end };
} else {
  const hero = unwrap(fs.readFileSync(path.join(framesDir, 'hero.js'), 'utf8'), 'window.TRAILER.hero');
  const paletteCSS = fs.readFileSync(path.join(framesDir, 'palette.css'), 'utf8');
  // the cat for the end card: its own frames and its own palette, scoped
  // under #cat so its classes cannot collide with the hero's
  const cat = unwrap(fs.readFileSync(path.join(framesDir, 'cat.js'), 'utf8'), 'window.TRAILER.cat');
  const catCSS = cat.css.replace(/\.p(\d+)\{/g, '#cat .p$1{');
  const M0 = opener.end;                       // the montage's t = 0: after the opener's black
  const clipEnd = hero.frames.length / hero.fps;
  const endAt = M0 + clipEnd;                  // the montage fades to black from here
  const E = shots.end;
  const blackAt = endAt + E.toBlack;           // fully black; the logo starts fading in
  const logoInAt = blackAt + E.logoIn;
  const urlAt = logoInAt + E.urlAfter;
  const urlEnd = urlAt + E.url.length / E.urlCPS;
  const catAt = urlEnd + E.catAfter;           // the cat starts fading in; its clock starts here
  const catDoneAt = catAt + cat.workFrames / cat.fps;
  const fadeAt = catDoneAt + E.hold;
  const end = fadeAt + E.fade;
  timeline = {
    fps: opener.fps, montageStart: M0, montageEnd: endAt, cut: hero.cuts.map((t) => M0 + t),
    asks: hero.asks.map((t) => M0 + t), dones: hero.dones.map((t) => M0 + t),
    keys: (() => { let at = opener.letters.start; return opener.letters.gaps.map((g) => (at += g)); })(),
    blackAt, logoInAt, urlAt, urlEnd, catAt, catDoneAt, fadeAt, end,
    endKeys: [...E.url].map((_, i) => urlAt + i / E.urlCPS),
  };
  const S = H / hero.rows;                     // the cell's height: the window fits the frame's height at zoom 1
  const framesJS = 'window.HERO=' + JSON.stringify(hero.frames) + ';';
  fs.writeFileSync(path.join(outDir, 'hero-frames.js'), framesJS);
  stage = `<!doctype html><meta charset="utf-8"><title>film</title>
<style>${fontsCSS}
${paletteCSS}
html,body{margin:0;width:${W}px;height:${H}px;overflow:hidden;background:#121212}
#stage{position:relative;width:${W}px;height:${H}px;overflow:hidden}
#win{position:absolute;left:0;top:0;transform-origin:0 0}
#win pre{margin:0;font-family:Menlo,"SF Mono","DejaVu Sans Mono",ui-monospace,monospace;font-size:${S}px!important;line-height:1;letter-spacing:0;white-space:pre;display:block}
#cap{position:absolute;left:0;right:0;bottom:108px;display:flex;justify-content:center;pointer-events:none}
#cap.top{bottom:auto;top:56px}
#cap span{font:400 ${shots.captionPx}px/1.4 "Geist Mono",monospace;color:#eeeeee;background:#1c1c1c;padding:4px 16px;border-radius:2px;opacity:0;letter-spacing:.01em}
${openerCSS}
#endcard{background:#121212;opacity:0}
#endcard .lockup{visibility:visible}
#endcard .url{font:400 ${shots.end.urlPx}px/1.2 "Geist Mono",monospace;color:#8a8a8a;margin-top:${Math.round(shots.end.urlPx * 0.6)}px;letter-spacing:.02em;white-space:pre}
#endcard .url span{visibility:hidden}
#endcard .url span.on{visibility:visible}
#endcard .url{position:relative}
#endcard .url .cur{position:absolute;left:calc(var(--n,0) * 1ch);top:.1em;width:.6em;height:1em;background:#8a8a8a;visibility:hidden}
#endcard .url .cur.on{visibility:visible}
#cat{margin-top:${shots.end.cat.gap}px;height:${cat.rows * shots.end.cat.px}px;opacity:0}
#blackUnder{position:absolute;inset:0;background:#121212;opacity:0}
#cat pre{margin:0;font-family:Menlo,"SF Mono","DejaVu Sans Mono",ui-monospace,monospace;font-size:${shots.end.cat.px}px!important;line-height:1;letter-spacing:0;white-space:pre;display:block}
${catCSS}
#black{position:absolute;inset:0;background:#121212;opacity:0}
</style>
<div id="stage">
  <div id="win"></div>
  <div id="cap"><span id="capt"></span></div>
  <div id="blackUnder"></div>
  ${openerHTML}
  <div id="endcard"><div class="lockup"><span class="blk" id="eblk">x</span>scapes</div><div class="url" id="eurl">${[...shots.end.url].map(ch => `<span>${esc(ch)}</span>`).join('')}<span class="cur" id="ecur"></span></div><div id="cat"></div></div>
  <div id="black"></div>
</div>
<script src="hero-frames.js"></script>
<script>window.CAT=${JSON.stringify(cat.frames)};</script>
<script>${openerJS}
const SH = ${JSON.stringify(shots)};
const TL = ${JSON.stringify(timeline)};
const FPS = ${hero.fps}, ROWS = ${hero.rows}, COLSW = ${hero.cols}, S = ${S};
const win = document.getElementById('win');
let shown = -1, cellW = 0, winW = 0, winH = ROWS * S;
function measure() {
  win.innerHTML = window.HERO[0];
  const pre = win.querySelector('pre');
  winW = pre.getBoundingClientRect().width;
  cellW = winW / COLSW;
  shown = 0;
  return { winW, cellW, winH };
}
window.measure = measure;
function cameraAt(t) {
  const k = SH.camera;
  if (t <= k[0].t) return k[0];
  for (let i = 1; i < k.length; i++) {
    if (t <= k[i].t) {
      if (k[i].cut) return t < k[i].t ? k[i - 1] : k[i];
      const u = smooth((t - k[i - 1].t) / (k[i].t - k[i - 1].t));
      const a = k[i - 1], b = k[i];
      return { col: a.col + (b.col - a.col) * u, row: a.row + (b.row - a.row) * u, zoom: a.zoom + (b.zoom - a.zoom) * u,
        free: (a.free || 0) + ((b.free || 0) - (a.free || 0)) * u };
    }
  }
  return k[k.length - 1];
}
function place(cam) {
  const z = cam.zoom;
  const px = cam.col * cellW, py = cam.row * S;
  // free: the subject at the frame's centre wherever it is; clamped: the
  // frame kept inside the window (centred when the window is smaller)
  const fx = ${W} / 2 - px * z, fy = ${H} / 2 - py * z;
  const w = winW * z, h = winH * z;
  const cx = w >= ${W} ? Math.min(0, Math.max(${W} - w, fx)) : (${W} - w) / 2;
  const cy = h >= ${H} ? Math.min(0, Math.max(${H} - h, fy)) : (${H} - h) / 2;
  const f = cam.free || 0;
  const tx = fx * f + cx * (1 - f), ty = fy * f + cy * (1 - f);
  win.style.transform = 'translate(' + tx.toFixed(2) + 'px,' + ty.toFixed(2) + 'px) scale(' + z.toFixed(4) + ')';
}
window.setFrame = function (T) {
  if (!cellW) measure();
  // the montage under everything: its frame at clip time, the camera over it
  const t = T - TL.montageStart;
  const k = Math.max(0, Math.min(window.HERO.length - 1, Math.floor(t * FPS + 1e-6)));
  if (k !== shown) { win.innerHTML = window.HERO[k]; shown = k; }
  place(cameraAt(Math.max(0, t)));
  // captions
  const capt = document.getElementById('capt');
  let cap = null;
  for (const c of SH.captions) if (t >= c.t0 - 0.25 && t <= c.t1 + 0.25) cap = c;
  if (cap) {
    capt.textContent = cap.text;
    document.getElementById('cap').classList.toggle('top', cap.pos === 'top');
    capt.style.transform = 'translateX(' + (cap.dx || 0) + 'px)';
    capt.style.opacity = Math.min(smooth((t - (cap.t0 - 0.25)) / 0.25), smooth(((cap.t1 + 0.25) - t) / 0.25));
  } else capt.style.opacity = 0;
  // the opener over it, fading to its own black; gone once the montage starts
  const op = document.getElementById('opener');
  if (T < O.end) {
    op.style.display = '';
    setOpener(T);
  } else op.style.display = 'none';
  // the montage to black, then the end card in his beats: the logo fades in
  // with the cursor blinking, the URL types under it, the cat appears and
  // emotes (its finished face holds), then black
  const E = SH.end;
  document.getElementById('blackUnder').style.opacity = smooth((T - TL.montageEnd) / E.toBlack);
  document.getElementById('endcard').style.opacity = smooth((T - TL.blackAt) / E.logoIn);
  const eblk = document.getElementById('eblk');
  const blinking = T < TL.catAt;
  const ephase = ((T - TL.blackAt) % E.blink) / E.blink;
  eblk.classList.toggle('hollow', blinking && ephase >= 0.5);
  const typed = Math.max(0, Math.min(E.url.length, Math.floor((T - TL.urlAt) * E.urlCPS + 1e-6)));
  document.querySelectorAll('#eurl span:not(.cur)').forEach((s, i) => s.classList.toggle('on', i < typed));
  const ecur = document.getElementById('ecur');
  ecur.classList.toggle('on', T >= TL.urlAt - 0.3 && T < TL.urlEnd + 0.4);
  ecur.style.setProperty('--n', typed);
  const catEl = document.getElementById('cat');
  catEl.style.opacity = smooth((T - TL.catAt) / E.catIn);
  const ck = Math.max(0, Math.min(window.CAT.length - 1, Math.floor((T - TL.catAt) * ${cat.fps} + 1e-6)));
  if (catEl.dataset.k !== String(ck)) { catEl.innerHTML = window.CAT[ck]; catEl.dataset.k = String(ck); }
  document.getElementById('black').style.opacity = smooth((T - TL.fadeAt) / SH.end.fade);
};
</script>`;
  fs.writeFileSync(path.join(outDir, 'film.json'), JSON.stringify(timeline, null, 2));
  console.log('timeline:', JSON.stringify(timeline));
}

const stagePath = path.resolve(outDir, 'stage.html');
fs.writeFileSync(stagePath, stage);

const browser = await chromium.launch({ channel: 'chrome' });
const page = await browser.newPage({ viewport: { width: W, height: H }, deviceScaleFactor: 1 });
await page.goto('file://' + stagePath);
await page.evaluate(async (px) => {
  await Promise.all([
    document.fonts.load(`700 ${px.cell}px "Geist Mono"`),
    document.fonts.load(`400 ${px.line}px "Geist Mono"`),
    document.fonts.load(`700 ${px.sea}px "JetBrains Mono"`),
  ]);
  await document.fonts.ready;
}, { cell: opener.cell.px, line: opener.line.px, sea: seaPx });
const loaded = await page.evaluate(() => [...document.fonts].filter(f => f.status === 'loaded').map(f => f.family + ' ' + f.weight));
console.log('fonts loaded:', [...new Set(loaded)].join(', '));
if (mode === 'film') console.log('window:', JSON.stringify(await page.evaluate(() => window.measure())));

const shot = async (t, file) => {
  await page.evaluate((t) => window.setFrame(t), t);
  await page.screenshot({ path: file, type: 'png', clip: { x: 0, y: 0, width: W, height: H }, animations: 'disabled' });
};

if (stillsOnly) {
  const stills = stillsAt && /^[\d.,]+$/.test(stillsAt) ? stillsAt.split(',').map(Number)
    : mode === 'opener' ? [0.4, 1.0, 1.6, 2.3, 3.2, 4.2, 4.95, 5.5, 6.4, 7.6]
      : [9.6, 12, 14, 17, 20.5, 23, 25.5, 28, 31, 33];
  for (const t of stills) await shot(t, path.join(outDir, `still-${t.toFixed(2)}.png`));
  console.log('stills:', stills.length);
} else {
  const fps = timeline.fps || opener.fps;
  const n = Math.round(timeline.end * fps);
  const t0 = Date.now();
  for (let f = 0; f < n; f++) {
    await shot(f / fps, path.join(outDir, `f${String(f).padStart(5, '0')}.png`));
  }
  console.log(`frames: ${n} in ${((Date.now() - t0) / 1000).toFixed(1)}s`);
}
await browser.close();
