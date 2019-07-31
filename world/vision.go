package world

import (
	"math"

	"github.com/daeseoklee/pixelworld/colour"
	"github.com/daeseoklee/pixelworld/minion"
	"github.com/daeseoklee/pixelworld/pos"
)

//Colour : colour of a cell, determined by the mineral
func (min Mineral) Colour() colour.Colour {
	p, a := float64(min.p), float64(min.a)
	return colour.Colour{M: math.Tanh((p + a) / C), R: math.Tanh(a / C), G: math.Tanh(p / C), B: 0}
}

//Render : rendering of the world in colour.Colour
func (w *World) Render() {
	var loc pos.Abs
	var obj Obj
	var ok bool
	for i := 0; i < w.Xlen; i++ {
		for j := 0; j < w.Ylen; j++ {
			loc = pos.Abs{X: i, Y: j}
			obj, ok = w.OccupiedBy[loc]
			if !ok {
				w.SnapShot[i][j] = w.Min[loc].Colour()
			} else {
				//fmt.Println("this object's head:", obj)
				//fmt.Println("location: ", w.Loc[obj])
				//fmt.Println("head:", w.Head[obj])
				//mi := obj.(*minion.Minion)
				//fmt.Println("in w.Minions?:", w.Minions[mi])
				//fmt.Println("in w.Objects?:", w.Objects[obj])
				w.SnapShot[i][j] = obj.Render(pos.ToRel(w.Loc[obj], w.Head[obj], loc))
			}
		}
	}
}

//WriteVision : write visual input in the given slice
func (w *World) WriteVision(mi *minion.Minion, here [][]colour.Colour) {
	var loc pos.Abs
	var x, y int
	for i := -3*mi.Alen() - 1; i <= 3*mi.Alen()+1; i++ {
		for j := -mi.Blen(); j <= 5*mi.Blen()+2; j++ {
			loc = w.ToAbs(mi, pos.Rel{Z: i, W: j})
			x, y = loc.X, loc.Y
			if w.InRange(loc) {
				here[3*mi.Alen()+1+i][mi.Blen()+j] = w.SnapShot[x][y]
			} else {
				here[3*mi.Alen()+1+i][mi.Blen()+j] = OuterColour
			}
		}
	}
}
