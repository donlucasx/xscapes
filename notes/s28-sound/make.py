#!/usr/bin/env python3
"""Synthesise xscapes' own notification cues.

HIS ASK, 2026-09-11: "would love to find a more on-theme sound cue for xscapes,
but that its not annoying. Looking for a fun sound that users can identify w
xscapes."

Why synthesise rather than pick a system sound: the shipped cues are
Glass.aiff and Submarine.aiff, which are macOS's own. They are fine and they
are NOT IDENTIFIABLE -- every Mac user has heard Glass for twenty years, and a
sound that belongs to the OS cannot belong to the product. A signature cue has
to be ours.

The design rules, taken from the scape's own vocabulary rather than invented:
  - ASK and DONE must differ in SHAPE, not just pitch, for the same reason
    Bubble and DoneBubble do: "shape carries the done/needs_input distinction
    wherever colour cannot". Ask RISES (a question). Done FALLS (a settling).
  - Short. Under 500 ms. It fires next to a permission prompt he is already
    reading, not as an event of its own.
  - Quiet and soft-edged. Every attack is ramped; a click is what makes a
    repeated sound hateful.
  - On theme is WATER and SHORE: drops, shells, a small bell over a bay. Not
    a chime-y app notification and not a game coin.

stdlib only -- no numpy, no pydub. Writes 16-bit mono WAV at 44.1k.
"""
import math, struct, wave, os

SR = 44100
OUT = os.path.dirname(os.path.abspath(__file__))

def env(n, total, attack=0.004, decay=None, curve=2.2):
    """Soft attack, exponential-ish decay. The attack ramp is the whole reason
    these do not click."""
    a = int(attack * SR)
    if n < a:
        return n / a
    if decay is None:
        decay = total
    t = (n - a) / max(1, decay * SR)
    return math.exp(-curve * t)

def sine(f, t):
    return math.sin(2 * math.pi * f * t)

def write(name, samples, gain=0.5):
    peak = max(1e-9, max(abs(s) for s in samples))
    path = os.path.join(OUT, name + ".wav")
    with wave.open(path, "w") as w:
        w.setnchannels(1); w.setsampwidth(2); w.setframerate(SR)
        w.writeframes(b"".join(
            struct.pack("<h", int(max(-1, min(1, s / peak * gain)) * 32767))
            for s in samples))
    return path

def drop(f0, f1, dur, curve=2.6, shimmer=0.0):
    """A water drop: pitch glides f0 -> f1 while the body decays. The glide is
    what the ear hears as a droplet rather than a beep."""
    n = int(dur * SR); out = []
    for i in range(n):
        t = i / SR
        k = i / n
        f = f0 * (f1 / f0) ** (k ** 0.6)
        s = sine(f, t)
        if shimmer:
            s += shimmer * sine(f * 2.02, t)
        out.append(s * env(i, dur, decay=dur * 0.45, curve=curve))
    return out

def bell(f, dur, partials=((1, 1.0), (2.76, 0.34), (5.4, 0.12)), curve=3.0):
    """A struck shell/bell. Inharmonic partials are what separate a bell from
    an organ; the higher ones decay faster, as they do in metal."""
    n = int(dur * SR); out = []
    for i in range(n):
        t = i / SR
        s = 0.0
        for mult, amp in partials:
            s += amp * sine(f * mult, t) * math.exp(-curve * mult * 0.35 * t)
        out.append(s * env(i, dur, decay=dur, curve=0.8))
    return out

def two(a, b, gap):
    """Two events, the second starting `gap` seconds in, overlapped."""
    n = max(len(a), int(gap * SR) + len(b))
    out = [0.0] * n
    for i, s in enumerate(a): out[i] += s
    off = int(gap * SR)
    for i, s in enumerate(b): out[off + i] += s * 0.9
    return out

# ---------------------------------------------------------------- candidates
made = []

# 1. DROPLET -- one drop into still water. The most literally on-theme thing
#    the shore has, and the shortest.
made.append(("ask-droplet",  drop(1180, 660, 0.26, shimmer=0.18)))
made.append(("done-droplet", drop(700, 420, 0.34, shimmer=0.10)))

# 2. SHELL -- a small struck bell, warm and round. Ask is two notes rising a
#    fourth; done is the same pair falling.
made.append(("ask-shell",  two(bell(784, 0.30), bell(1046, 0.42), 0.085)))   # G5 -> C6
made.append(("done-shell", two(bell(1046, 0.26), bell(698, 0.46), 0.085)))   # C6 -> F5

# 3. TIDE -- two drops, the second answering the first. Reads as a pair, which
#    is what makes it a "cue" rather than a "ping".
made.append(("ask-tide",  two(drop(900, 780, 0.20), drop(1280, 1100, 0.30, shimmer=0.2), 0.10)))
made.append(("done-tide", two(drop(1280, 1100, 0.18), drop(760, 560, 0.38, shimmer=0.12), 0.10)))

# 4. PEBBLE -- a dry wooden knock, almost no tail. The least musical option and
#    the least likely to wear out, which is the thing he is worried about.
def pebble(f, dur):
    n = int(dur * SR); out = []
    for i in range(n):
        t = i / SR
        s = sine(f, t) * 0.6 + sine(f * 1.51, t) * 0.4
        out.append(s * math.exp(-38 * t) * env(i, dur, attack=0.002))
    return out
made.append(("ask-pebble",  two(pebble(520, 0.12), pebble(700, 0.16), 0.055)))
made.append(("done-pebble", two(pebble(700, 0.12), pebble(460, 0.20), 0.055)))

for name, buf in made:
    p = write(name, buf, gain=0.42)
    print(f"{os.path.basename(p):18s} {len(buf)/SR*1000:6.0f} ms")
