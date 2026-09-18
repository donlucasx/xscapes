#!/usr/bin/env python3
"""The trailer's sound, stdlib only like notes/s34-sound/make.py.

  python3 audio.py opener <opener.json> <out.wav>
  python3 audio.py film   <film.json> <sounds dir> <out.wav>

The opener carries only the six keystrokes: a key is a few milliseconds of
broadband noise shaped by a short body resonance, one soft tick per letter,
placed where the letter appears.

The film adds the product's own cues at the product's own moments (the
shipped ask.wav and done.wav, read from internal/notify/sounds, never
re-synthesised) and a quiet bed under the scapes: wind on the vista, the sea's
wash on the shore, crossfaded at the cut, in with the dissolve and out with
the end card. His ruling (2026-09-17): the cues plus a quiet bed, no VO.
"""
import json
import math
import random
import struct
import sys
import wave

RATE = 44100


def silence(seconds):
    return [0.0] * int(RATE * seconds)


def tick(gain=0.22, seed=1):
    """A soft key: 2 ms of noise through a decaying body at ~1.9 kHz, then a
    faint lower thud. Peak-normalised to gain."""
    rnd = random.Random(seed)
    n = int(RATE * 0.045)
    out = [0.0] * n
    for i in range(int(RATE * 0.0025)):
        out[i] += (rnd.random() * 2 - 1) * math.exp(-i / (RATE * 0.0007))
    f, tau = 1900.0, 0.0045
    for i in range(n):
        t = i / RATE
        out[i] += 0.55 * math.sin(2 * math.pi * f * t) * math.exp(-t / tau)
    f2, tau2 = 320.0, 0.012
    for i in range(n):
        t = i / RATE
        out[i] += 0.25 * math.sin(2 * math.pi * f2 * t) * math.exp(-t / tau2)
    peak = max(abs(v) for v in out) or 1.0
    return [v / peak * gain for v in out]


def place(buf, sample, at, gain=1.0):
    i0 = int(at * RATE)
    for i, v in enumerate(sample):
        j = i0 + i
        if 0 <= j < len(buf):
            buf[j] += v * gain


