package world

import (
	"fmt"

	"github.com/daeseoklee/pixelworld/minion"
	"github.com/daeseoklee/pixelworld/pos"
	"github.com/daeseoklee/pixelworld/util"
)

//Intercourse : sexual intercourse
func (w *World) Intercourse(female, male *minion.Minion) {
	if female.Pregnant() {
		return
	}
	//what happens?
	male.Triple.Pleasure = 1.0
	female.Triple.Pleasure = 1.0
	if !female.Marriageable(male) {
		return
	}
	//the result
	female.SetHusbandTrait(male)
	female.SetChild(male)
	female.SetPregnant(true)
}

//ChildBirth : delievery. Return (#delievered,completeness,survival)
func (w *World) ChildBirth(mi *minion.Minion) (int, bool, bool) {
	w.SetVisible(mi, false)
	iniChild := mi.GetChild()
	finished := true
	for mi.GetChild() > 0 {
		babyTrait := mi.ChildTrait()
		baby := minion.Construct(babyTrait)
		direc := pos.DirecFromNum(util.RandInt(4))
		w.Register(baby, w.Loc[mi], direc, false)
		mi.HpReduce(baby.Mass())
		if mi.HpLack() {
			fmt.Println("death by delievery!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
			w.Kill(mi)
		}
		mi.ReduceChild()
		w.BirthDamage(mi)
		if w.Available(baby, w.Loc[baby], direc) {
			w.SetVisible(baby, true)
			w.BirthMove(baby)
		} else {
			w.Kill(baby)
			if mi.GetChild() > 0 {
				finished = false
			}
			break
		}
	}
	if finished {
		mi.SetPregnant(false)
	}
	//fmt.Println("objects?: ", w.Objects[mi])
	//fmt.Println("minion?:", w.Minions[mi])
	//fmt.Println("loc?:", w.Loc[mi])
	//fmt.Println("head?:", w.Head[mi])
	if w.Objects[mi] {
		if w.Available(mi, w.Loc[mi], w.Head[mi]) {
			w.SetVisible(mi, true)
			if finished {
				return iniChild, true, true
			}
			return iniChild - mi.GetChild(), false, true
		}
	}
	w.Kill(mi) //when not avilable
	if finished {
		return iniChild, true, false
	}
	return iniChild - mi.GetChild(), false, false
}
