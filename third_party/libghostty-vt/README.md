# libghostty-vt

Ghostty's terminal engine as a C library, used by `internal/gvt` (build tag `ghosttyvt`).

`include/` is committed. The static library is NOT (12.6 MB a platform); build it:

    git clone --depth 1 --branch vt-1.0.0 https://github.com/ghostty-org/ghostty
    cd ghostty && zig build -Demit-lib-vt -Doptimize=ReleaseFast
    cp zig-out/lib/libghostty-vt.a <repo>/third_party/libghostty-vt/lib/darwin-arm64/

Built here from tag vt-1.0.0 (0fa3d577a5614150ff8e2c152de5fb9d5c6d2dec).
Then: `go build -tags ghosttyvt .`
