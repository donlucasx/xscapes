// Command wetsand answers one question: what does the 256 cube do with the
// waterline, and where does it stop being able to draw wet sand at all?
//
// He photographed the answer on 2026-09-09 at 111 columns -- neutral grey
// blocks along the shore, five cells wide, under the litter. Measured at that
// moment: the palette asked for (93,76,66), a warm brown, and the terminal drew
// (78,78,78), a flat grey, right beside dry sand at an exact (175,135,95).
//
// THE CAUSE IS NOT THE QUANTISER. Index256Keeping picks the genuinely nearest
// entry; the cube simply has no warm colour down there. Its channel levels are
// 0, 95, 135, 175, 215, 255, so below 95 in green and blue there is only zero:
// the darkest entry that keeps a warm ordering R>G>=B is 135/95/0 at luma 96,
// and the next is 135/95/95 at 107. Under luma 96 the only warm entries left
// are the pure reds. The grey ramp has steps of ten all the way down, so for
// any dark near-warm colour a grey is always nearer.
//
// AND IT IS NOT FIXABLE WITH A RAMP. Painting the waterline through
// term.NewRamp(WetSand, SeaNear) the way the sky and the sea are painted was
// tried and MEASURED WORSE -- 1.42% of daylight waterline cells neutral before,
// 3.12% after -- because a tan-to-blue path crosses the same dead zone and the
// path's own tones land in it.
//
// What is left is the palette: it is asking for a colour that does not exist.
//
//	go run ./notes/wetsand
package main

func main() { run() }
