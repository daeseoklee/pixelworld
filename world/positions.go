package world

import (
	"errors"

	"github.com/daeseoklee/pixelworld/minion"
	"github.com/daeseoklee/pixelworld/pos"
)

//Poss : slice of all positions
func (w *World) Poss() []pos.Abs {
	l := make([]pos.Abs, w.Xlen*w.Ylen)
	for i := 0; i < w.Xlen; i++ {
		for j := 0; j < w.Ylen; j++ {
			l[i*w.Ylen+j] = pos.Abs{X: i, Y: j}
		}
	}
	return l
}

//InRange : check whether the absolute position is in the world range
func (w *World) InRange(p pos.Abs) bool {
	return (p.X >= 0) && (p.X < w.Xlen) && (p.Y >= 0) && (p.Y < w.Ylen)
}

//ToAbs : given object and a relative position, return the absolute position
func (w *World) ToAbs(obj Obj, rp pos.Rel) pos.Abs {
	loc, ok := w.Loc[obj]
	head := w.Head[obj]
	if !ok {
		panic(errors.New("trying to find the location of an unexisting object"))
	}
	return pos.ToAbs(loc, head, rp)
}

//ToAbss : given object and a slice of relative positions, return a slice of absolute positions
func (w *World) ToAbss(obj Obj, points []pos.Rel) []pos.Abs {
	loc, ok := w.Loc[obj]
	head := w.Head[obj]
	if !ok {
		panic(errors.New("trying to find the location of an unexisting object"))
	}
	l := make([]pos.Abs, len(points))
	for i, a := range points {
		l[i] = pos.ToAbs(loc, head, a)
	}
	return l
}

//Occupying : list of pos.Abs that the object is occupying
func (w *World) Occupying(obj Obj) []pos.Abs {
	return w.ToAbss(obj, obj.Shape())
}

//Body : list of pos.Abs that the minion body is on
func (w *World) Body(mi *minion.Minion) []pos.Abs {
	return w.ToAbss(mi, mi.BodyShape())
}

//Available : whether the allocation is available
func (w *World) Available(obj Obj, loc pos.Abs, head pos.Direc) bool {
	for _, a := range obj.Shape() {
		if w.OccupiedBy[pos.ToAbs(loc, head, a)] != nil {
			return false
		}
	}
	return true
}
