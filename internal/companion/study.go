package companion

import "github.com/donlucasx/xscapes/internal/canvas"

// PlotRim is plotRim for the studies under notes/: the one-cell clearing
// around a sprite that keeps two bodies from reading as one shape. Exported
// so a study animal goes through exactly the pipeline the cat does.
func PlotRim(l *canvas.Layer, rows []string, x, y int) { plotRim(l, rows, x, y) }
