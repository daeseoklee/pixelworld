package world

import (
	"github.com/daeseoklee/pixelworld/minion"
	"github.com/daeseoklee/pixelworld/pos"
	"github.com/daeseoklee/pixelworld/util"
)

//Digest : digestion
func (w *World) Digest(obj Obj) {
	if !w.Objects[obj] {
		return
	}
	var mi *minion.Minion
	switch obj.(type) {
	case *minion.Minion:
		mi = obj.(*minion.Minion)
	default:
		return
	}
	//nutrition absorbing
	c := mi.GutXlen()
	a := int((c - 1) / 2)
	d := mi.GutYlen()
	b := int((d - 1) / 2)
	//fmt.Println("size: ", mi.Xlen(), "c:", c, "d: ", d, w.Loc[mi])
	var plantEat, animalEat int
	var totalPlantEat, totalAnimalEat int
	full := false
	var plant, animal int
	var ratio float64
	var loc pos.Abs
	for i := -a - 1; i <= a+1; i++ {
		for j := -b - 1; j < b+1; j++ {
			loc = w.ToAbs(mi, pos.Rel{Z: i, W: j})
			//fmt.Println(loc)
			plant, animal = w.Min[loc].p, w.Min[loc].a
			switch {
			case (plant == 0) && (animal == 0):
				continue
			case (i == -a-1) || (i == a+1):
				ratio = (c - 1 - 2*float64(a)) / 2
			case (j == -b-1) || (j == b+1):
				ratio = (d - 1 - 2*float64(b)) / 2
			default:
				ratio = 1
			}
			plantEat, animalEat = int(ratio*float64(plant)), int(ratio*float64(animal))
			if plantEat+animalEat >= mi.HpGap() {
				full = true
				plantEat = int(float64(mi.HpGap()*plant) / float64(plant+animal))
				animalEat = int(float64(mi.HpGap()*animal) / float64(plant+animal))
			}
			mi.HpIncrease(plantEat + animalEat)
			w.Min[loc].p -= plantEat
			w.Min[loc].a -= animalEat
			totalPlantEat += plantEat
			totalAnimalEat += animalEat
			if full {
				break
			}
		}
	}
	//childbirth
	if mi.Pregnant() && full {
		w.ChildBirth(mi)
	}
	//excretion
	w.Excrete(mi, int(mi.GetTaste()*float64(totalPlantEat)+(1-mi.GetTaste())*float64(totalAnimalEat)))

}

//Excrete : excrete given amount
func (w *World) Excrete(mi *minion.Minion, amount int) {
	if !w.Objects[mi] {
		return
	}
	//doesn't excrete in pregnancy
	if mi.Pregnant() {
		return
	}
	//prevent overflow
	if amount > mi.Mass() {
		amount = mi.Mass()
	}
	//nutrition distribution
	nutPerCell := amount / mi.GeniLen()
	remain := amount % mi.GeniLen()
	var loc pos.Abs
	for i, rp := range mi.GeniShape() {
		loc = w.ToAbs(mi, rp)
		if i < remain {
			w.Nut[loc] += nutPerCell + 1
		} else {
			w.Nut[loc] += nutPerCell
		}
	}
	//mass
	mi.HpReduce(amount)
	//death-by-excretion
	if mi.Mass() < mi.HpMin() {
		w.Kill(mi)
	}
}

//Attack : Minion damage
func (w *World) Attack(obj Obj, amount int) {
	if !w.Objects[obj] {
		return
	}
	var mi *minion.Minion
	switch obj.(type) {
	case *minion.Minion:
		mi = obj.(*minion.Minion)
	default:
		return
	}
	//fatal damage
	if amount > mi.Mass()-mi.HpMin() {
		w.Kill(mi)
		return
	}
	//damage
	mi.HpReduce(amount)
	//mineral distribution
	minPerCell := amount / mi.Area()
	remain := amount % mi.Area()
	l := w.Body(mi)
	for i, loc := range l {
		if i < remain {
			w.Min[loc].a += minPerCell + 1
		} else {
			w.Min[loc].a += minPerCell + 1
		}
	}
}

//BirthDamage : damage after each delievary
func (w *World) BirthDamage(mi *minion.Minion) {
	if !w.Objects[mi] { //need because of the death from delievery
		return
	}
	w.Attack(mi, util.LinVal(minion.BirthDamage, mi.Area()))
}

//Conversion : mineral conversion
func (w *World) Conversion() {
	var loc pos.Abs
	var nut int
	for i := 0; i < w.Xlen; i++ {
		for j := 0; j < w.Ylen; j++ {
			loc = pos.Abs{X: i, Y: j}
			nut = w.Nut[loc]
			w.Min[loc].p += nut
			w.Nut[loc] = 0
		}
	}
}
