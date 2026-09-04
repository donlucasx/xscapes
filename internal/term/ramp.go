package term

import (
	"math"
	"strconv"
	"sync"

	"github.com/donlucasx/xscapes/internal/envx"
)

// A Ramp is a gradient the way a 256-colour terminal can actually show one:
// a PATH through the palette from one end to the other, walked in order, each
// tone given the same share of the rows.
//
// The renderer used to blend the two ends in RGB and round every row to the
// nearest palette entry on its own. That is the obvious thing and it is where
// the edge he asked about came from. A ramp from a deep blue zenith to a
// peach horizon passes close to neutral in the middle, and there the nearest
// entry flips between the grey ramp and whichever cube colour is a hair
// closer -- measured at 05:00: blue, grey, teal, grey, grey, grey, olive,
// grey, rose, with a 30-point CIE76 step at every flip. No choice of endpoints
// fixes it, because the flips happen between them, and no per-row quantiser
// can, because a row does not know what the row above it chose.
//
// So the choice is made once for the whole ramp. The 240 fixed palette entries
// are a graph; two are joined when they are one lattice step apart or within
// one small perceptual step of each other; and the ramp is the cheapest walk
// from the quantised start to the quantised end, where a walk pays for the
// SQUARE of every step it takes (so two steps of 15 beat one of 30) and for
// how far the tones it passes through sit from the true-colour ramp they
// stand in for (so it does not wander off through cyan because the steps were
// cheap there). The second charge is what RampLambda scales, and it is
// charged per unit of distance walked rather than per tone visited: charged
// per tone, a walk was rewarded for skipping tones, and the night sky went
// down its greys two at a time.
//
// What this cannot do is make a single cube step small. A green step in the
// middle of the cube is about 30 CIE76 by itself, and every ramp that crosses
// green has to take it somewhere. The path takes it once, on its own, rather
// than together with a red step at the same row.
//
// Truecolor terminals get the true ramp. The path exists for the picture the
// terminal he uses shows.
type Ramp struct {
	A, B RGB
	Path []RGB
}

// Ramps switches the path painter on. XSCAPES_RAMP=0 turns it off and puts
// every sky and sea row back through the row-by-row quantiser, which is the
// picture the change was reviewed against.
var Ramps = true

// RampLambda is how much a tone pays for standing away from the true ramp,
// against how much a step pays for its size. Higher keeps the path closer to
// what the palette asked for, including its greys where the true colour is
// dark and neutral; lower prefers small steps even where they run brighter
// than the truth. XSCAPES_RAMP_LAMBDA overrides it.
var RampLambda = 1.0

// RampHop is the largest single step, in CIE76, the walk may take between two
// entries that are not lattice neighbours. Lattice neighbours are always
// joined, or a ramp that has to cross green could not. XSCAPES_RAMP_HOP
// overrides it.
var RampHop = 26.0

func init() {
	switch envx.Lookup("RAMP") {
	case "0", "off", "no":
		Ramps = false
	}
	if v := envx.Lookup("RAMP_LAMBDA"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 0 && f <= 100 {
			RampLambda = f
		}
	}
	if v := envx.Lookup("RAMP_HOP"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 0 && f <= 100 {
			RampHop = f
		}
	}
}

// rampOrderPenalty is the most a tone pays, in CIE76 of off-line distance,
// for reversing or flattening a channel ordering the true colour states --
// the relationship the background quantiser's ordersKept defends row by row.
//
// From the dusk zenith rgb(0,95,175) the cheapest step is red up to 95,
// which lands on rgb(95,95,175): blue with no green lead, which is violet,
// which is the stripe the quantiser exists to stop. Charged rather than
// forbidden: forbidden, the dawn has no legal first step at all (its true
// ramp is blue-over-green then green-over-red and the cube has no entry that
// keeps both), and a walk that cannot start is worse than one that pays.
//
// The charge is scaled twice, and both scalings were forced by the audit.
// By how strongly the truth states the ordering: at the quantiser's flat 20
// the charge landed on the night's greys for not being blue and on the
// grey-teal a sea passes through for a 24-point lead, and the walk skipped
// tones to avoid them; the violet's lead is 74. And by the tone's own
// chroma: a muted slate that flattens a lead is a tint, a saturated violet
// that flattens the same lead is a stripe, and charged alike the dawn gave
// up its slate for a teal that broke no rule and looked worse.
const rampOrderPenalty = 100.0

// rampOrderCharge is the off-line charge for the orderings a tone breaks.
func rampOrderCharge(src, q RGB) float64 {
	mx, mn := q.R, q.R
	for _, v := range []uint8{q.G, q.B} {
		if v > mx {
			mx = v
		}
		if v < mn {
			mn = v
		}
	}
	chroma := float64(mx-mn) / 255
	if chroma == 0 {
		return 0
	}
	worst := 0.0
	pair := func(a, b, qa, qb uint8) {
		d := int(a) - int(b)
		broken := (d > 0 && qa <= qb) || (d < 0 && qa >= qb)
		if !broken {
			return
		}
		if d < 0 {
			d = -d
		}
		// Nothing under a lead of 20 (the quantiser's "clearly"), the whole
		// charge from 60 up, a straight line between.
		if e := (float64(d) - 20) / 40; e > worst {
			worst = e
		}
	}
	pair(src.R, src.G, q.R, q.G)
	pair(src.G, src.B, q.G, q.B)
	pair(src.R, src.B, q.R, q.B)
	if worst > 1 {
		worst = 1
	}
	if worst < 0 {
		worst = 0
	}
	return rampOrderPenalty * worst * chroma
}

