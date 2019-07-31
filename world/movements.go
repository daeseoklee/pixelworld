package world

import (
	"errors"

	"github.com/daeseoklee/pixelworld/minion"
	"github.com/daeseoklee/pixelworld/pos"
	"github.com/daeseoklee/pixelworld/util"
)

func (w *World) spConsum(obj Obj, moveTy int) (int, bool) {
	var mi *minion.Minion
	switch obj.(type) {
	case *minion.Minion:
		mi = obj.(*minion.Minion)
	default:
		return 0, true
	}
	amount, ok := mi.SpConsum(moveTy)
	return amount, ok
}

func (w *World) spReduce(obj Obj, moveTy int) {
	var mi *minion.Minion
	switch obj.(type) {
	case *minion.Minion:
		mi = obj.(*minion.Minion)
	default:
		return
	}
	mi.SpReduce(moveTy)
}

func (w *World) spIncrease(obj Obj) {
	var mi *minion.Minion
	switch obj.(type) {
	case *minion.Minion:
		mi = obj.(*minion.Minion)
	default:
		return
	}
	mi.SpIncrease()
}

//adjacent : set of adjacent free-moving objects, whether there is any adjacent non free-moving object
func (w *World) adjacent(obj0 Obj, direc pos.Direc) (map[Obj]bool, bool) {
	adj := make(map[Obj]bool)
	crash := false
	for _, loc := range w.Occupying(obj0) {
		switch obj := w.OccupiedBy[pos.Next(loc, direc)]; obj {
		case obj0:
			{
			}
		case nil:
			{
			}
		default:
			if obj.Free() {
				adj[obj] = true
			} else {
				crash = true
			}
		}
	}
	return adj, crash
}

//crash : whether obj0 is blocked to the direc. cyclic search avoided using 'exclude'
func (w *World) crash(obj0 Obj, direc pos.Direc, exclude map[Obj]bool) bool {
	//fmt.Println("in crash")
	exclude[obj0] = true
	adj, crash := w.adjacent(obj0, direc)
	if crash {
		return true
	}
	for obj := range adj {
		if exclude[obj] {
			continue
		}
		if w.crash(obj, direc, exclude) {
			delete(exclude, obj0)
			return true
		}
	}
	delete(exclude, obj0)
	return false
}

//bundle :bundle of obj0 to direc, as in the doc. Added to hereAdd
func (w *World) bundle(obj0 Obj, direc pos.Direc, crash0 bool, hereAdd map[Obj]bool) {
	//fmt.Println("in bundle")
	hereAdd[obj0] = true
	exclude := make(map[Obj]bool)
	if !crash0 {
		adj, _ := w.adjacent(obj0, direc)
		for obj := range adj {
			if !hereAdd[obj] {
				w.bundle(obj, direc, false, hereAdd)
			}
		}
	} else {
		adj, _ := w.adjacent(obj0, direc)
		for obj := range adj {
			if (!hereAdd[obj]) && w.crash(obj, direc, exclude) {
				w.bundle(obj, direc, true, hereAdd)
			}
		}
	}
}

func (w *World) mass(bundle map[Obj]bool) int {
	sum := 0
	for obj := range bundle {
		sum += obj.Mass()
	}
	return sum
}

//updateTranslation : apply the move and digest
func (w *World) updateTranslation(objs map[Obj]bool, direc pos.Direc) {
	toBeEmpty := make(map[pos.Abs]bool)
	for obj := range objs {
		for _, loc := range w.Occupying(obj) {
			if !objs[w.OccupiedBy[pos.Previous(loc, direc)]] {
				toBeEmpty[loc] = true
			}
		}
	}
	for loc := range toBeEmpty {
		delete(w.OccupiedBy, loc)
	}
	for obj := range objs {
		for _, loc := range w.Occupying(obj) {
			w.OccupiedBy[pos.Next(loc, direc)] = obj
		}
		w.Loc[obj] = pos.Next(w.Loc[obj], direc)
	}
	for obj := range objs {
		w.Digest(obj)
	}
}

