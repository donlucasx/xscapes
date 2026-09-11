package main

import (
	"fmt"
	"sort"
	"strings"
)

// double2x scales a source bitmap by 2 in both axes the only way that makes
// "exact 2x" mean anything: every '#' becomes a 2x2 block of '#'.
func double2x(rows []string) []string {
	var out []string
	for _, r := range rows {
		var sb strings.Builder
		for _, c := range r {
			sb.WriteRune(c)
			sb.WriteRune(c)
		}
		d := sb.String()
		out = append(out, d, d)
	}
	return out
}

func diffRows(name string, want, got []string) {
	if len(want) != len(got) {
		fmt.Printf("%-12s LENGTH MISMATCH: exact-2x gives %d rows, theirs has %d\n", name, len(want), len(got))
		return
	}
	bad := 0
	for i := range want {
		if want[i] != got[i] {
			if bad < 6 {
				fmt.Printf("%-12s row %2d differs\n   want %s\n   got  %s\n", name, i, want[i], got[i])
			}
			bad++
		}
	}
	if bad == 0 {
		fmt.Printf("%-12s IDENTICAL to an exact 2x of the shipped rows (%d rows)\n", name, len(want))
	} else {
		fmt.Printf("%-12s %d of %d rows differ from an exact 2x\n", name, bad, len(want))
	}
}

func vocab(name string, q []string) {
	m := map[rune]int{}
	for _, r := range q {
		for _, c := range r {
			m[c]++
		}
	}
	var ks []rune
	for k := range m {
		if k != ' ' {
			ks = append(ks, k)
		}
	}
	sort.Slice(ks, func(i, j int) bool { return m[ks[i]] > m[ks[j]] })
	fmt.Printf("%-10s %d distinct ink glyphs:", name, len(ks))
	for _, k := range ks {
		fmt.Printf(" %c(%d)", k, m[k])
	}
	fmt.Println()
}

func partTwo(theirQ, shippedQ []string) {
	fmt.Println("\n=== 8. IS IT ACTUALLY AN EXACT 2x? ===")
	diffRows("upper", double2x(shippedAsk), theirUpper)
	diffRows("lower", double2x(shippedLower), theirLower)

	fmt.Println("\n=== 9. GLYPH VOCABULARY (what the doubling costs) ===")
	vocab("shipped", shippedQ)
	vocab("theirs", theirQ)
	fmt.Println("An exact 2x can only ever emit ' ', '▀', '▄', '█': the doubled source makes both")
	fmt.Println("subpixels of a half identical, so no partial quadrant (▖▘▝▗▌▐▞▚▛▜▟▙) can occur.")
	fmt.Println("The shipped crab's roundness lives in exactly those glyphs.")
}
