package scape

import "github.com/donlucasx/xscapes/internal/term"

// Test-only windows onto the shore's internals, for tests in this package that
// have to measure what the renderer actually painted from.
//
// They lived in zzscratch_test.go, which is a scratch file that reached main by
// way of `git add -A` -- the same accident as zzambientprobe_test.go. Moved
// here on 2026-09-13 because TestSittersAndTheTide now depends on
// LastEdgeExport, so deleting the scratch file would have broken the suite in a
// way its name promises is safe.

func writeBandColorExport(s *Shore) term.RGB { return writeBandColor(s.pal) }

func (s *Shore) LastEdgeExport() []float64 { return s.lastEdge }
