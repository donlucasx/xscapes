// render-card.mjs -- the QT card's stage. The site's "Scapes and companions"
// section as a 16:9 loop: the section's own header rule, the four cast
// portraits, their captions. The browser is the compositor (as the trailer's
// render.mjs): every frame is the page's markup for that instant,
// screenshotted at 1920x1080 by the installed Chrome. Glyphs are vector.
//
//   node render-card.mjs <framesDir> <outDir> [--stills t,t,..]
//
// <framesDir>/cast.js is written by TestQTExport (qt_test.go):
//   window.TRAILER.cast = {fps, css, crab:{cols,rows,frames}, cat:…, owl:…, frog:…}
import { chromium } from 'playwright';
import fs from 'node:fs';
import path from 'node:path';

const here = path.dirname(new URL(import.meta.url).pathname);
const [framesDir, outDir, ...rest] = process.argv.slice(2);
if (!framesDir || !outDir) {
  console.error('usage: node render-card.mjs <framesDir> <outDir> [--stills t,t,..]');
  process.exit(2);
}
const stillsOnly = rest.includes('--stills');
const stillsAt = stillsOnly ? rest[rest.indexOf('--stills') + 1] : null;
fs.mkdirSync(outDir, { recursive: true });

const card = JSON.parse(fs.readFileSync(path.join(here, 'card.json'), 'utf8'));
const fontsCSS = fs.readFileSync(path.join(here, 'fonts', 'fonts-local.css'), 'utf8')
  .replace(/url\(([^)]+\.woff2)\)/g, (_, f) => `url(${path.join(here, 'fonts', f)})`);
const unwrap = (js, name) => JSON.parse(js.slice(js.indexOf(name + '=') + name.length + 1).replace(/;\s*$/, ''));
const cast = unwrap(fs.readFileSync(path.join(framesDir, 'cast.js'), 'utf8'), 'window.TRAILER.cast');

const { W, H } = card;
const esc = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;');

// The loop's length: the animals' own loops, all of them closing together.
const lens = card.cast.map((c) => cast[c.key].frames.length);
const gcd = (a, b) => (b ? gcd(b, a % b) : a);
const lcm = (a, b) => (a * b) / gcd(a, b);
const nFrames = lens.reduce(lcm);
const fps = cast.fps;
console.log('animals:', card.cast.map((c) => `${c.key} ${cast[c.key].cols}x${cast[c.key].rows} ${cast[c.key].frames.length}f`).join(' · '), `→ loop ${nFrames}f = ${(nFrames / fps).toFixed(2)}s at ${fps}fps`);

// Frames on the page: one array per animal, the page indexes them.
const framesJS = 'window.CAST=' + JSON.stringify(Object.fromEntries(card.cast.map((c) => [c.key, cast[c.key].frames]))) + ';';
fs.writeFileSync(path.join(outDir, 'cast-frames.js'), framesJS);

const cols = card.cast.map((c) => `
<div class="one">
  <div class="port"><div class="fx" id="fx-${c.key}"></div></div>
  <p class="cap"><b>${esc(c.name)}</b> ${esc(c.line)}</p>
</div>`).join('');

