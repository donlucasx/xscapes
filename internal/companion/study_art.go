package companion

import "github.com/donlucasx/xscapes/internal/term"

// Study doors. These exist so a DESIGN ROUND can render candidate art through
// the same Draw path the product uses -- the same breathing, the same tail, the
// same mirror, the same eye pass -- instead of re-implementing all of it in a
// study and then judging a picture the product would never draw.
//
// That replication is not a hypothetical risk here: notes/s28-starfun copied
// the constellation's alpha as a constant the shipped code had already deleted,
// and rendered 25 of 32 stars differently from the product while reporting
// itself as "today".
//
// Neither door changes what ships. SetAskArt's field is nil by default and Draw
// falls through to exactly the behaviour it has now.

// SetBodyArt replaces the seated body bitmap.
func (c *Cat) SetBodyArt(rows []string) { c.body = ParseBitmap(rows) }

// SetAskArt gives the cat a body of its own for the NeedsYou pose.
//
// It is nil today, which is the defect this is drafted against: measured at
// HEAD 7060844, the cat's Working and NeedsYou poses differ by ZERO cells of
// body, so the ask is carried entirely by one eye glyph changing weight from
// 'o' to 'O'. The crab got a shape difference in session 28 (a 2x2 ring against
// a 2x2 bead) and the cat never did.
func (c *Cat) SetAskArt(rows []string) { c.ask = ParseBitmap(rows) }

// AskArt says whether an ask pose has been set, so a study can label a frame
// honestly rather than asserting which arm it is rendering.
func (c *Cat) AskArt() bool { return c.ask != nil }

// SetNoRim suppresses the cleared ring. Study-only, and it exists so a probe
// can diff the ring against no ring on an otherwise identical frame -- the
// check that says a golden-hash change is the change that was intended.
func (c *Cat) SetNoRim(v bool) { c.noRim = v }

// CrabWorkingArt is the crab's shipped working body, upper block then lower,
// exported so a study can diff a candidate crab pose against what actually
// ships. Without it a study has only CatBody to compare against, and a crab
// scored against the cat's bitmap produces a number that means nothing.
func CrabWorkingArt() []string {
	return append(append([]string{}, crabWork...), crabLower...)
}

// StudyNearEyeAlert overrides the near pose's NeedsYou eye, for a design round
// only. nil is the shipped ring.
//
// A package variable rather than a field because nearEye is a free function and
// the alternative was threading a parameter through drawCrabNear, nearPoseFor
// and every caller to serve a study. It is written once by a page generator and
// read on the same goroutine; nothing in the product ever sets it.
var (
	StudyNearEyeAlert []string
	StudyNearEyeCol   *term.RGB
)

// NearEyeShipped is the shipped alert ring, so a study can show the control
// beside the candidates without transcribing it and drifting.
func NearEyeShipped() []string { return append([]string{}, nearEyeAlert...) }

// NearEyeBead is the shipped working eye, which is what "still reads as a
// different shape" has to be measured against.
func NearEyeBead() []string { return append([]string{}, nearEyeOpen...) }
