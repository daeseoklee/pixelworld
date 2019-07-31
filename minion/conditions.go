package minion

import (
	"errors"

	"github.com/daeseoklee/pixelworld/util"
)

//IncreaseAge : age++
func (mi *Minion) IncreaseAge() {
	mi.age++
}

//setLife : determine lifespan based on trait
func (mi *Minion) setLife() {
	mi.life = -1
	//mi.life = util.Poisson(life)
}

//FullAge : check whether its the age to die
func (mi *Minion) FullAge() bool {
	return mi.age == mi.life
}

//HpGet : get HP
func (mi *Minion) HpGet() int {
	return mi.hp
}

//HpMax : HP capacity
func (mi *Minion) HpMax() int {
	if mi.pregnant {
		return (1 + mi.child) * util.LinVal(hpMax, mi.Area())
	}
	return util.LinVal(hpMax, mi.Area())
}

//HpLack : hp is too low
func (mi *Minion) HpLack() bool {
	return mi.hp < mi.HpMin()
}

//HpGap : Hp remains to achieve the max
func (mi *Minion) HpGap() int {
	return mi.HpMax() - mi.Mass()
}

//HpMin : HP lower limit
func (mi *Minion) HpMin() int {
	return util.LinVal(hpMin, mi.Area())
}

//HpConsum : HP comsumed per moment
func (mi *Minion) HpConsum() int {
	return util.LinVal(hpConsum, mi.Area())
}

//HpReduce : reduce HP and return whether it became lower than the HpMin
func (mi *Minion) HpReduce(amount int) bool {
	mi.hp -= amount
	return mi.hp < mi.HpMin()
}

//HpIncrease : increase hp and return whether it is full
func (mi *Minion) HpIncrease(amount int) bool {
	if mi.hp+amount >= mi.HpMax() {
		mi.hp = mi.HpMax()
		return true
	}
	mi.hp = mi.hp + amount
	return false
}

//SpGet : return stamina
func (mi *Minion) SpGet() int {
	return mi.sp
}

//SpMax : stamina capacity
func (mi *Minion) SpMax() int {
	return util.SqrtVal(spMax, mi.Area())
}

//SpMaxFromSize : ~
func SpMaxFromSize(x, y int) int {
	return util.SqrtVal(spMax, x*y)
}

//SpGap : stamina remains to achieve the max
func (mi *Minion) SpGap() int {
	return mi.SpMax() - mi.SpGet()
}

//SpConsum : per-move stamina need(if moveTy<=4 exact value, else per-sqrt(area) value), whether the move can be performed
func (mi *Minion) SpConsum(moveTy int) (int, bool) {
	var amount int
	switch {
	case moveTy == -1:
		amount = util.SqrtVal(spJaw, mi.Area())
	case moveTy <= 4:
		amount = int(spConsum[moveTy])
	case moveTy <= 7:
		amount = util.SqrtVal(spConsum[moveTy], mi.Area())
	default:
		panic(errors.New("invalid moveTy"))
	}
	return amount, mi.sp >= amount
}

//SpReduce : reduce sp if possible
func (mi *Minion) SpReduce(moveTy int) {
	amount, ok := mi.SpConsum(moveTy)
	if !ok {
		panic(errors.New("lack sp"))
	}
	mi.sp -= amount
}

//SpIncrease : increase sp
func (mi *Minion) SpIncrease() {
	amount := util.SqrtVal(spRecharge, mi.Area())
	if amount > mi.SpGap() {
		mi.sp = mi.SpMax()
	} else {
		mi.sp += amount
	}
}
