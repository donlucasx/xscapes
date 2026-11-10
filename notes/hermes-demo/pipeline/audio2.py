"""audio2.py <timeline.json> <beds.json> <out.wav>: real recordings (CC0, BigSoundBank)
for the beds and the keys, the SHIPPED ask/done cues at the recorded moments."""
import json, sys, subprocess, struct, random, math, wave
R = 44100
SRC = "/private/tmp/claude-501/-Users-lucasgarzoli-Documents-claude/30c97ba9-3828-4cc5-bc2a-533746fce114/scratchpad/sfx-src/"
SND = "/Users/lucasgarzoli/Documents/claude/xscapes/internal/notify/sounds/"
def load(path, t0=0.0, dur=None):
    a = ["ffmpeg", "-v", "error", "-ss", str(t0), "-i", path]
    if dur: a += ["-t", str(dur)]
    raw = subprocess.run(a + ["-ac", "1", "-ar", str(R), "-f", "s16le", "-"], capture_output=True).stdout
    return [v / 32768 for v in struct.unpack(f"<{len(raw)//2}h", raw)]
def rms(x): return math.sqrt(sum(v * v for v in x) / max(1, len(x)))
def smooth(u): u = max(0.0, min(1.0, u)); return u * u * (3 - 2 * u)
tl = json.load(open(sys.argv[1])); beds = json.load(open(sys.argv[2])); out = sys.argv[3]
n = int(tl["total"] * R); buf = [0.0] * n
for b in beds:
    # b: file, src (s into the file), v0, v1, level (rms), fin, fout, env: [[v, gain], ...] optional
    x = load(SRC + b["file"], b["src"], b["v1"] - b["v0"] + 0.1)
    g = min(b["level"] / (rms(x) or 1), b.get("peak", 0.35) / (max(abs(v) for v in x) or 1))
    print("bed", b["file"], "gain", round(g, 2), "rms", round(rms(x) * g, 4))
    env = b.get("env")
    for i, v in enumerate(x):
        t = b["v0"] + i / R
        if t >= b["v1"]: break
        j = int(t * R)
        if j >= n: break
        e = min(smooth((t - b["v0"]) / b["fin"]) if b["fin"] else 1, smooth((b["v1"] - t) / b["fout"]) if b["fout"] else 1)
        if env:
            for k in range(len(env) - 1):
                if env[k][0] <= t < env[k + 1][0]:
                    u = (t - env[k][0]) / (env[k + 1][0] - env[k][0]); e *= env[k][1] + (env[k + 1][1] - env[k][1]) * u; break
            else:
                e *= env[-1][1] if t >= env[-1][0] else env[0][1]
        buf[j] += v * g * e
# keys: single strokes cut from a real slow-typing recording at their onsets
k = load(SRC + "1733.flac")
win = int(0.005 * R); hits = []; last = -R
en = [rms(k[i:i + win]) for i in range(0, len(k) - win, win)]
floor = sorted(en)[len(en) // 2]
for idx in range(1, len(en)):
    if en[idx] > max(0.02, floor * 8) and en[idx - 1] < en[idx] * 0.5 and idx * win - last > int(0.12 * R):
        s = max(0, idx * win - int(0.004 * R)); hits.append(k[s:s + int(0.11 * R)]); last = idx * win
hits = [h for h in hits if len(h) > 100]
peak = lambda h: max(abs(v) for v in h) or 1
hits = [[v / peak(h) for v in h] for h in hits]
rnd = random.Random(7)
for t in tl["keys"]:
    h = hits[rnd.randrange(len(hits))]; j0 = int(t * R); gk = 0.32 + rnd.random() * 0.08
    for i, v in enumerate(h):
        fade = 1 - i / len(h)
        if 0 <= j0 + i < n: buf[j0 + i] += v * gk * fade
# ticks: the Hermes screen drawing itself. Muffled (one-pole lowpass), quiet, even:
# a soft patter, not a typist.
for t in tl.get("ticks", [])[::2]:
    h = hits[rnd.randrange(len(hits))]; j0 = int(t * R); gk = 0.09 + rnd.random() * 0.03; y = 0.0
    for i, v in enumerate(h):
        y += 0.18 * (v - y)
        if 0 <= j0 + i < n: buf[j0 + i] += y * gk * (1 - i / len(h))
def wavread(p):
    with wave.open(p) as w:
        f = w.readframes(w.getnframes()); return [v / 32768 for v in struct.unpack(f"<{len(f)//2}h", f)]
ask, done = wavread(SND + "ask.wav"), wavread(SND + "done.wav")
for c in tl["cues"]:
    s = ask if c["kind"] == "ask" else done; j0 = int(c["v"] * R)
    for i, v in enumerate(s):
        if 0 <= j0 + i < n: buf[j0 + i] += v * 0.8
pk = max(abs(v) for v in buf) or 1
if pk > 0.95: buf = [v * 0.95 / pk for v in buf]
with wave.open(out, "wb") as w:
    w.setnchannels(2); w.setsampwidth(2); w.setframerate(R)
    w.writeframes(b"".join(struct.pack("<hh", int(v * 32767), int(v * 32767)) for v in buf))
print("audio2", tl["total"], "keys", len(tl["keys"]), "key samples", len(hits), "cues", [(c["kind"], c["v"]) for c in tl["cues"]], "peak", round(pk, 2))
