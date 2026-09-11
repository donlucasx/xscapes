package main

import "fmt"

// paperRejects are the ideas that do not need rendering, because the rule
// settles them before a pixel is drawn. Each one is written down so nobody
// spends an evening on it later.
func paperRejects() {
	fmt.Println("\n\n==== 8. REJECTED ON THE RULE, NO RENDER NEEDED ====")
	for _, r := range [][2]string{
		{"COLOUR BY WHAT FINISHED (a blue star for a read, a warm one for an edit)",
			"Identity is the SAND's channel -- the locked table: \"the activity tail names\n     the tool and the file. The sea says how much; the sand says what.\" Putting\n     identity in the sky binds a second variable AND duplicates a channel."},
		{"THE STARS FORM A NAMED ASTERISM AT CERTAIN COUNTS",
			"Requires placement to depend on the TOTAL, which breaks the one guarantee\n     the table locks: \"Position is fixed by index and seed so a star lights where\n     it always was.\" A sky that rearranges itself is noise, and a glance is the\n     whole budget."},
		{"A FAINT PLACEHOLDER WHERE A STAR IS STILL TO COME",
			"Already shipped, already killed. His ruling 2026-09-05: \"discard the ring\n     altogether, it's not clear what it means.\" Do not bring it back in a new\n     glyph."},
		{"THE STARS RISE FROM THE HORIZON AS THEY AGE",
			"Two failures at once: altitude is the MOON's, carrying context remaining,\n     and a star that moves after it lights encodes in rate and loses its fixed\n     position."},
		{"THE CONSTELLATION DRIFTS WITH THE HOUR, LIKE A REAL SKY",
			"Beautiful, and it is the SKY's channel: \"the water is the work, the sky is\n     the world.\" Time of day already owns sky colour. Adding star drift makes the\n     agent's accumulator move for a reason that has nothing to do with the agent."},
	} {
		fmt.Printf("\n  ✗ %s\n     %s\n", r[0], r[1])
	}
}
