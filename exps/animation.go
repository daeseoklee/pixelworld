package exps

import (
	"github.com/daeseoklee/pixelworld/colour"
	"github.com/daeseoklee/pixelworld/world"
	"github.com/hajimehoshi/ebiten"
)

var (
	w *world.World
	m int
	n int
	k float64
)

func setVars() {
	m, n = w.Xlen, w.Ylen
	switch m {
	case 70:
		k = 11.0
	case 150:
		k = 5.0
	case 400:
		k = 2.0
	case 800:
		k = 1.0
	default:
		k = 1.0
	}
}

func animate() {
	if err := ebiten.Run(update, m, n, k, "pixelworld"); err != nil {
		panic(err)
	}
}

func update(screen *ebiten.Image) error {
	if ebiten.IsDrawingSkipped() {
		return nil
	}
	ebitim, err := ebiten.NewImageFromImage(colour.ToImage(w.SnapShot, 1), 0)
	if err != nil {
		panic(err)
	}
	screen.DrawImage(ebitim, nil)
	return nil
}
