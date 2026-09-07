#!/bin/bash
# drive-dl.sh -- runs widthprobe's DL arm in a Terminal window this script opens,
# and prints the read-back. His OK, 2026-09-07: "Yes, in its own window".
#
# The window is resolved by the TTY of the tab `do script` returns, never by
# `front window`: on 2026-09-03 a `front window` driver typed into HIS Claude
# session. Nothing here sends keystrokes, and every command names the window by
# the id resolved from that tty.
#
# The question: a stale strip at columns 144-145 of a 143-column terminal is the
# window's own inset holding RETAINED cells. Erase does not reach them
# (width-audit item 2), nor does a repaint at the narrow width (item 3). DL is
# the last mechanism that could -- it replaces rows rather than erasing cells.
#
#   GREEN: rows 10-20 come back with NO tail past the narrow width, while the
#          control rows still have theirs.
#   RED:   rows 10-20 keep their tails too. Then it is not fixable from inside
#          the app, and the 123x55 verdict stands for this too.
set -uo pipefail

OUT="${1:?usage: drive-dl.sh <output-dir>}"
mkdir -p "$OUT"
BIN="$OUT/widthprobe-dl"
go build -o "$BIN" ./notes/widthprobe || exit 1

WIDE=120
NARROW=74

# 1. Open a window running the probe, and resolve it by tty.
read -r TTY WINID <<<"$(osascript <<'APPLESCRIPT'
tell application "Terminal"
    set newTab to do script ""
    set theTty to tty of newTab
    set theId to 0
    repeat with w in windows
        repeat with tb in tabs of w
            if tty of tb is theTty then set theId to id of w
        end repeat
    end repeat
    return theTty & " " & theId
end tell
APPLESCRIPT
)"

if [ -z "${WINID:-}" ] || [ "$WINID" = "0" ]; then
    echo "FATAL: could not resolve the window from its tty; refusing to touch any window" >&2
    exit 1
fi
echo "probe window: tty=$TTY id=$WINID"

# Assert the resolved window is the one we opened and nothing else, by writing
# ONLY to that tty. No keystrokes, no `front window`, no System Events.
run_in_probe() { osascript -e "tell application \"Terminal\" to do script \"$1\" in (first tab of (first window whose id is $WINID))" >/dev/null; }
set_size() { osascript -e "tell application \"Terminal\" to set number of columns of (first window whose id is $WINID) to $1" \
                       -e "tell application \"Terminal\" to set number of rows of (first window whose id is $WINID) to $2" >/dev/null; }
read_back() { osascript -e "tell application \"Terminal\" to get history of (first tab of (first window whose id is $WINID))"; }

set_size "$WIDE" 40
run_in_probe "exec '$BIN'"

# The probe paints at t=3 (at the WIDE size), fires its arms at t=8.
sleep 5;  set_size "$NARROW" 40; echo "narrowed to $NARROW"
sleep 6;  set_size "$WIDE"   40; echo "widened back to $WIDE"
sleep 2;  read_back > "$OUT/readback.txt"; echo "read back $(wc -l < "$OUT/readback.txt") lines"

osascript -e "tell application \"Terminal\" to close (first window whose id is $WINID)" >/dev/null 2>&1
echo "closed the probe window"