const stage = `<!doctype html><meta charset="utf-8"><title>card</title>
<style>
${fontsCSS}
${cast.css}
:root{--g:#121212;--d:#8a8a8a;--dd:#bcbcbc;--i:#eeeeee;--eye:#a8ecb0;
--mono:"JetBrains Mono",ui-monospace,"SF Mono",Menlo,monospace}
html,body{margin:0;width:${W}px;height:${H}px;overflow:hidden;background:var(--g)}
#stage{position:relative;width:${W}px;height:${H}px;box-sizing:border-box;
  padding:${card.pad.top}px ${card.pad.side}px ${card.pad.bottom}px;display:flex;flex-direction:column}
.boxtop{font-family:var(--mono);font-weight:400;font-size:${card.rule.px}px;line-height:1.15;color:var(--d);white-space:pre;overflow:hidden}
.boxtop b{color:var(--i);font-weight:700;letter-spacing:.08em;text-transform:uppercase}
.boxtop i{color:var(--d);font-style:normal}
.cast{flex:1;display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:${card.port.gap}px;align-content:center}
.one{min-width:0}
.port{height:${card.port.h}px;display:flex;align-items:center;justify-content:center;overflow:visible}
.fx pre{margin:0;font-family:Menlo,"SF Mono","DejaVu Sans Mono",ui-monospace,monospace;font-size:${card.cell.px}px!important;line-height:1;letter-spacing:0;white-space:pre;display:block}
.cap{font-family:var(--mono);font-weight:400;font-size:${card.cap.px}px;line-height:${card.cap.lh};color:var(--d);text-align:center;margin:${card.cap.top}px 0 0}
.cap b{color:var(--i);font-weight:500}
</style>
<div id="stage">
  <div class="boxtop" id="rule"></div>
  <div class="cast">${cols}</div>
</div>
<script src="cast-frames.js"></script>
<script>
const CARD = ${JSON.stringify({ title: card.title, right: card.right || '' })};
// The section header, drawn exactly as the site's drawBox draws a .sec: a
// rule of box-drawing characters filling the width in whole characters,
// the title bold and uppercase after the corner.
function chWidth(el){
  const s = document.createElement('span'); s.textContent = '─'.repeat(50);
  s.style.cssText = 'position:absolute;visibility:hidden;white-space:pre;font:inherit';
  el.appendChild(s); const w = s.getBoundingClientRect().width / 50; s.remove(); return w;
}
function drawRule(){
  const top = document.getElementById('rule');
  const n = Math.max(14, Math.floor(top.clientWidth / chWidth(top)));
  const title = CARD.title.trim(), right = CARD.right.trim();
  const head = title ? '┌─ ' + title + ' ' : '┌─';
  const tail = right ? ' ' + right + ' ─┐' : '┐';
  let fill = n - head.length - tail.length;
  const dash = '─'.repeat(Math.max(1, fill));
  const escp = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;');
  let html = title ? '┌─ <b>' + escp(title) + '</b> ' + dash : '┌─' + dash;
  html += right ? ' <i>' + escp(right) + '</i> ─┐' : '┐';
  top.innerHTML = html;
}
const FPS = ${fps};
const shown = {};
window.setFrame = function (t) {
  const k = Math.floor(t * FPS + 1e-6);
  for (const key of Object.keys(window.CAST)) {
    const fr = window.CAST[key];
    const i = ((k % fr.length) + fr.length) % fr.length;
    if (shown[key] !== i) { document.getElementById('fx-' + key).innerHTML = fr[i]; shown[key] = i; }
  }
};
window.measure = function () {
  const r = (id) => { const b = document.getElementById(id).getBoundingClientRect(); return [Math.round(b.left), Math.round(b.top), Math.round(b.width), Math.round(b.height)]; };
  return { rule: r('rule'), crab: r('fx-crab'), cat: r('fx-cat'), owl: r('fx-owl'), frog: r('fx-frog') };
};
</script>`;

const stagePath = path.resolve(outDir, 'stage.html');
fs.writeFileSync(stagePath, stage);

const browser = await chromium.launch({ channel: 'chrome' });
const page = await browser.newPage({ viewport: { width: W, height: H }, deviceScaleFactor: 1 });
await page.goto('file://' + stagePath);
await page.evaluate(async (px) => {
  await Promise.all([
    document.fonts.load(`700 ${px.rule}px "JetBrains Mono"`),
    document.fonts.load(`400 ${px.rule}px "JetBrains Mono"`),
    document.fonts.load(`500 ${px.cap}px "JetBrains Mono"`),
  ]);
  await document.fonts.ready;
  drawRule();
}, { rule: card.rule.px, cap: card.cap.px });
const loaded = await page.evaluate(() => [...document.fonts].filter(f => f.status === 'loaded').map(f => f.family + ' ' + f.weight));
console.log('fonts loaded:', [...new Set(loaded)].join(', '));
await page.evaluate(() => window.setFrame(0));
console.log('layout:', JSON.stringify(await page.evaluate(() => window.measure())));

const shot = async (t, file) => {
  await page.evaluate((t) => window.setFrame(t), t);
  await page.screenshot({ path: file, type: 'png', clip: { x: 0, y: 0, width: W, height: H }, animations: 'disabled' });
};

if (stillsOnly) {
  const stills = stillsAt && /^[\d.,]+$/.test(stillsAt) ? stillsAt.split(',').map(Number) : [0, 0.5, 1.5, 3.0];
  for (const t of stills) await shot(t, path.join(outDir, `still-${t.toFixed(2)}.png`));
  console.log('stills:', stills.length);
} else {
  const t0 = Date.now();
  for (let f = 0; f < nFrames; f++) {
    await shot(f / fps, path.join(outDir, `f${String(f).padStart(5, '0')}.png`));
  }
  console.log(`frames: ${nFrames} in ${((Date.now() - t0) / 1000).toFixed(1)}s`);
}
await browser.close();
