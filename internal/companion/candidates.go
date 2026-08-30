package companion

import "github.com/donlucasx/asciiscapes/internal/term"

var (
	warm  = term.RGB{R: 232, G: 224, B: 206} // moonlit fur
	rust  = term.RGB{R: 226, G: 150, B: 96}  // fox
	shell = term.RGB{R: 232, G: 132, B: 110} // crab
	dusty = term.RGB{R: 206, G: 200, B: 214} // moth
	green = term.RGB{R: 150, G: 208, B: 150} // frog
	glow  = term.RGB{R: 216, G: 236, B: 255} // wisp
	eye   = term.RGB{R: 40, G: 44, B: 56}
	spark = term.RGB{R: 255, G: 250, B: 226}
)

// Candidates is the breadth sheet. Each creature is drawn in whichever register
// flatters it — forcing all of them into one style would bias the choice of
// animal, which is the thing this sheet is meant to decide.
var Candidates = []Sprite{
	{
		Name: "cat", Register: Line, Note: "tail flick, slow blink",
		Rows:   []string{` /\_/\`, `( o.o )`, ` > ^ <`},
		Body:   warm,
		Accent: eye, AccentOf: "o^",
	},
	{
		Name: "bird", Register: Line, Note: "hop, head tilt, takes off",
		Rows:   []string{`  __`, ` (o )>`, `  ||`},
		Body:   warm,
		Accent: eye, AccentOf: "o",
	},
	{
		Name: "fox", Register: Line, Note: "curls up, ears prick, tail sweeps",
		Rows:   []string{` /\_/\   _`, `( o.o )_/ )`, ` > ^ <___/`},
		Body:   rust,
		Accent: eye, AccentOf: "o",
	},
	{
		Name: "wisp", Register: Braille, Note: "drifts, pulses; no body at all",
		Rows:   []string{`  ⡀ `, ` ⢰⣿⡆`, `  ⠉ `},
		Body:   glow,
		Accent: spark, AccentOf: "⣿",
	},
	{
		Name: "crab", Register: Line, Note: "sideways scuttle, claw raise = !",
		Rows:   []string{` \(oo)/`, ` /~~~~\`, `  ^  ^`},
		Body:   shell,
		Accent: eye, AccentOf: "o",
	},
	{
		Name: "moth", Register: Line, Note: "flies to light — notification built in",
		Rows:   []string{` \\|//`, `  (o)`, ` //|\\`},
		Body:   dusty,
		Accent: eye, AccentOf: "o",
	},
	{
		Name: "otter", Register: Line, Note: "floats on its back; hard to animate",
		Rows:   []string{` (o.o)`, `/|   |\`, ` ^^^^^`},
		Body:   warm,
		Accent: eye, AccentOf: "o",
	},
	{
		Name: "frog", Register: Line, Note: "stillness, then one hop",
		Rows:   []string{` @..@`, `(----)`, ` (\/)`},
		Body:   green,
		Accent: eye, AccentOf: "@",
	},
	{
		Name: "owl", Register: Line, Note: "head swivel, blink; nocturnal fits the shore",
		Rows:   []string{` ,___,`, ` (o,o)`, ` /)_)`},
		Body:   warm,
		Accent: eye, AccentOf: "o",
	},
	{
		Name: "cat", Register: Block, Note: "silhouette variant, for comparison",
		Rows:  []string{` ▟▌▐▙`, `▐███▌`, ` ▘ ▝`},
		Body:  term.RGB{R: 24, G: 22, B: 30},
		Alpha: 1,
	},
	{
		Name: "fox", Register: Block, Note: "silhouette variant — the tail is the signature",
		Rows:  []string{` ▟▙▟▙   ▄`, `▐████▙▄▟█▘`, ` ▘▘  ▝▀▀`},
		Body:  term.RGB{R: 26, G: 20, B: 24},
		Alpha: 1,
	},
	{
		Name: "bird", Register: Braille, Note: "braille variant, finest detail",
		Rows: []string{` ⢀⣀`, `⢰⡿⠋⠓`, ` ⠈⠙`},
		Body: warm,
	},
}