def read_wav(path):
    """A shipped cue: 16-bit mono 44.1k (sounds_test.go holds that), as floats."""
    with wave.open(path, "rb") as w:
        assert w.getsampwidth() == 2 and w.getnchannels() == 1 and w.getframerate() == RATE, path
        raw = w.readframes(w.getnframes())
    return [s / 32768.0 for s in struct.unpack("<%dh" % (len(raw) // 2), raw)]


def write(path, buf):
    with wave.open(path, "wb") as w:
        w.setnchannels(2)
        w.setsampwidth(2)
        w.setframerate(RATE)
        frames = bytearray()
        for v in buf:
            v = max(-1.0, min(1.0, v))
            s = struct.pack("<h", int(v * 32767))
            frames += s + s
        w.writeframes(bytes(frames))


# ---- the bed ------------------------------------------------------------------

def lowpass(xs, hz):
    a = 1 - math.exp(-2 * math.pi * hz / RATE)
    y, out = 0.0, [0.0] * len(xs)
    for i, x in enumerate(xs):
        y += a * (x - y)
        out[i] = y
    return out


def noise(n, seed):
    rnd = random.Random(seed)
    return [rnd.random() * 2 - 1 for _ in range(n)]


def rms(xs):
    return math.sqrt(sum(v * v for v in xs) / max(1, len(xs)))


def wind(n, seed=21):
    """Air over the meadow: noise low-passed to a breath, its level breathing
    on two slow, unrelated cycles so the gusts never repeat inside the clip."""
    x = lowpass(noise(n, seed), 520)
    x = lowpass(x, 900)
    out = [0.0] * n
    for i in range(n):
        t = i / RATE
        g = 0.55 + 0.45 * (0.5 + 0.5 * math.sin(2 * math.pi * 0.09 * t + 1.1)) * (0.6 + 0.4 * math.sin(2 * math.pi * 0.23 * t + 2.6))
        out[i] = x[i] * g
    return out


def sea(n, seed=22, period=8.5):
    """The wash: a body of pink-ish noise under a swell that rises slowly and
    breaks, with a thread of hiss riding each crest for the foam."""
    body = lowpass(noise(n, seed), 1100)
    deep = lowpass(noise(n, seed + 1), 160)
    raw = noise(n, seed + 2)
    hiss = [raw[i] - v for i, v in enumerate(lowpass(raw, 2400))]
    out = [0.0] * n
    for i in range(n):
        t = i / RATE
        ph = (t / period) % 1.0
        # asymmetric: a slow rise (0..0.62) and a quicker fall
        w = (ph / 0.62) if ph < 0.62 else (1 - (ph - 0.62) / 0.38)
        w = max(0.0, w) ** 1.7
        out[i] = (0.7 * body[i] + 0.9 * deep[i]) * (0.25 + 0.75 * w) + 0.35 * hiss[i] * (w ** 2.2)
    return out


def ramp(t, t0, dur):
    if dur <= 0:
        return 1.0 if t >= t0 else 0.0
    u = (t - t0) / dur
    u = max(0.0, min(1.0, u))
    return u * u * (3 - 2 * u)


# ---- the two scores ----------------------------------------------------------------

def opener(spec_path, out):
    spec = json.load(open(spec_path))
    buf = silence(spec["end"])
    at = spec["letters"]["start"]
    for i, g in enumerate(spec["letters"]["gaps"]):
        at += g
        place(buf, tick(gain=0.20 + 0.03 * (i % 2), seed=11 + i), at)
    write(out, buf)
    print(f"opener: {spec['end']} s, {len(spec['letters']['gaps'])} keys, peak {max(abs(v) for v in buf):.2f}")


def film(tl_path, sounds, out, bed_rms=0.022, cue_gain=0.85):
    tl = json.load(open(tl_path))
    n = int(RATE * tl["end"])
    buf = silence(tl["end"])
    # the keys, where the letters appear
    for i, at in enumerate(tl["keys"]):
        place(buf, tick(gain=0.20 + 0.03 * (i % 2), seed=11 + i), at)
    # the bed: wind from the dissolve, the sea from the cut, out with the card
    m0, cut, end = tl["montageStart"], tl["cut"][0], tl["montageEnd"]
    w, s = wind(n), sea(n)
    gw, gs = bed_rms / rms(w[int(m0 * RATE):int(cut * RATE)]), bed_rms / rms(s[int(cut * RATE):int(end * RATE)])
    for i in range(int(m0 * RATE), n):
        t = i / RATE
        inn = ramp(t, m0, 1.5)
        outt = 1 - ramp(t, end, 1.5)
        x = ramp(t, cut - 0.5, 1.0)  # wind -> sea across the cut
        buf[i] += (w[i] * gw * (1 - x) + s[i] * gs * x) * inn * outt
    # the cues, the product's own files at the product's own moments
    ask, done = read_wav(f"{sounds}/ask.wav"), read_wav(f"{sounds}/done.wav")
    for at in tl["asks"]:
        place(buf, ask, at, cue_gain)
    for at in tl["dones"]:
        place(buf, done, at, cue_gain)
    if "catDoneAt" in tl:  # the cat's finish face on the end card
        place(buf, done, tl["catDoneAt"], cue_gain)
    for i, at in enumerate(tl.get("endKeys", [])):  # the URL typed under the mark
        place(buf, tick(gain=0.16 + 0.03 * (i % 2), seed=41 + i), at)
    write(out, buf)
    print(f"film: {tl['end']} s, bed rms {bed_rms}, cues at asks {tl['asks']} dones {tl['dones']}, peak {max(abs(v) for v in buf):.2f}")


if __name__ == "__main__":
    a = sys.argv[1:]
    if len(a) == 3 and a[0] == "opener":
        opener(a[1], a[2])
    elif len(a) == 4 and a[0] == "film":
        film(a[1], a[2], a[3])
    else:
        print(__doc__)
        sys.exit(2)