//updateRotation : apply the rotation and digest
func (w *World) updateRotation(obj Obj, angle int) {
	toBeEmpty := make(map[pos.Rel]bool)
	for _, rp := range obj.Shape() {
		if w.OccupiedBy[pos.ToAbs(w.Loc[obj], pos.Rotate(w.Head[obj], -angle), rp)] != obj {
			toBeEmpty[rp] = true
		}
	}
	for rp := range toBeEmpty {
		delete(w.OccupiedBy, w.ToAbs(obj, rp))
	}
	for _, rp := range obj.Shape() {
		w.OccupiedBy[pos.ToAbs(w.Loc[obj], pos.Rotate(w.Head[obj], angle), rp)] = obj
	}
	w.Head[obj] = pos.Rotate(w.Head[obj], angle)
	w.Digest(obj)
}

//Move : move
func (w *World) Move(obj Obj, moveTy int, moveDist int, voluntary bool, force, damage int) string {
	if !w.Objects[obj] {
		return "already dead"
	}
	switch {
	case moveTy == 0:
		w.Digest(obj)
		if voluntary {
			w.spIncrease(obj)
		}
		return "pause"
	case moveTy <= 4:
		if moveDist == 0 {
			if voluntary {
				w.spIncrease(obj)
			}
			return "success"
		}
		_, spok := w.spConsum(obj, moveTy)
		if voluntary && !spok {
			w.spIncrease(obj)
			return "lack stamina"
		}
		direc := pos.ToDirec(w.Head[obj], pos.RelDirecFromNum(moveTy))
		crash := w.crash(obj, direc, make(map[Obj]bool))
		bundle := make(map[Obj]bool)
		w.bundle(obj, direc, crash, bundle)
		if !crash {
			if voluntary && force < w.mass(bundle)-obj.Mass() {
				w.spIncrease(obj)
				return "heavy"
			}
			if (!voluntary) && force < w.mass(bundle) {
				return "heavy"
			}
			w.updateTranslation(bundle, direc)
			if voluntary {
				w.spReduce(obj, moveTy)
			}
			return w.Move(obj, moveTy, moveDist-1, voluntary, force, damage)
		}
		for victim := range bundle {
			if voluntary {
				if victim != obj {
					w.Attack(victim, damage)
				}
			} else {
				w.Attack(victim, damage)
			}
		}
		if voluntary {
			w.spReduce(obj, moveTy)
		}
		return "crash"
	case moveTy <= 7:
		var angle int
		switch moveTy {
		case 5:
			angle = 1
		case 6:
			angle = 3
		case 7:
			angle = 2
		}
		_, spok := w.spConsum(obj, moveTy)
		if voluntary && !spok {
			w.Digest(obj)
			w.spIncrease(obj)
			return "lack stamina"
		}
		//already occupied
		for _, rp := range obj.Shape() {
			switch w.OccupiedBy[pos.ToAbs(w.Loc[obj], pos.Rotate(w.Head[obj], angle), rp)] {
			case nil:
				{
				}
			case obj:
				{
				}
			case w.Outer:
				//v := obj.(*minion.Minion)
				//fmt.Println("prevented from going outside", w.Head[obj], angle, w.Loc[obj], v)
				w.Digest(obj)
				if voluntary {
					w.spIncrease(obj)
				}
				return "occupied"
			default:
				w.Digest(obj)
				if voluntary {
					w.spIncrease(obj)
				}
				return "occupied"
			}
		}
		//turning
		w.updateRotation(obj, angle)
		w.Digest(obj)
		if voluntary {
			w.spReduce(obj, moveTy)
			w.spIncrease(obj)
		}
		return "turned"
	default:
		panic(errors.New("invalid moveTy"))
	}

}

//BirthMove : powerful forward movement at the birth
func (w *World) BirthMove(mi *minion.Minion) {
	force := util.LinVal(BirthForce, mi.Mass())
	damage := force
	perMove, _ := w.spConsum(mi, 1)
	dist := util.RandInt(1 + (mi.SpGet() / perMove))
	w.Move(mi, 1, dist, true, force, damage)
}

