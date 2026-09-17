#!/bin/sh
# publish.sh puts the submission page online: site/index.html and its clips
# become the whole of an orphan gh-pages branch, force-pushed, which GitHub
# Pages serves at https://donlucasx.github.io/xscapes/ (Settings -> Pages ->
# branch gh-pages, folder /). No workflow file, because the stored GitHub
# credential has no `workflow` scope and a push carrying one is refused.
#
# Run it from the repo root after `go run . -site site`. Nothing else travels:
# the page is self-contained.
set -eu
root=$(git rev-parse --show-toplevel)
cd "$root"
stage=$(mktemp -d)
trap 'rm -rf "$stage"' EXIT
# The page is ONE file now. Every animation is embedded as text rather than
# screenshotted into a GIF, so there is nothing beside it to copy: 5.1MB of
# clips became 355KB of markup, and anim/ is build scratch. The one thing
# that travels with it is the share card, og.png, which the page's Open
# Graph tags point at by absolute URL (a card scraper needs an image URL,
# and the Commons proxy serves only the page).
cp site/index.html site/og.png "$stage/"
touch "$stage/.nojekyll"
git -C "$stage" init -q -b gh-pages
git -C "$stage" add -A
git -C "$stage" -c user.name="$(git config user.name)" -c user.email="$(git config user.email)" \
  commit -q -m "site: $(git rev-parse --short HEAD)"
git -C "$stage" push -q --force "$(git remote get-url origin)" gh-pages:gh-pages
echo "pushed gh-pages from $(git rev-parse --short HEAD): https://donlucasx.github.io/xscapes/"