// rampEnd is the palette entry a ramp starts or ends on: the background
// quantiser's answer, the same one every flat background cell gets, so a
// ramp meets the flat tones beside it on the entry they landed on.
//
// (A variant that let plain rounding overrule the ordering rule when it was
// perceptually nearer was measured against the audit and removed: it changed
// four half-hours, two for the better and one for the worse.)
func rampEnd(c RGB) int { return c.Index256Keeping() - 16 }

// NewRamp builds the ramp between two colours. Paths are cached by their
// quantised ends, so a palette drifting through the day costs a search only
// when an end actually lands on a new entry.
//
// The knobs were chosen against notes/gradientaudit at his geometry. The hop
// is 26 because the grey-ramp entries beside a slate blue sit 24.2 to 24.7
// away: at 24 the walk was forbidden the one detour that beats the 32-point
// green step every dawn otherwise takes, and at 28 it started skipping tones.
func NewRamp(a, b RGB) *Ramp {
	return &Ramp{A: a, B: b, Path: rampPath(a, b)}
}

// True is the colour the palette asked for at t: what a truecolor terminal
// paints, and what BGAt reports.
func (r *Ramp) True(t float64) RGB { return Lerp(r.A, r.B, t) }

// Tone is the palette entry the path puts at t. Equal shares: a path of nine
// tones gives each a ninth of the rows, in order.
func (r *Ramp) Tone(t float64) RGB {
	k := len(r.Path)
	if k == 0 {
		return FromIndex256(r.True(t).Index256Keeping())
	}
	i := int(t * float64(k))
	if i < 0 {
		i = 0
	} else if i >= k {
		i = k - 1
	}
	return r.Path[i]
}

// The graph. Node i is palette index 16+i: the 216 cube entries, then the 24
// greys. Built once.
const rampNodes = 240

var (
	rampOnce sync.Once
	rampRGB  [rampNodes]RGB
	rampLab  [rampNodes][3]float64
	rampDE   [rampNodes][rampNodes]float32
)

func rampInit() {
	for i := range rampRGB {
		rampRGB[i] = FromIndex256(16 + i)
		rampLab[i] = lab(rampRGB[i])
	}
	for i := range rampRGB {
		for j := range rampRGB {
			rampDE[i][j] = float32(labDist(rampLab[i], rampLab[j]))
		}
	}
}

var (
	rampMu    sync.Mutex
	rampCache = map[[4]int][]RGB{}
)

func rampPath(a, b RGB) []RGB {
	rampOnce.Do(rampInit)
	s, g := rampEnd(a), rampEnd(b)
	key := [4]int{s, g, int(RampLambda * 1000), int(RampHop * 10)}
	rampMu.Lock()
	defer rampMu.Unlock()
	if p, ok := rampCache[key]; ok {
		return p
	}
	p := rampSearch(a, b, s, g)
	rampCache[key] = p
	return p
}

// rampAdjacent says whether two nodes are joined: one lattice step apart in
// the cube, neighbours on the grey ramp, or within rampHop of each other.
func rampAdjacent(u, v int) bool {
	if float64(rampDE[u][v]) <= RampHop {
		return true
	}
	switch {
	case u < 216 && v < 216:
		d := 0
		for _, p := range [3][2]int{{u / 36, v / 36}, {(u / 6) % 6, (v / 6) % 6}, {u % 6, v % 6}} {
			if p[0] > p[1] {
				d += p[0] - p[1]
			} else {
				d += p[1] - p[0]
			}
		}
		return d == 1
	case u >= 216 && v >= 216:
		return u-v == 1 || v-u == 1
	}
	return false
}

func rampLuma(c RGB) float64 { return 0.30*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B) }

