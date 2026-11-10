# rec.py <outdir> : records one real `xscapes hermes` session in a hidden pty.
# Writes out/tty.jsonl ({t, b} base64 output chunks), out/marks.jsonl
# (keys, scape switches, events seen), all t in seconds from the start.
import os, pty, sys, time, select, fcntl, termios, struct, json, base64, random, glob, subprocess

OUT = sys.argv[1]; os.makedirs(OUT, exist_ok=True)
HERE = os.path.dirname(os.path.abspath(__file__))
COLS, ROWS = 124, 44
P1 = "lets get busy. Run a multi agent showcase of xscapes"
P2 = os.environ.get("DEMO_P2", "count the lines of Go in this repo, per top-level folder, with one shell command")
P3 = os.environ.get("DEMO_P3", "think it over first, then create an empty file demo.sh here and chmod 777 it")

tty = open(os.path.join(OUT, "tty.jsonl"), "w")
marks = open(os.path.join(OUT, "marks.jsonl"), "w")
T0 = time.time()
def now(): return time.time() - T0
def mark(kind, **kw):
    kw.update(t=round(now(), 3), kind=kind); marks.write(json.dumps(kw) + "\n"); marks.flush()
    print(f"{kw['t']:8.2f} {kind} {kw.get('text','') or kw.get('ch','')}", flush=True)

env = dict(os.environ)
env.update(TERM="xterm-256color", ZDOTDIR=os.path.join(HERE, "zd"), COLUMNS=str(COLS), LINES=str(ROWS))
env.pop("TERM_PROGRAM", None); env["XSCAPES_SILENT"] = "1"; env["XSCAPES_SCAPE"] = os.environ["DEMO_SCAPE"]
pid, fd = pty.fork()
if pid == 0:
    os.chdir(os.path.join(HERE, "work"))
    os.execvpe("zsh", ["zsh", "-i"], env)
fcntl.ioctl(fd, termios.TIOCSWINSZ, struct.pack("HHHH", ROWS, COLS, 0, 0))
os.kill(pid, 28)  # SIGWINCH

last_out = [now()]
def pump(secs):
    end = time.time() + secs
    while True:
        left = end - time.time()
        if left <= 0: return
        r, _, _ = select.select([fd], [], [], min(left, 0.05))
        if r:
            try: d = os.read(fd, 65536)
            except OSError: return
            if not d: return
            tty.write(json.dumps({"t": round(now(), 4), "b": base64.b64encode(d).decode()}) + "\n")
            last_out[0] = now()
        poll_events()

RUN = os.path.expanduser("~/.config/xscapes/run")
launched = time.time()
old = set(glob.glob(RUN + "/*.recv.jsonl"))
seen = {}  # file -> offset
events = []
def poll_events():
    for f in glob.glob(RUN + "/*.recv.jsonl"):
        if f in old: continue
        off = seen.get(f, 0)
        with open(f) as h:
            h.seek(off); data = h.read(); seen[f] = h.tell()
        for line in data.splitlines():
            try: e = json.loads(line)
            except Exception: continue
            e["_t"] = round(e["ts"] / 1000 - T0, 3)
            events.append(e)
            mark("event", ev=e["kind"], at=e["_t"], text=e["kind"] + " " + (e.get("text") or e.get("tool") or ""), op=e.get("op", ""), file=os.path.basename(f))

def typ(s, cps=13):
    for ch in s:
        os.write(fd, ch.encode()); mark("key", ch=ch)
        pump(max(0.03, random.gauss(1 / cps, 0.025)))
def enter():
    os.write(fd, b"\r"); mark("key", ch="\n")

def wait_quiet(q, mn, mx):
    st = time.time()
    while time.time() - st < mx:
        pump(0.25)
        if time.time() - st > mn and now() - last_out[0] > q: return
def wait_turn(mx, approve_after=3.0):
    """Wait for done with no subagent traffic for a while; approve asks."""
    st = time.time(); n0 = len(events); asked_at = None
    while time.time() - st < mx:
        pump(0.25)
        new = events[n0:]
        ask = [e for e in new if e["kind"] == "needs_input"]
        ans = [e for e in new if e["kind"] == "answered"]
        if len(ask) > len(ans):
            if asked_at is None: asked_at = time.time()
            elif time.time() - asked_at > approve_after:
                mark("approve"); enter(); asked_at = None; pump(1.5)
        dones = [e for e in new if e["kind"] == "done"]
        subs = [e for e in new if e["kind"].startswith("sub_")]
        lastsub = max([e["_t"] for e in subs], default=0)
        if dones and now() - max(dones[-1]["_t"], lastsub) > 12:
            return True
    mark("timeout"); return False

def scape(name):
    subprocess.run(["xscapes", "scape", name], capture_output=True); mark("scape", text=name)

SC = os.environ["DEMO_SCAPE"]; PR = os.environ["DEMO_PROMPT"]; IDLE = float(os.environ.get("DEMO_IDLE", "4"))
pump(1.0)
os.write(fd, b"clear\r"); pump(1.5)
mark("start")
pump(1.8)
typ("xscapes hermes", 9); pump(0.4); enter()
pump(9); mark("ready")
pump(IDLE)
typ(PR); pump(0.3); enter(); mark("p")
wait_turn(480, approve_after=6.0); mark("pdone")
# async children report back after the parent's done: hold until 25 s of no events
st = time.time()
while time.time() - st < 240:
    pump(0.5)
    if events and now() - events[-1]["_t"] > 25: break
mark("end")
os.write(fd, b"/exit\r"); pump(3)
try: os.kill(pid, 15)
except Exception: pass
print("done", now())
