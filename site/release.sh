#!/bin/sh
# release.sh builds the binaries install.sh downloads and publishes them as
# the GitHub release for one tag:
#
#     sh site/release.sh -n v0.4.2   # build dist/ and list it, publish nothing
#     sh site/release.sh v0.4.2      # build dist/, push the tag, create the release
#
# Run it from a CLEAN tree with HEAD at the tag: go build stamps the
# revision, and a dirty tree stamps vcs.modified=true and the wrong SHA,
# which is the same rule the installed binary follows. The four targets are
# the platforms internal/host has a pty file for, darwin and linux on arm64
# and amd64, built with cgo off so the build runs anywhere.
#
# install.sh fetches <release>/latest/download/xscapes-<os>-<arch>, so the
# asset names here are part of the contract.
set -eu
dry=no
if [ "${1:-}" = -n ]; then
  dry=yes
  shift
fi
tag=${1:?usage: sh site/release.sh [-n] vX.Y.Z}
root=$(git rev-parse --show-toplevel)
cd "$root"
if [ -n "$(git status --porcelain)" ]; then
  echo "release.sh: the tree is not clean; commit first" >&2
  exit 1
fi
at=$(git describe --exact-match --tags HEAD 2>/dev/null || true)
if [ "$at" != "$tag" ]; then
  echo "release.sh: HEAD is tagged '${at:-nothing}', not $tag; tag first" >&2
  exit 1
fi
rm -rf dist
mkdir dist
for t in darwin/arm64 darwin/amd64 linux/amd64 linux/arm64; do
  GOOS=${t%/*} GOARCH=${t#*/} CGO_ENABLED=0 go build -trimpath -o "dist/xscapes-${t%/*}-${t#*/}" .
done
(cd dist && shasum -a 256 xscapes-* >checksums.txt)
ls -l dist
go version -m dist/xscapes-darwin-arm64 | grep -E 'vcs\.(revision|modified)'
if [ $dry = yes ]; then
  exit 0
fi
# The active gh account is host-global and flips back to another login
# between commands; switch immediately before every use.
gh auth switch --user donlucasx
git push origin "refs/tags/$tag"
gh release create "$tag" dist/* --verify-tag --title "$tag" \
  --notes "Binaries for the install line: \`curl -fsSL https://donlucasx.github.io/xscapes/install.sh | sh\`"
echo "released $tag: https://github.com/donlucasx/xscapes/releases/tag/$tag"
