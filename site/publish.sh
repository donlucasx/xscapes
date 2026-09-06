#!/bin/sh
# publish.sh puts the submission page online: site/index.html and its clips
# become the whole of an orphan gh-pages branch, force-pushed, which GitHub
# Pages serves at https://donlucasx.github.io/xscapes/ (Settings -> Pages ->
# branch gh-pages, folder /). No workflow file, because the stored GitHub
# credential has no `workflow` scope and a push carrying one is refused.
#
# Run it from the repo root after `go run . -site site && python3
# site/make-gifs.py`. The frame pages under anim/ stay home; only the page
# and the GIFs travel.
set -eu
root=$(git rev-parse --show-toplevel)
cd "$root"
stage=$(mktemp -d)
trap 'rm -rf "$stage"' EXIT
mkdir -p "$stage/anim"
cp site/index.html "$stage/"
cp site/anim/*.gif "$stage/anim/"
touch "$stage/.nojekyll"
git -C "$stage" init -q -b gh-pages
git -C "$stage" add -A
git -C "$stage" -c user.name="$(git config user.name)" -c user.email="$(git config user.email)" \
  commit -q -m "site: $(git rev-parse --short HEAD)"
git -C "$stage" push -q --force "$(git remote get-url origin)" gh-pages:gh-pages
echo "pushed gh-pages from $(git rev-parse --short HEAD): https://donlucasx.github.io/xscapes/"
