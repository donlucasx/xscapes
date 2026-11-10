#!/bin/sh
# assemble2.sh <framesDir> <audio.wav> <name> : the X upload (H.264 High, keyframe every 2 s) and a
# REVIEW copy that scrubs smoothly (a keyframe every 0.5 s, no B-frames), both 1920x1080 30 fps.
set -e
d="$1"; a="$2"; n="$3"; D=/Users/lucasgarzoli/Documents/claude/xscapes/previews/hermes-demo
ffmpeg -y -v error -framerate 30 -i "$d/f%05d.png" -i "$a" -c:v libx264 -profile:v high -preset slow -crf 17 -pix_fmt yuv420p -g 60 \
  -c:a aac -b:a 192k -ar 44100 -ac 2 -shortest -movflags +faststart "$D/$n.mp4"
ffmpeg -y -v error -framerate 30 -i "$d/f%05d.png" -i "$a" -c:v libx264 -profile:v high -preset fast -crf 18 -pix_fmt yuv420p -g 15 -keyint_min 15 -bf 0 -sc_threshold 0 \
  -c:a aac -b:a 192k -ar 44100 -ac 2 -shortest -movflags +faststart "$D/$n-review.mov"
ls -la "$D/$n.mp4" "$D/$n-review.mov"
