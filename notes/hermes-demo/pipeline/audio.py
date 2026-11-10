"""audio.py <timeline.json> <out.wav>: key clicks at each typed key, the shipped
ask.wav / done.wav at the recorded moments, a quiet bed per scape (wind on the
vista, the sea's wash on the shore, filtered rain on the harbour)."""
import json, sys, importlib.util, random
spec = importlib.util.spec_from_file_location("a", "/Users/lucasgarzoli/Documents/claude/xscapes/notes/trailer/audio.py")
A = importlib.util.module_from_spec(spec); spec.loader.exec_module(A)
R = A.RATE
tl = json.load(open(sys.argv[1])); out = sys.argv[2]
n = int(tl["total"] * R); buf = [0.0] * n
snd = "/Users/lucasgarzoli/Documents/claude/xscapes/internal/notify/sounds/"
ask, done = A.read_wav(snd + "ask.wav"), A.read_wav(snd + "done.wav")
def rain(n, seed=23):
    xs = A.noise(n, seed); hp = [0.0] * n; prev = 0.0; y = 0.0
    for i, x in enumerate(xs):  # a crude high-pass: rain is hiss, not rumble
        y = 0.97 * (y + x - prev); prev = x; hp[i] = y
    return A.lowpass(hp, 6000)
beds = {"vista": A.wind(n), "shore": A.sea(n), "harbour": rain(n)}
bed_rms = 0.02
for name, b in beds.items():
    spans = [s for s in tl["scapes"] if s["name"] == name]
    if not spans: continue
    seg = b[int(spans[0]["v0"] * R):int(spans[0]["v1"] * R)] or b
    g = bed_rms / (A.rms(seg) or 1)
    for s in spans:
        for i in range(int(s["v0"] * R), min(n, int(s["v1"] * R))):
            t = i / R
            buf[i] += b[i] * g * min(A.ramp(t, s["v0"], 0.8), 1 - A.ramp(t, s["v1"] - 0.8, 0.8))
ticks = [A.tick(seed=k) for k in range(8)]
rnd = random.Random(4)
for k in tl["keys"]:
    A.place(buf, ticks[rnd.randrange(8)], k, 0.9 + rnd.random() * 0.2)
for c in tl["cues"]:
    A.place(buf, ask if c["kind"] == "ask" else done, c["v"], 0.85)
A.write(out, buf)
print("audio", tl["total"], "keys", len(tl["keys"]), "cues", tl["cues"])
