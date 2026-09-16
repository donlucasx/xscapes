#!/usr/bin/env python3
"""Candidates for a more natural ask/done cue (s34, 2026-09-15).

HIS ASK: "I dont love the sound that xscapes makes rn when it prompts the user.
Can we change it for something that sounds more 'natural' or 'organic'? a
water dropplet? can draft a few sample sounds I can test and compare"

WHY THE SHIPPED CUE DOES NOT READ AS WATER. notes/s28-sound's "droplet" is one
sine whose pitch glides DOWN (ask 1180 -> 660 Hz, done 700 -> 420) under a
smooth envelope. A real drop is the reverse, and it has a part the sine lacks:
the impact is a few milliseconds of broadband noise (the TICK), and the sound
that follows is an air bubble the impact pinched off, ringing at its Minnaert
frequency. That ring CLIMBS, because the free surface lowers a bubble's
resonance and the bubble is sinking away from it. Tick, then a ring that
rises and dies inside ~150 ms: that is the anatomy the ear files as "drop",
and the shipped cue has neither the tick nor the direction.

Every water candidate here is built from that anatomy. The two non-water ones
(bamboo, bird) are here because he said "natural or organic", not "water",
and a knock or a whistle is the honest way to find out which he meant.

The pair rule from s28 still holds: ask and done must differ in SHAPE, not
only pitch, because a listener in another pane has nothing but the shape.
Here ask is the smaller, brighter, clearly climbing event (a question) and
done is the heavier, lower one that lands (a settling): a deeper drop, a
falling pair, a knock that rocks back.

stdlib only, like s28. 16-bit mono 44.1k, every file peak-normalised to the
same 0.42 the shipped cues use, so the page compares them at one volume.
"""
import json
import math
import os
import random
import struct
import wave

SR = 44100
OUT = os.path.dirname(os.path.abspath(__file__))
GAIN = 0.42  # the shipped cues' peak; keep every candidate at the same level


def n_of(sec):
    return int(sec * SR)


def mix(*parts):
    """parts: (offset_sec, samples, gain). Overlapping events summed."""
    n = max(int(o * SR) + len(s) for o, s, _ in parts)
    out = [0.0] * n
    for o, s, g in parts:
        off = int(o * SR)
        for i, v in enumerate(s):
            out[off + i] += v * g
    return out


class Biquad:
    """RBJ bandpass. Used only to colour the impact noise."""

    def __init__(self, fc, q):
        w = 2 * math.pi * fc / SR
        sn, cs = math.sin(w), math.cos(w)
        al = sn / (2 * q)
        a0 = 1 + al
        self.b0, self.b1, self.b2 = al / a0, 0.0, -al / a0
        self.a1, self.a2 = -2 * cs / a0, (1 - al) / a0
        self.x1 = self.x2 = self.y1 = self.y2 = 0.0

    def step(self, x):
        y = self.b0 * x + self.b1 * self.x1 + self.b2 * self.x2 - self.a1 * self.y1 - self.a2 * self.y2
        self.x2, self.x1 = self.x1, x
        self.y2, self.y1 = self.y1, y
        return y


def tick(dur=0.007, fc=4200, q=2.5, tau=0.0016, seed=1):
    """The impact: a burst of noise, coloured, gone in a few milliseconds."""
    rnd = random.Random(seed)
    bp = Biquad(fc, q)
    out = []
    for i in range(n_of(dur)):
        t = i / SR
        out.append(bp.step(rnd.uniform(-1, 1)) * math.exp(-t / tau))
    return out


def ring(f0, f1, dur, tau, rise=0.06, h2=0.12, attack=0.0012):
    """The bubble: a decaying sine whose pitch climbs f0 -> f1 over `rise`
    seconds (fast at first, landing soft) and then holds. tau is the amplitude
    time constant. h2 is a touch of second harmonic, decaying twice as fast,
    which keeps the first millisecond from sounding like a test tone."""
    out = []
    ph = 0.0
    span = 1 - math.exp(-3.5)
    for i in range(n_of(dur)):
        t = i / SR
        k = min(1.0, t / rise)
        f = f0 + (f1 - f0) * (1 - math.exp(-3.5 * k)) / span
        ph += 2 * math.pi * f / SR
        a = math.exp(-t / tau)
        if t < attack:
            a *= t / attack
        out.append((math.sin(ph) + h2 * math.sin(2 * ph) * math.exp(-t / (tau * 0.5))) * a)
    return out


