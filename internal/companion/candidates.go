package companion

import "github.com/donlucasx/asciiscapes/internal/term"

var (
	warm  = term.RGB{R: 232, G: 224, B: 206}
	rust  = term.RGB{R: 226, G: 150, B: 96}
	green = term.RGB{R: 150, G: 208, B: 150}
	slate = term.RGB{R: 186, G: 194, B: 208}
	eye   = term.RGB{R: 40, G: 44, B: 56}
)

// Candidates are Lucas's references translated onto our canvas. Every CJK and
// Ambiguous-width glyph in the originals has been swapped for its narrow
// equivalent -- kaomoji descend from ASCII emoticons, so this mostly means
// putting them back: (.w.) for (・ω・), ('^') for ('人'), ~ for the arcs,
// o for the circles. TestCandidatesAreNarrowSafe enforces it.
//
// Tall and short cuts of the same creature sit next to each other so the cost
// of compression is visible rather than argued about.
var Candidates = []Sprite{
	{
		Name: "alpaca", Register: Line, Source: "ref 1", Note: "7 rows -- full fluff, stacked-paren neck",
		Rows: []string{
			` /\  __  /\`,
			`((o)('^')(o))`,
			`   (  ~  )`,
			`   (     )`,
			`   (     )   /\`,
			`  (       )~(  )`,
			`   \_/ \_/  \_/`,
		},
		Body: warm, Accent: eye, AccentOf: "o",
	},
	{
		Name: "alpaca", Register: Line, Source: "ref 1", Note: "4 rows -- neck sacrificed, face survives",
		Rows: []string{
			` /\ __ /\`,
			`((o)('^')(o))`,
			`   (     )`,
			`   \_/ \_/`,
		},
		Body: warm, Accent: eye, AccentOf: "o",
	},
	{
		Name: "cat, detailed", Register: Line, Source: "ref 2", Note: "8 rows -- original was already pure ASCII",
		Rows: []string{
			`(_\`,
			` ) )`,
			`( (   .-""-.   A.-.A`,
			` \ \/      \/  , , \`,
			`  \  \     =;  t  /=`,
			`   \ |""".    ',--'`,
			`    / //   | ||`,
			`   /_,))   |_,))`,
		},
		Body: warm, Accent: eye, AccentOf: "=;",
	},
	{
		Name: "cat, meow", Register: Line, Source: "ref 4", Note: "7 rows + bubble -- the dialogue idea",
		Rows: []string{
			`  |\__/|`,
			` (_ ^-^)`,
			`   )   (`,
			`_  )   (`,
			`(( /    \`,
			` (  ) || ||`,
			` '--'  '--'`,
		},
		Body: warm, Accent: eye, AccentOf: "^",
		Say: "meow!",
	},
	{
		Name: "cat, meow", Register: Line, Source: "ref 4", Note: "4 rows + bubble -- compressed",
		Rows: []string{
			` |\__/|`,
			`(_ ^-^)`,
			`  )   (`,
			` '--' '--'`,
		},
		Body: warm, Accent: eye, AccentOf: "^",
		Say: "tests passed",
	},
	{
		Name: "goat", Register: Line, Source: "ref 5", Note: "7 rows -- the shi/kana legs became (_/",
		Rows: []string{
			`      ,,`,
			`     //`,
			` (\,'"~-,`,
			`~=/  . -)`,
			`  <'"~>  \`,
			`  ) )_, ;\`,
			` (_/(_/ |_)`,
		},
		Body: slate, Accent: eye, AccentOf: ".",
		Say: "meh meh",
	},
	{
		Name: "goat", Register: Line, Source: "ref 5", Note: "4 rows -- horns and beard only",
		Rows: []string{
			`   ,,`,
			`(\,'"~-,`,
			`~=/ . -)`,
			` (_/(_/`,
		},
		Body: slate, Accent: eye, AccentOf: ".",
	},
	{
		Name: "koala", Register: Line, Source: "ref 3", Note: "4 rows -- clinging to a trunk; Omega nose became w",
		Rows: []string{
			` (\  /)`,
			` ( .w. )`,
			` (  )) |:|`,
			`  \__/ |:|`,
		},
		Body: slate, Accent: eye, AccentOf: ".",
	},
	{
		Name: "bunny", Register: Line, Source: "ref 6 grid", Note: "3 rows -- ((.w.)) from ((kana))",
		Rows: []string{
			` /\_/\`,
			`((.w.))`,
			` (   )`,
		},
		Body: warm, Accent: eye, AccentOf: ".",
	},
	{
		Name: "cat, wand", Register: Line, Source: "ref 6 grid", Note: "3 rows -- the arm-and-star kaomoji",
		Rows: []string{
			` /\_/\`,
			`(.w.)~~--*`,
			` (  )`,
		},
		Body: warm, Accent: eye, AccentOf: ".",
	},
	{
		Name: "cat, classic", Register: Line, Source: "ref 6 grid", Note: "4 rows -- the /l, cat, kana legs redrawn",
		Rows: []string{
			` /l,`,
			`( ., 7`,
			` l  ~\`,
			` U_,)J`,
		},
		Body: warm, Accent: eye, AccentOf: ".",
	},
	{
		Name: "bear", Register: Line, Source: "ref 6 grid", Note: "3 rows -- smallest thing that still has a face",
		Rows: []string{
			` (\_/)`,
			`('.w.')`,
			` (   )`,
		},
		Body: rust, Accent: eye, AccentOf: ".",
	},
	{
		Name: "frog", Register: Line, Source: "ours, kept", Note: "3 rows -- from the first sheet, for comparison",
		Rows: []string{
			` @..@`,
			`(----)`,
			` (\/)`,
		},
		Body: green, Accent: eye, AccentOf: "@",
	},
}
