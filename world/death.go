package world

import (
	"errors"

	"github.com/daeseoklee/pixelworld/minion"
)

//Kill : Minion death
func (w *World) Kill(mi *minion.Minion) {
	if !w.Objects[mi] {
		return
	}
	//fmt.Println("killed!!")
	if !w.Minions[mi] {
		panic(errors.New("trying to kill non-existing minion"))
	}
	//mineral distribution

	l := w.Body(mi)
	num := 0 //number of inrange body positions. careful consideration needed for newborn ill-fated near-edge babies
	for _, loc := range l {
		if w.InRange(loc) {
			num++
		}
	}
	minPerCell := mi.Mass() / num
	remain := mi.Mass() % num
	sofar := 0
	for _, loc := range l {
		if w.InRange(loc) {
			if sofar < remain {
				w.Min[loc].a += minPerCell + 1
				sofar++
			} else {
				w.Min[loc].a += minPerCell
			}
		}
	}
	//temporary invisibles can also be killed
	if mi.Visible() {
		for _, loc := range w.Occupying(mi) {
			delete(w.OccupiedBy, loc)
		}
	}
	//removing
	delete(w.Objects, mi)
	delete(w.Minions, mi)
	delete(w.Loc, mi)
	delete(w.Head, mi)
}

//IncreaseAge : Minion ageing with death when required
func (w *World) IncreaseAge(mi *minion.Minion) {
	mi.IncreaseAge()
	if mi.FullAge() {
		w.Kill(mi)
	}
}
