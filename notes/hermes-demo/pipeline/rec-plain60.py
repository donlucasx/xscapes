# records plain `hermes` starting at 124x44 (its full welcome screen), then exits; no prompt is sent
import os, pty, time, select, fcntl, termios, struct, json, base64, sys
OUT = sys.argv[1]; os.makedirs(OUT, exist_ok=True)
env = dict(os.environ, TERM="xterm-256color", COLUMNS="124", LINES="60"); env.pop("TERM_PROGRAM", None)
pid, fd = pty.fork()
if pid == 0:
    os.chdir(os.path.join(os.path.dirname(os.path.abspath(__file__)), "work")); os.execvpe("hermes", ["hermes"], env)
fcntl.ioctl(fd, termios.TIOCSWINSZ, struct.pack("HHHH", 60, 124, 0, 0))
T0 = time.time(); f = open(os.path.join(OUT, "tty.jsonl"), "w")
def pump(s):
    end = time.time() + s
    while time.time() < end:
        r, _, _ = select.select([fd], [], [], 0.05)
        if r:
            try: d = os.read(fd, 65536)
            except OSError: return
            if not d: return
            f.write(json.dumps({"t": round(time.time() - T0, 4), "b": base64.b64encode(d).decode()}) + "\n")
pump(14)
os.write(fd, b"/exit\r"); pump(3)
try: os.kill(pid, 15)
except Exception: pass
open(os.path.join(OUT, "marks.jsonl"), "w").write(json.dumps({"kind": "start", "t": 0}) + "\n")
print("ok")
