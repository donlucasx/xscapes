#!/bin/bash
# Audition the candidate cues. Each pair plays ASK then DONE, a second apart,
# the way you would actually hear them: a question, then the answer.
#
#   bash notes/s28-sound/audition.sh          # all four families
#   bash notes/s28-sound/audition.sh droplet  # just one
#   bash notes/s28-sound/audition.sh shipped  # what xscapes plays TODAY
cd "$(dirname "$0")"
one() {
  printf '\n  %-9s  ask ' "$1"; afplay "ask-$1.wav"; sleep 0.55
  printf '... done '; afplay "done-$1.wav"; sleep 0.9
}
if [ "$1" = "shipped" ]; then
  printf '\n  %-9s  ask ' "SHIPPED"; afplay /System/Library/Sounds/Glass.aiff; sleep 0.55
  printf '... done '; afplay /System/Library/Sounds/Submarine.aiff; echo; exit
fi
if [ -n "$1" ]; then one "$1"; echo; exit; fi
echo "  (each: ask, then done)"
for f in droplet shell tide pebble; do one "$f"; done
printf '\n  %-9s  ask ' "shipped"; afplay /System/Library/Sounds/Glass.aiff; sleep 0.55
printf '... done '; afplay /System/Library/Sounds/Submarine.aiff
echo; echo