func rampSearch(a, b RGB, s, g int) []RGB {
	if s == g {
		return []RGB{rampRGB[s]}
	}
	qa, qb := rampRGB[s], rampRGB[g]

	// Which entries the walk may visit. Cube entries: no channel leaves the
	// range the two ends span, so a blue-to-peach dawn cannot detour through
	// green because the steps happened to be cheap there. Greys: those
	// between the ends in brightness, a ramp step of slack either side.
	lo := RGB{minU8(qa.R, qb.R), minU8(qa.G, qb.G), minU8(qa.B, qb.B)}
	hi := RGB{maxU8(qa.R, qb.R), maxU8(qa.G, qb.G), maxU8(qa.B, qb.B)}
	lmin, lmax := rampLuma(qa), rampLuma(qb)
	if lmin > lmax {
		lmin, lmax = lmax, lmin
	}
	// The true ramp, sampled, in Lab. A tone stands in for the point of the
	// ramp it is perceptually NEAREST to. Projecting in RGB was tried first
	// and put the violet rgb(95,95,175) a third of the way down a dusk sky,
	// where the line has turned toward orange and its green lead is weak, so
	// the ordering charge never fired -- red spans 255 on that ramp and drags
	// the projection along. In Lab the same tone is nearest the blue top,
	// where the lead is 74, which is where the terminal shows it.
	const samples = 48
	var lineRGB [samples]RGB
	var lineLab [samples][3]float64
	for k := range lineRGB {
		lineRGB[k] = Lerp(a, b, float64(k)/float64(samples-1))
		lineLab[k] = lab(lineRGB[k])
	}
	var allowed []int
	var off [rampNodes]float64
	for i, c := range rampRGB {
		ok := false
		switch {
		case i == s || i == g:
			ok = true
		case i < 216:
			ok = c.R >= lo.R && c.R <= hi.R && c.G >= lo.G && c.G <= hi.G && c.B >= lo.B && c.B <= hi.B
		default:
			l := rampLuma(c)
			ok = l >= lmin-10 && l <= lmax+10
		}
		if !ok {
			continue
		}
		// The tone's distance from the true ramp, in CIE76 like the steps,
		// so a wrong hue and a wrong brightness are charged in the same coin.
		// (The background quantiser's hue-weighted RGB distance was tried
		// here first and could not tell a khaki from a tan against a beige:
		// both were "close", and the walk took the khaki.)
		near, nearD := 0, math.Inf(1)
		for k := range lineLab {
			if d := labDist(lineLab[k], rampLab[i]); d < nearD {
				near, nearD = k, d
			}
		}
		off[i] = nearD + rampOrderCharge(lineRGB[near], c)
		allowed = append(allowed, i)
	}

	// Dijkstra over a few dozen allowed entries. Two objectives were tried
	// and rejected against the rendered audit before this one was kept:
	// minimising the LARGEST step first found smaller steps by leaving the
	// ramp altogether (a dawn through hot pink, a noon through violet, a sea
	// through cyan); and refusing any tone more than a bound from the true
	// ramp cut the graph in two wherever the ramp passes through a neutral
	// no entry sits near. The sum of squared steps, plus the off-line charge
	// along the way, is the one that produced honest walks.
	var dist [rampNodes]float64
	var prev [rampNodes]int
	var done [rampNodes]bool
	for i := range dist {
		dist[i] = math.Inf(1)
		prev[i] = -1
	}
	dist[s] = 0
	for {
		u, best := -1, math.Inf(1)
		for _, i := range allowed {
			if !done[i] && dist[i] < best {
				u, best = i, dist[i]
			}
		}
		if u == -1 || u == g {
			break
		}
		done[u] = true
		for _, v := range allowed {
			if done[v] || !rampAdjacent(u, v) {
				continue
			}
			step := float64(rampDE[u][v])
			if c := dist[u] + step*step + RampLambda*step*(off[u]+off[v])/2; c < dist[v] {
				dist[v], prev[v] = c, u
			}
		}
	}
	if prev[g] == -1 {
		// Unreachable inside the box, which the lattice edges should make
		// impossible; the two ends on their own is the honest fallback.
		return []RGB{qa, qb}
	}
	var path []RGB
	for v := g; v != -1; v = prev[v] {
		path = append(path, rampRGB[v])
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

func minU8(a, b uint8) uint8 {
	if a < b {
		return a
	}
	return b
}

func maxU8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}

// DeltaE is CIE76 between two sRGB colours: the size of a step as the eye
// ranks it. Under 5 is barely visible; over 20 is a hard edge.
func DeltaE(a, b RGB) float64 { return labDist(lab(a), lab(b)) }

func labDist(p, q [3]float64) float64 {
	return math.Sqrt((p[0]-q[0])*(p[0]-q[0]) + (p[1]-q[1])*(p[1]-q[1]) + (p[2]-q[2])*(p[2]-q[2]))
}

func lab(c RGB) [3]float64 {
	lin := func(v uint8) float64 {
		f := float64(v) / 255
		if f <= 0.04045 {
			return f / 12.92
		}
		return math.Pow((f+0.055)/1.055, 2.4)
	}
	r, g, bl := lin(c.R), lin(c.G), lin(c.B)
	x := (0.4124*r + 0.3576*g + 0.1805*bl) / 0.95047
	y := (0.2126*r + 0.7152*g + 0.0722*bl) / 1.0
	z := (0.0193*r + 0.1192*g + 0.9505*bl) / 1.08883
	f := func(t float64) float64 {
		if t > 0.008856 {
			return math.Cbrt(t)
		}
		return 7.787*t + 16.0/116
	}
	fx, fy, fz := f(x), f(y), f(z)
	return [3]float64{116*fy - 16, 500 * (fx - fy), 200 * (fy - fz)}
}
