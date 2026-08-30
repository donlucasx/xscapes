package companion

// Terminals render CJK and many symbol characters two cells wide, which shears
// a monospace grid apart. Rather than carry a Unicode East Asian Width table,
// sprites are restricted to an ALLOW-LIST of runes verified narrow.
//
// The fail direction is the point: an unrecognised rune is rejected, not
// admitted. A deny-list would silently pass every character nobody thought to
// ban, which is exactly how kaomoji art gets into a codebase and breaks it.
//
// Verified against unicodedata.east_asian_width: every rune below is N or Na.
// Notably NOT here, despite being tempting: · ° ≈ • ▒ ▓ █ ─ ⌒ ● ∧ ☆ ω (all
// Ambiguous — one cell in a Western terminal, two in a CJK locale) and 人 つ
// ・ ミ ノ (Wide — two cells everywhere).
func isNarrow(r rune) bool {
	switch {
	case r >= 0x20 && r <= 0x7E: // printable ASCII
		return true
	case r == 0x203F: // ‿ UNDERTIE
		return true
	case r == 0x2218: // ∘ RING OPERATOR
		return true
	case r == 0x2591: // ░ LIGHT SHADE (medium/dark/full are Ambiguous)
		return true
	case r >= 0x2800 && r <= 0x28FF: // braille
		return true
	}
	return false
}

// UnsafeRunes reports every rune in the sprite that is not verified narrow.
func (s *Sprite) UnsafeRunes() []rune {
	var bad []rune
	seen := map[rune]bool{}
	for _, row := range s.Rows {
		for _, r := range row {
			if !isNarrow(r) && !seen[r] {
				seen[r] = true
				bad = append(bad, r)
			}
		}
	}
	return bad
}
