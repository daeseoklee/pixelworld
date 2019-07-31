package world

import (
	"errors"

	"github.com/daeseoklee/pixelworld/minion"
	"github.com/daeseoklee/pixelworld/pos"
	"github.com/daeseoklee/pixelworld/util"
)

//Register : locate and register a object
func (w *World) Register(obj Obj, loc pos.Abs, head pos.Direc, changeVisible bool) {
	w.Objects[obj] = true
	w.Loc[obj] = loc
	w.Head[obj] = head
	switch v := obj.(type) {
	case *minion.Minion:
		v.SetID(w.ID)
		w.ID++
		w.Minions[v] = true
		if changeVisible {
			if w.Available(obj, loc, head) {
				w.SetVisible(obj, true)
			} else {
				panic(errors.New("trying to visibly register when it is not available"))
			}
		}
	case *outer:
		w.SetVisible(obj, true)
	case *wall:
		w.SetVisible(obj, true)
	default:
		{
		}
	}
}

//FindAvailable : try finding available (location,head)
func (w *World) FindAvailable(obj Obj) (pos.Abs, pos.Direc, bool) {
	var loc pos.Abs
	var head pos.Direc
	for i := 0; i < LocateTry; i++ {
		loc = pos.Abs{X: util.RandInt(w.Xlen), Y: util.RandInt(w.Ylen)}
		head = pos.DirecFromNum(util.RandInt(4))
		if w.Available(obj, loc, head) {
			return loc, head, true
		}
	}
	return loc, head, false
}

//LocateAndRegister : try FindAvailable and register if succeeded
func (w *World) LocateAndRegister(obj Obj) bool {
	loc, head, ok := w.FindAvailable(obj)
	if ok {
		w.Register(obj, loc, head, true)
		//fmt.Println("approved : ", loc)
		return true
	}
	return false
}

//SetVisible : make invisible/visible in the world
func (w *World) SetVisible(obj Obj, b bool) {
	switch {
	case (obj.Visible() && b) || (!obj.Visible() && !b):
		panic(errors.New("visibility errorrrrrrrr"))
	case obj.Visible() && !b:
		obj.SetVisible(false)
		for _, loc := range w.Occupying(obj) {
			delete(w.OccupiedBy, loc)
		}
	case (!obj.Visible()) && b:
		obj.SetVisible(true)
		if !w.Available(obj, w.Loc[obj], w.Head[obj]) {
			panic(errors.New("trying to be visible when it is not available"))
		}
		for _, loc := range w.Occupying(obj) {
			w.OccupiedBy[loc] = obj
		}
	}

}