def thump(f=110, dur=0.09, tau=0.028):
    """The low body of a heavier drop: the water itself moving."""
    out = []
    ph = 0.0
    for i in range(n_of(dur)):
        t = i / SR
        ph += 2 * math.pi * (f * (1 - 0.3 * min(1, t / 0.05))) / SR
        out.append(math.sin(ph) * math.exp(-t / tau) * min(1, t / 0.002))
    return out


def tail(x, delays=(0.0231, 0.0371, 0.0533, 0.0797), fb=0.58, damp=0.22, wet=0.30, extra=0.32):
    """A small stone room around a sound: four damped feedback delays, wet
    part mixed in low. Enough to say 'cistern', not enough to say 'hall'."""
    n = len(x) + n_of(extra)
    dry = [x[i] if i < len(x) else 0.0 for i in range(n)]
    wetsum = [0.0] * n
    for d in delays:
        D = n_of(d)
        buf = [0.0] * n
        lp = 0.0
        for i in range(n):
            back = buf[i - D] if i >= D else 0.0
            lp += damp * (back - lp)
            buf[i] = dry[i] + fb * lp
            wetsum[i] += buf[i] - dry[i]
    return [dry[i] + wet * wetsum[i] / len(delays) for i in range(n)]


def knock(f, dur=0.16, partials=((1.0, 1.0, 0.055), (2.31, 0.45, 0.028), (3.87, 0.22, 0.016), (5.6, 0.08, 0.010))):
    """A hollow wooden knock: inharmonic partials with their own decays, a
    sharp attack, and a tick for the strike itself."""
    body = []
    for i in range(n_of(dur)):
        t = i / SR
        s = 0.0
        for mult, amp, tau in partials:
            s += amp * math.sin(2 * math.pi * f * mult * t) * math.exp(-t / tau)
        body.append(s * min(1, t / 0.0008))
    return mix((0.0, tick(dur=0.004, fc=2600, q=1.2, tau=0.0009, seed=7), 0.5),
               (0.0015, body, 1.0),
               (0.0, thump(f * 0.28, dur=0.05, tau=0.014), 0.35))


def whistle(segs, gap=0.028, vib=0.012, vibf=38.0):
    """A small bird: pure-tone segments, each a pitch glide with a soft edge.
    segs: (f0, f1, dur)."""
    parts = []
    at = 0.0
    for f0, f1, dur in segs:
        out = []
        ph = 0.0
        n = n_of(dur)
        for i in range(n):
            t = i / SR
            k = i / max(1, n - 1)
            f = f0 * (f1 / f0) ** k
            f *= 1 + vib * math.sin(2 * math.pi * vibf * t)
            ph += 2 * math.pi * f / SR
            a = min(1, t / 0.012, (dur - t) / 0.02)
            out.append(math.sin(ph) * a)
        parts.append((at, out, 1.0))
        at += dur + gap
    return mix(*parts)


# ---------------------------------------------------------------- the drops

def plink(f0, f1, dur=0.20, tau=0.045, tick_fc=5000, seed=1, h2=0.12, rise=0.06):
    """One drop into still water: tick, then the climbing ring 4 ms later."""
    return mix((0.0, tick(fc=tick_fc, seed=seed), 0.55),
               (0.004, ring(f0, f1, dur, tau, rise=rise, h2=h2), 1.0))


def write(name, samples, level=1.0):
    """level scales below the shared peak: a sustained whistle at the drops'
    peak is ~8 dB louder in RMS, so bird is written at half amplitude."""
    peak = max(1e-9, max(abs(s) for s in samples)) / level
    path = os.path.join(OUT, name + ".wav")
    with wave.open(path, "w") as w:
        w.setnchannels(1)
        w.setsampwidth(2)
        w.setframerate(SR)
        w.writeframes(b"".join(
            struct.pack("<h", int(max(-1, min(1, s / peak * GAIN)) * 32767))
            for s in samples))
    rms = math.sqrt(sum((s / peak * GAIN) ** 2 for s in samples) / len(samples))
    return path, len(samples) / SR, 20 * math.log10(max(rms, 1e-9))


