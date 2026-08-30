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
	case r == 0x2590: // RIGHT HALF BLOCK -- Narrow, unlike the other three halves
		return true
	case r >= 0x2596 && r <= 0x259F: // quadrants -- all Narrow, unlike the halves
		return true
	}
	return false
}

// isBlockSafe covers the four block characters quadrant rendering cannot do
// without: top half, bottom half, left half and full.
//
// These are AMBIGUOUS width, not Narrow: one cell in a Western terminal, two
// where the terminal is configured for East Asian ambiguous-width. That is a
// deliberate, documented exception rather than an oversight -- there is no
// narrow substitute for a full block. Any sprite relying on these is reported
// by TestReportsAmbiguousReliance so the exposure stays visible.
func isBlockSafe(r rune) bool {
	switch r {
	case '\u2580', '\u2584', '\u258C', '\u2588': // top, bottom, left, full
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
