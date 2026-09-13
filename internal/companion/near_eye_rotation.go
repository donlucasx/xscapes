package companion

// THE ROTATING NEAR EYE. His ruling of 2026-09-12, from a screenshot of his own
// live ask: "Id like to keep the round off, the high pupil, the hood and the
// hooded ring on rotation. It could be random each time, or tied to something
// if we can differentiate types of moments."
//
// RANDOM, and the reason is the encoding rule. A random eye carries NO
// information, so it cannot put a second meaning on a channel that already says
// "the agent needs you" -- the same argument that permits the crab's claw
// twitch. Tying it to a kind of moment would have bound a second variable to
// one channel, which the rule forbids.
//
// ⚠ The eye is picked from the ASK ORDINAL, never from the clock, so it is
// constant for the whole of one ask. An eye chosen per frame would change the
// animal's face twenty times a second while he is reading the question.
//
// A fifth candidate, "Sea Glass", was drafted and dropped. It was the only one
// of the five that changed the COLOUR rather than the shape, and the quantiser
// says colour has almost nowhere to go: EyeAlert's chroma is 26, under the
// neutralChroma cliff of 30, so every "softer mint" from 220,244,220 down to
// 210,238,212 collapses onto the SAME cube entry; crossing 30 does move the
// entry, but toward the working bead, cutting the ask-vs-working colour
// distance from 89 to as little as 40. Shape was the only axis with room.

var nearEyeRotation = [][]string{
	// Round Off.
	{
		".##.",
		".##.",
		"#..#",
		"#..#",
		"#..#",
		"#..#",
		".##.",
		".##.",
	},
	// The high pupil.
	{
		"####",
		"####",
		"#..#",
		"#..#",
		"####",
		"####",
		"####",
		"####",
	},
	// The hood.
	{
		"####",
		"####",
		"#..#",
		"#..#",
		"#..#",
		"#..#",
		"....",
		"....",
	},
	// the hooded ring.
	{
		"....",
		"....",
		"####",
		"####",
		"#..#",
		"#..#",
		".##.",
		".##.",
	},
}

// nearEyeNames is the rotation in order, for tests and for a study page that
// has to label which face it is showing.
var nearEyeNames = []string{
	"Round Off",
	"The high pupil",
	"The hood",
	"the hooded ring",
}

// eyePick is which of the four this ask gets.
//
// A hash rather than askN % 4, because a modulo cycles in a fixed order and
// four asks in a row would walk the same carousel every session. The scene seed
// is not available here, so the counter alone is mixed -- splitmix64's
// finaliser, which is enough to decorrelate consecutive integers.
func (c *Cat) eyePick() int {
	h := uint64(c.askN) * 0x9E3779B97F4A7C15
	h ^= h >> 31
	h *= 0x94D049BB133111EB
	h ^= h >> 29
	return int(h % uint64(len(nearEyeRotation)))
}

// NearEyeRotationArt is one rotating eye's art, exported so a study page can
// show which face an ask would get without transcribing the bitmaps.
func NearEyeRotationArt(i int) []string {
	if i < 0 || i >= len(nearEyeRotation) {
		return nearEyeAlert
	}
	return nearEyeRotation[i]
}

// NearEyeRotationNames is the rotation's labels, in order.
func NearEyeRotationNames() []string { return append([]string{}, nearEyeNames...) }