FAMILIES = [
    # (key, name, what it is, ask samples, done samples, what differs)
    ("plink", "Plink",
     "One drop into still water. A tick, then the bubble's ring climbing and dying inside a fifth of a second.",
     plink(950, 1500),
     mix((0.0, plink(520, 700, dur=0.27, tau=0.062, tick_fc=3200, seed=2, rise=0.09), 1.0),
         (0.0, thump(), 0.32)),
     "Ask is small and bright and clearly climbs. Done is a heavier drop, lower, with the water's own thump under it."),
    ("cave", "Cistern",
     "The same drop inside a stone cistern: a short, damped tail says the room is small and wet.",
     tail(plink(720, 1120, dur=0.20, tau=0.05, tick_fc=4200, seed=3)),
     tail(mix((0.0, plink(430, 560, dur=0.28, tau=0.07, tick_fc=2800, seed=4, rise=0.1), 1.0),
              (0.0, thump(95), 0.3))),
     "Same pair as Plink, an octave lower and in a room. The longest of the water options."),
    ("twin", "Drip and answer",
     "Two drops, the second answering the first. Drips come in pairs, and a pair reads as a cue rather than a ping.",
     mix((0.0, plink(700, 1050, dur=0.17, tau=0.04, seed=5), 1.0),
         (0.14, plink(1050, 1600, dur=0.19, tau=0.042, tick_fc=5600, seed=6), 0.85)),
     mix((0.0, plink(1000, 1350, dur=0.16, tau=0.04, tick_fc=4600, seed=8), 0.9),
         (0.16, plink(540, 700, dur=0.26, tau=0.06, tick_fc=3000, seed=9, rise=0.09), 1.0),
         (0.16, thump(100), 0.28)),
     "Ask: the second drop is higher, so the pair rises. Done: the second is lower and heavier, so the pair falls and lands."),
    ("bloop", "Pebble",
     "A pebble into deep water. A bigger bubble rings lower and longer, and the splash has more body.",
     mix((0.0, tick(dur=0.014, fc=2400, q=1.4, tau=0.004, seed=11), 0.6),
         (0.006, ring(230, 420, 0.30, 0.075, rise=0.13, h2=0.28), 1.0)),
     mix((0.0, tick(dur=0.018, fc=1800, q=1.2, tau=0.005, seed=12), 0.6),
         (0.007, ring(150, 235, 0.38, 0.095, rise=0.16, h2=0.3), 1.0),
         (0.0, thump(80, dur=0.12, tau=0.04), 0.45)),
     "Ask climbs a full octave. Done barely climbs and sits on a low thump: something went in and the water closed over it."),
    ("bamboo", "Bamboo",
     "A hollow wooden knock, the deer-scarer in a garden. Not water: the one dry option, and the least likely to wear out.",
     knock(640),
     mix((0.0, knock(430), 1.0), (0.13, knock(380, dur=0.14), 0.55)),
     "Ask is one knock, high. Done is a lower knock and the arm rocking back, softer."),
    ("bird", "Bird",
     "A small whistle, two notes. The most 'alive' of the six and the riskiest: a voice in the room every time the agent asks.",
     whistle([(1900, 2500, 0.075), (2300, 2950, 0.095)]),
     whistle([(2600, 2000, 0.10), (1750, 1350, 0.13)]),
     "Ask: two notes rising, tu-wee. Done: two notes falling, wee-oo."),
]

manifest = []
for key, name, what, ask, done, differs in FAMILIES:
    rows = {}
    for kind, buf in (("ask", ask), ("done", done)):
        path, sec, rms = write(f"{key}-{kind}", buf, level=0.5 if key == "bird" else 1.0)
        rows[kind] = {"file": os.path.basename(path), "ms": round(sec * 1000), "rms_db": round(rms, 1)}
        print(f"{os.path.basename(path):18s} {sec*1000:6.0f} ms  rms {rms:6.1f} dBFS")
    manifest.append({"key": key, "name": name, "what": what, "differs": differs, "ask": rows["ask"], "done": rows["done"]})

with open(os.path.join(OUT, "manifest.json"), "w") as f:
    json.dump(manifest, f, indent=1)
print("manifest.json written")
