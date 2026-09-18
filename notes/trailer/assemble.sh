#!/bin/sh
# assemble.sh <framesDir> <audio.wav|none> <out.mp4> [fps]
# X's own spec: MP4, H.264 (High), yuv420p, AAC-LC stereo, 30 fps, faststart.
# crf 17 lands 1080p at roughly 8-12 Mbps on this material, inside X's
# recommended range; X re-encodes anyway.
set -e
dir="$1"; audio="$2"; out="$3"; fps="${4:-30}"
if [ "$audio" = "none" ]; then
  ffmpeg -y -v error -framerate "$fps" -i "$dir/f%05d.png" \
    -c:v libx264 -profile:v high -preset slow -crf 17 -pix_fmt yuv420p -r "$fps" \
    -movflags +faststart "$out"
else
  ffmpeg -y -v error -framerate "$fps" -i "$dir/f%05d.png" -i "$audio" \
    -c:v libx264 -profile:v high -preset slow -crf 17 -pix_fmt yuv420p -r "$fps" \
    -c:a aac -b:a 192k -ar 44100 -ac 2 -shortest \
    -movflags +faststart "$out"
fi
ffprobe -v error -show_entries stream=codec_name,width,height,r_frame_rate,duration,bit_rate,channels -of compact "$out"
