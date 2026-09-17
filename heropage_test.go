package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
)

// TestHeroPage writes the six hero variants (montage.go) as a review page for
// him to pick from:
//
//	XSCAPES_HEROPAGE=<dir> go test -run TestHeroPage .
//
// index.html plus one frames file per variant (the six together are near the
// size a single page may be), in the page's own window furniture and player,
// so what he rules on is what the site would play. Skipped unless the
// directory is named, so the suite never writes anywhere.
func TestHeroPage(t *testing.T) {
	dir := os.Getenv("XSCAPES_HEROPAGE")
	if dir == "" {
		t.Skip("set XSCAPES_HEROPAGE=<dir> to write the review page")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	pal := &canvas.HTMLPalette{}
	var secs strings.Builder
	var scripts strings.Builder
	var sheets [][2]string
	for i, m := range heroVariants() {
		frames, err := montageFrames(7, m, pal)
		if err != nil {
			t.Fatal(err)
		}
		b, err := json.Marshal(fxPayload{Cols: m.cols, Rows: m.rows + m.agentRows, FPS: m.fps,
			Label: m.label, Note: m.note, Frames: frames})
		if err != nil {
			t.Fatal(err)
		}
		js := "window.FX=window.FX||{};window.FX[" + strconv(m.key) + "]=" + string(b) + ";"
		if err := os.WriteFile(filepath.Join(dir, m.key+".js"), []byte(js), 0o644); err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&scripts, `<script src="%s.js"></script>`+"\n", m.key)
		// A contact sheet per variant: six frames evenly spaced plus the
		// frame just after every scape or companion switch, for a
		// headless-Chrome look before the page is published (the s32 rule:
		// every visual defect was seen in a screenshot, none in a dump).
		var sheet strings.Builder
		sheet.WriteString(`<meta charset="utf-8"><style>body{margin:0;padding:8px;background:#000;color:#ccc;font:12px Menlo,monospace}` +
			`.fx pre{margin:0;font-family:Menlo,"SF Mono",monospace;line-height:1;letter-spacing:0;white-space:pre;display:block;font-size:9px!important}` +
			`.row{display:flex;gap:8px;align-items:flex-start;margin-bottom:6px}.lab{width:60px;padding-top:4px}</style>`)
		picks := map[int]bool{}
		for k := 0; k < 6; k++ {
			picks[k*len(frames)/6] = true
		}
		for _, sg := range m.segs {
			if f := int(sg.at/m.speed*float64(m.fps)) + 4; f < len(frames) {
				picks[f] = true
			}
		}
		var order []int
		for f := range picks {
			order = append(order, f)
		}
		sort.Ints(order)
		for _, at := range order {
			fmt.Fprintf(&sheet, `<div class="row"><div class="lab">%d<br>%.1fs</div><div class="fx">%s</div></div>`, at, float64(at)/float64(m.fps), frames[at])
		}
		sheets = append(sheets, [2]string{m.key, sheet.String()})
		fmt.Fprintf(&secs, heroSection, m.key, i+1, strings.ToUpper(m.key), esc(m.label), m.secs, len(frames), esc(m.note))
		t.Logf("%s: %d frames, %.1f s, %.2f MB", m.key, len(frames), m.secs, float64(len(b))/1e6)
	}
	for _, sh := range sheets {
		if err := os.WriteFile(filepath.Join(dir, "sheet-"+sh[0]+".html"), []byte(sh[1]+"<style>"+pal.CSS()+"</style>"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	page := strings.Replace(heroPageHTML, "{{css}}", pal.CSS(), 1)
	page = strings.Replace(page, "{{sections}}", secs.String(), 1)
	page = strings.Replace(page, "{{scripts}}", scripts.String(), 1)
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}
}

func strconv(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func esc(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

// heroSection is one variant: key, number, KEY, label, seconds, frames, note.
const heroSection = `<section id="%s">
<p class="kick"><b>%d &middot; %s</b> &nbsp;%s &nbsp;<i>%.0f s loop &middot; %d frames</i></p>
<div class="win" data-title="xscapes claude" data-right="%[4]s"><div class="boxtop" aria-hidden="true"></div>
  <div class="stage" aria-hidden="true"><div class="fx" data-clip="%[1]s"></div></div>
  <div class="boxbot" aria-hidden="true"></div></div>
<p class="cap">%[7]s</p>
</section>
`

const heroPageHTML = `<title>Hero Variants</title>
<style>
:root{--g:#121212;--s:#1a1a1a;--ln:#2e2e2e;--d:#8a8a8a;--dd:#bcbcbc;--i:#eeeeee;--acc:#87afff;--warm:#d7af5f;--lh:1.55rem}
*{box-sizing:border-box}
html,body{margin:0;padding:0;background:var(--g);color:var(--i)}
body{font-family:"JetBrains Mono",ui-monospace,"SF Mono",Menlo,monospace;font-size:14px;line-height:var(--lh);
  font-variant-numeric:tabular-nums lining-nums;-webkit-font-smoothing:antialiased}
main{max-width:min(124ch,100%);margin:0 auto;padding:var(--lh) 16px calc(var(--lh)*4)}
h1{font-size:18px;font-weight:700;letter-spacing:-.01em;margin:0 0 6px;text-wrap:balance}
p{max-width:72ch;margin:0 0 var(--lh);color:var(--dd)}
p.kick{margin:calc(var(--lh)*2) 0 8px;color:var(--d)}
p.kick b{color:var(--i);letter-spacing:.06em}
p.kick i{font-style:normal;color:var(--d)}
p.cap{color:var(--dd);font-size:13.5px;margin-top:10px}
p.rule{color:var(--d);font-size:13px}
nav{display:flex;flex-wrap:wrap;gap:6px 14px;margin:0 0 var(--lh);font-size:13px}
nav a{color:var(--acc);text-decoration:none;border-bottom:1px solid rgba(135,175,255,.35)}
nav a:focus-visible{outline:2px solid var(--acc);outline-offset:2px}
.win{max-width:100%;margin:0}
.boxtop,.boxbot{white-space:pre;overflow:hidden;color:var(--d);font-size:13px;line-height:1.15;-webkit-user-select:none;user-select:none}
.boxtop b{color:var(--i);font-weight:500}
.boxtop i{color:var(--d);font-style:normal}
.stage{overflow-x:auto;overflow-y:hidden;background:#000;cursor:ew-resize;touch-action:pan-y}
.fx{display:block;width:max-content;min-height:1px;padding-bottom:.45em;font-size:10px}
.fx pre{margin:0;font-family:Menlo,"SF Mono","DejaVu Sans Mono",ui-monospace,monospace;line-height:1;letter-spacing:0;white-space:pre;display:block;font-size:inherit!important;
  -webkit-text-size-adjust:100%%;text-size-adjust:100%%}
ul{max-width:72ch;color:var(--dd);padding-left:22px;margin:0 0 var(--lh)}
li{margin:0 0 6px}
.rulings{border-top:1px solid var(--ln);margin-top:calc(var(--lh)*2);padding-top:var(--lh)}
@media (prefers-reduced-motion:reduce){.stage{cursor:default}}
{{css}}
</style>
<main>
<h1>The hero, six ways</h1>
<p>Each window is the page's own hero player at its real size, 124 columns, fourteen rows of transcript over a thirty-row scape, on the site's furniture. Drag across a window to scrub the session; leave it and it plays on. The first two are the vista alone, the next two show both scapes in one session, the last two are the trailer you described.</p>
<nav><a href="#v1">1 V1</a><a href="#v2">2 V2</a><a href="#h1">3 H1</a><a href="#h2">4 H2</a><a href="#m1">5 M1</a><a href="#m2">6 M2</a><a href="#rulings">what to rule on</a></nav>
{{sections}}
<section id="rulings" class="rulings">
<p class="kick"><b>WHAT TO RULE ON</b></p>
<ul>
<li><b>Which one leads the page</b>, or which two to keep working on. The wide shot and its narrow twin for phones are rendered from the same script, so the pick costs nothing extra.</li>
<li><b>In the trailer, what each companion does when it arrives.</b> Today the cat and the crab show their finished face, and the owl's own finish knock is still up when the cat takes its place. The alternative is each one asking: the cat with its ears up, the crab with its claws down, no balloon.</li>
<li><b>The cut from the vista to the shore.</b> Today it is a hard cut at the same hour, the context and the transcript carried across, which puts the sun at a similar height in both. The two scapes place the disc differently, so it is the same hour and the same fullness, not the same pixel.</li>
<li><b>The typed commands in M2.</b> They are the product's real commands, typed as shell lines at the agent's prompt, which is what the switch looks like in a session today. If they feel like clutter, M1 is the same clip without them.</li>
</ul>
<p class="rule">Every clip is the reducer's own state folded through the same session machinery as the live scape, at the page's truecolor; nothing in them is posed except the two finished faces in the trailer, which are held on purpose.</p>
</section>
</main>
{{scripts}}
<script>
(function(){
  var FX = window.FX || {};
  var reduce = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  function esc(s){ return s.replace(/&/g,"&amp;").replace(/</g,"&lt;"); }
  function chWidth(el){
    var p = document.createElement("span");
    p.style.cssText = "position:absolute;visibility:hidden;white-space:pre";
    p.textContent = new Array(101).join("0");
    el.appendChild(p);
    var w = p.getBoundingClientRect().width / 100;
    el.removeChild(p);
    return w > 1 ? w : 8;
  }
  var ADV = null;
  function advance(el){
    if(ADV) return ADV;
    var p = document.createElement("pre");
    p.style.cssText = "position:absolute;visibility:hidden;margin:0;font-size:100px!important;line-height:1;white-space:pre";
    p.textContent = "0123456789";
    el.appendChild(p);
    ADV = (p.getBoundingClientRect().width / 10) / 100;
    el.removeChild(p);
    if(!(ADV > 0.3 && ADV < 0.9)) ADV = 0.6;
    return ADV;
  }
  function fit(el, cols){
    var box = el.parentElement, w = box ? box.clientWidth : 0;
    if(!w) return;
    var px = (w / cols) / advance(el);
    el.style.fontSize = Math.max(7, Math.min(44, px)) + "px";
  }
  var fitted = [];
  function play(el, key, startFrac){
    var clip = FX[key];
    if(!el || !clip || !clip.frames.length) return null;
    var n = clip.frames.length, i = Math.floor((startFrac||0) * n) % n, id = null;
    function show(k){ i = ((k % n) + n) % n; el.innerHTML = clip.frames[i]; }
    show(i); fit(el, clip.cols); fitted.push(function(){ fit(el, clip.cols); });
    function start(){ if(reduce || id) return; id = setInterval(function(){ show(i+1); }, 1000/(clip.fps||6)); }
    function stop(){ if(id){ clearInterval(id); id = null; } }
    start();
    var stage = el.parentElement, scrubbing = false;
    function toFrame(ev){
      var r = stage.getBoundingClientRect();
      var x = (ev.clientX - r.left) / r.width;
      show(Math.floor(Math.max(0, Math.min(0.999, x)) * n));
    }
    if(!reduce){
      stage.addEventListener("pointerenter", function(ev){ if(ev.pointerType === "touch") return; scrubbing = true; stop(); });
      stage.addEventListener("pointermove", function(ev){ if(scrubbing) toFrame(ev); });
      stage.addEventListener("pointerleave", function(ev){ if(ev.pointerType === "touch") return; scrubbing = false; start(); });
      stage.addEventListener("pointerdown", function(ev){ if(ev.pointerType !== "touch") return; scrubbing = true; stop(); toFrame(ev); if(stage.setPointerCapture) try{ stage.setPointerCapture(ev.pointerId); }catch(e){} });
      var end = function(ev){ if(ev.pointerType !== "touch") return; scrubbing = false; start(); };
      stage.addEventListener("pointerup", end);
      stage.addEventListener("pointercancel", end);
    }
    return {show:show, start:start, stop:stop};
  }
  function drawBox(win){
    var top = win.querySelector(".boxtop"), bot = win.querySelector(".boxbot");
    var n = Math.max(14, Math.floor(win.clientWidth / chWidth(top)));
    var title = (win.getAttribute("data-title") || "").trim(), right = (win.getAttribute("data-right") || "").trim();
    var head = "┌─ " + title + " ", tail = " " + right + " ─┐";
    var fill = n - head.length - tail.length;
    if(fill < 2){ tail = "┐"; right = ""; fill = n - head.length - tail.length; }
    var dash = new Array(Math.max(1, fill) + 1).join("─");
    top.innerHTML = "┌─ <b>" + esc(title) + "</b> " + dash + (right ? " <i>" + esc(right) + "</i> ─┐" : "┐");
    bot.textContent = "└" + new Array(Math.max(2, n - 2) + 1).join("─") + "┘";
  }
  Array.prototype.forEach.call(document.querySelectorAll("[data-clip]"), function(el){ play(el, el.getAttribute("data-clip"), 0); });
  function refit(){ ADV = null; fitted.forEach(function(f){ f(); }); Array.prototype.forEach.call(document.querySelectorAll(".win"), drawBox); }
  refit();
  window.addEventListener("resize", refit);
  requestAnimationFrame(refit);
  if(document.fonts && document.fonts.ready){ document.fonts.ready.then(refit); }
})();
</script>
`
