// render-fire.mjs -- the fire clip's stage: the vista alone under a camera
// (the trailer's render.mjs film camera, nothing else). The browser is the
// compositor: every frame is the page's markup for that instant under a CSS
// transform, screenshotted at 1920x1080 by the installed Chrome. Glyphs are
// vector, so the closeup is crisp.
//
//   node render-fire.mjs <framesDir> <outDir> [--stills t,t,..]
//
// <framesDir>/fire.js is written by TestFireExport (fire_test.go):
//   window.TRAILER.fire = {cols, rows, fps, secs, frames, css}
// fire.json beside this script holds the camera (keyframes in clip seconds).
import { chromium } from 'playwright';
import fs from 'node:fs';
import path from 'node:path';

const here = path.dirname(new URL(import.meta.url).pathname);
const [framesDir, outDir, ...rest] = process.argv.slice(2);
if (!framesDir || !outDir) {
  console.error('usage: node render-fire.mjs <framesDir> <outDir> [--name fire|shore] [--stills t,t,..]');
  process.exit(2);
}
// --name picks the clip: <framesDir>/<name>.js (window.TRAILER.<name>) and
// <name>.json beside this script for the camera. The default is the fire.
const name = rest.includes('--name') ? rest[rest.indexOf('--name') + 1] : 'fire';
const stillsOnly = rest.includes('--stills');
const stillsAt = stillsOnly ? rest[rest.indexOf('--stills') + 1] : null;
fs.mkdirSync(outDir, { recursive: true });

const cam = JSON.parse(fs.readFileSync(path.join(here, name + '.json'), 'utf8'));
const unwrap = (js, name) => JSON.parse(js.slice(js.indexOf(name + '=') + name.length + 1).replace(/;\s*$/, ''));
const clip = unwrap(fs.readFileSync(path.join(framesDir, name + '.js'), 'utf8'), 'window.TRAILER.' + name);
const { W, H } = cam;
const S = H / clip.rows;
const fps = clip.fps;
const nFrames = clip.frames.length;
console.log(`fire: ${clip.cols}x${clip.rows}, ${nFrames} frames at ${fps} fps = ${(nFrames / fps).toFixed(2)} s; cell ${S.toFixed(2)} px`);

fs.writeFileSync(path.join(outDir, 'fire-frames.js'), 'window.FIRE=' + JSON.stringify(clip.frames) + ';');

const stage = `<!doctype html><meta charset="utf-8"><title>fire</title>
<style>
${clip.css}
html,body{margin:0;width:${W}px;height:${H}px;overflow:hidden;background:#121212}
#stage{position:relative;width:${W}px;height:${H}px;overflow:hidden}
#win{position:absolute;left:0;top:0;transform-origin:0 0;will-change:transform}
#win pre{margin:0;font-family:Menlo,"SF Mono","DejaVu Sans Mono",ui-monospace,monospace;font-size:${S}px!important;line-height:1;letter-spacing:0;white-space:pre;display:block}
</style>
<div id="stage"><div id="win"></div></div>
<script src="fire-frames.js"></script>
<script>
const CAM = ${JSON.stringify(cam.camera)};
const FPS = ${fps}, ROWS = ${clip.rows}, COLSW = ${clip.cols}, S = ${S};
const smooth = (u) => { u = Math.max(0, Math.min(1, u)); return u * u * (3 - 2 * u); };
const win = document.getElementById('win');
let shown = -1, cellW = 0, winW = 0, winH = ROWS * S;
function measure() {
  win.innerHTML = window.FIRE[0];
  let pre = win.querySelector('pre');
  // WHOLE-PIXEL CELLS. The font is sized from the frame's height, so Menlo's
  // advance (0.602 em) lands on a fraction of a pixel, the cell edges drift
  // across the row, and Chrome blends the edge pixel wherever two backgrounds
  // meet: 1-px seams 3-7% off their neighbours at cell boundaries (measured
  // 2026-09-21 on his "look at all the lines in the water and the crabs").
  // So the font is resized to the size whose advance is the nearest whole
  // pixel, with the line height kept at the row's height: every cell edge is
  // then an integer and no edge pixel is blended. The glyphs shrink under 1%.
  const natural = pre.getBoundingClientRect().width / COLSW;
  const ratio = natural / S;
  const target = Math.round(natural) * COLSW;
  // Chrome quantises advances, so the linear guess lands a few thousandths
  // off; search the sizes around it for the one whose row is exactly the
  // target width.
  const st = document.createElement('style');
  document.head.appendChild(st);
  let best = null;
  const guess = Math.round(natural) / ratio;
  for (let k = -40; k <= 40; k++) {
    const f = guess + k * 0.002;
    st.textContent = '#win pre{font-size:' + f.toFixed(4) + 'px!important;line-height:' + S + 'px}';
    const w = win.querySelector('pre').getBoundingClientRect().width;
    const err = Math.abs(w - target);
    if (best === null || err < best.err) best = { f, w, err };
    if (err === 0) break;
  }
  st.textContent = '#win pre{font-size:' + best.f.toFixed(4) + 'px!important;line-height:' + S + 'px}';
  pre = win.querySelector('pre');
  winW = pre.getBoundingClientRect().width;
  cellW = winW / COLSW;
  shown = 0;
  return { winW, cellW, winH, natural, fontPx: best.f, err: best.err };
}
window.measure = measure;
function cameraAt(t) {
  const k = CAM;
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
  const fx = ${W} / 2 - px * z, fy = ${H} / 2 - py * z;
  const w = winW * z, h = winH * z;
  const cx = w >= ${W} ? Math.min(0, Math.max(${W} - w, fx)) : (${W} - w) / 2;
  const cy = h >= ${H} ? Math.min(0, Math.max(${H} - h, fy)) : (${H} - h) / 2;
  const f = cam.free || 0;
  const tx = fx * f + cx * (1 - f), ty = fy * f + cy * (1 - f);
  win.style.transform = 'translate(' + tx.toFixed(2) + 'px,' + ty.toFixed(2) + 'px) scale(' + z.toFixed(4) + ')';
}
window.setFrame = function (t) {
  if (!cellW) measure();
  const k = Math.max(0, Math.min(window.FIRE.length - 1, Math.floor(t * FPS + 1e-6)));
  if (k !== shown) { win.innerHTML = window.FIRE[k]; shown = k; }
  place(cameraAt(Math.max(0, t)));
};
</script>`;

const stagePath = path.resolve(outDir, 'stage.html');
fs.writeFileSync(stagePath, stage);

const browser = await chromium.launch({ channel: 'chrome' });
const page = await browser.newPage({ viewport: { width: W, height: H }, deviceScaleFactor: 1 });
await page.goto('file://' + stagePath);
console.log('window:', JSON.stringify(await page.evaluate(() => window.measure())));

const shot = async (t, file) => {
  await page.evaluate((t) => window.setFrame(t), t);
  await page.screenshot({ path: file, type: 'png', clip: { x: 0, y: 0, width: W, height: H }, animations: 'disabled' });
};

if (stillsOnly) {
  const stills = stillsAt && /^[\d.,]+$/.test(stillsAt) ? stillsAt.split(',').map(Number) : [1, 5, 8, 12, 20, 27];
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