//Jaw : jaw movement and consequences, return false if cancelled
func (w *World) Jaw(mi *minion.Minion, n int) bool {
	switch {
	case n == 0:
		return true
	case n == 2:
		if !w.Jaw(mi, 1) {
			return false
		}
		return w.Jaw(mi, 1)
	case n == 1 && !mi.JawOpen():
		_, spok := w.spConsum(mi, -1)
		if !spok {
			return false
		}
		w.spReduce(mi, -1)
		for i := 1; i <= mi.Xlen()-2; i++ {
			delete(w.OccupiedBy, w.ToAbs(mi, pos.Rel{Z: -mi.Alen() + i, W: mi.Blen()}))
		}
		mi.SetJawState(0, 0)
		return true
	case n == 1 && mi.JawOpen():
		_, spok := w.spConsum(mi, -1)
		if !spok {
			return false
		}
		w.spReduce(mi, -1)
		//intercourse mount
		sexy := false
		var partner *minion.Minion
		var obj Obj
		var loc pos.Abs
		for _, rp := range mi.JawPoss() {
			loc = w.ToAbs(mi, rp)
			obj = w.OccupiedBy[loc]
			switch v := obj.(type) {
			case *minion.Minion:
				switch {
				case (partner != nil) && (v != partner):
					sexy = false
					break
				case !loc.Member(w.ToAbss(v, v.GeniShape())):
					sexy = false
					break
				default: //(either partner==nil or v==partner) and loc\in genishape
					partner = v
					sexy = true
				}
			case nil:
				continue
			default:
				sexy = false
				break
			}
		}
		if sexy {
			w.Intercourse(partner, mi)
			return false
		}
		//bite
		mi.SetFree(false)
		turnBack := false
		lBlocked := false
		rBlocked := false
		left := true
		var rp pos.Rel
		var newLoc pos.Abs
		var direc pos.Direc
		//var obj Obj (declared above)
		//var loc pos.Abs (declared above)
		var a, b int
		var outcome string
		exit := false
		for mi.GetJawStateSum() < mi.Xlen()-2 {
			//fmt.Println("in jawloop with", lBlocked, rBlocked)
			switch {
			case (!lBlocked) && (!rBlocked):
				left = util.RandBool()
			case (!lBlocked) && (rBlocked):
				left = true
			case (lBlocked) && (!rBlocked):
				left = false
			default:
				turnBack = true
				exit = true
			}
			if exit {
				exit = false
				break
			}
			if left {
				direc = pos.Rotate(w.Head[mi], -1)
				rp, _ = mi.JawPos()
				newLoc = pos.Next(w.ToAbs(mi, rp), direc)
			} else {
				direc = pos.Rotate(w.Head[mi], 1)
				_, rp = mi.JawPos()
				newLoc = pos.Next(w.ToAbs(mi, rp), direc)
			}
			obj = w.OccupiedBy[newLoc]
			//fmt.Println("got object")
			switch {
			case obj == nil:
				a, b = mi.GetJawState()
				if left {
					mi.SetJawState(a+1, b)
				} else {
					mi.SetJawState(a, b+1)
				}
				w.OccupiedBy[newLoc] = mi
			case !obj.Free():
				if left {
					lBlocked = true
				} else {
					rBlocked = true
				}
			default: //obj!=nil and obj.Free()
				outcome = w.Move(obj, pos.ToMoveTy(w.Head[obj], direc), 1, false, mi.JawForce(), mi.JawDamage())
				switch outcome {
				case "success":
					a, b = mi.GetJawState()
					if left {
						mi.SetJawState(a+1, b)
					} else {
						mi.SetJawState(a, b+1)
					}
					w.OccupiedBy[newLoc] = mi
				case "heavy":
					if left {
						lBlocked = true
					} else {
						rBlocked = true
					}
					continue
				case "crash":
					turnBack = true
					exit = true
				case "already dead":
					turnBack = true
					exit = true
				default:
					panic(errors.New("impossible jawmove outcome"))
				}
				if exit {
					exit = false
					break
				}
			}
		}
		if turnBack {
			a, b = mi.GetJawState()
			for i := 1; i <= a; i++ {
				loc = w.ToAbs(mi, pos.Rel{Z: -mi.Alen() + i, W: mi.Blen()})
				delete(w.OccupiedBy, loc)
			}
			for i := 1; i <= b; i++ {
				loc = w.ToAbs(mi, pos.Rel{Z: mi.Alen() - i, W: mi.Blen()})
				delete(w.OccupiedBy, loc)
			}
			mi.SetJawState(0, 0)
			mi.SetFree(true)
			return false
		}
		mi.SetFree(true)
		return true
	default:
		panic(errors.New("wrong jawMoveTy"))
	}
}
