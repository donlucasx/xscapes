package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/notes/s28-closer/b-push/pose"
)

// The shipped crab lives in pose.Ship*, hand-copied from internal/companion and
// proven against the REAL companion by shipCheck().
var (
	shipAsk   = pose.ShipAsk
	shipWork  = pose.ShipWork
	shipLower = pose.ShipLower

	midAsk      = pose.MidAsk
	midWork     = pose.MidWork
	midLower    = pose.MidLower
	midEye      = pose.MidEye
	midEyeCells = pose.MidEyeCells
	midEyeRow   = pose.MidEyeRow

	bigAsk      = pose.BigAsk
	bigWork     = pose.BigWork
	bigLower    = pose.BigLower
	bigEyeSolid = pose.BigEyeSolid
	bigEyeRing  = pose.BigEyeRing
	bigEyeCells = pose.BigEyeCells
	bigEyeRow   = pose.BigEyeRow
)

func todayRef() {
	fmt.Println("=== TODAY, the shipped ask pose (12 cells x 7) ===")
	fmt.Println("eyes are single-cell glyphs 'O' plotted on top at cells {4,7}, row 2")
	show("ship ask", append(append([]string{}, shipAsk...), shipLower...))
}
