package minion

import (
	"errors"
	"math"

	"github.com/daeseoklee/pixelworld/colour"
	"github.com/daeseoklee/pixelworld/pos"
	"github.com/daeseoklee/pixelworld/util"
)

//Xlen :
func (mi *Minion) Xlen() int {
	return mi.trait.appear.xlen
}

//Alen :
func (mi *Minion) Alen() int {
	return (mi.trait.appear.xlen - 1) / 2
}

//Ylen :
func (mi *Minion) Ylen() int {
	return mi.trait.appear.ylen
}

//Blen :
func (mi *Minion) Blen() int {
	return (mi.trait.appear.ylen - 1) / 2
}

//Area :
func (mi *Minion) Area() int {
	return (mi.trait.appear.xlen) * (mi.trait.appear.ylen)
}

//BodyColour :
func (mi *Minion) BodyColour() colour.Colour {
	return mi.trait.appear.bodyColour
}

//GeniColour :
func (mi *Minion) GeniColour() colour.Colour {
	return mi.trait.appear.geniColour
}

//GeniLen :
func (mi *Minion) GeniLen() int {
	return util.SqrtVal(geniLen, mi.Area())
}

//GutXlen :
func (mi *Minion) GutXlen() float64 {
	return math.Pow(float64(mi.trait.appear.xlen), gutSize)
}

//GutYlen :
func (mi *Minion) GutYlen() float64 {
	return math.Pow(float64(mi.trait.appear.ylen), gutSize)
}

// GeniShape : Genitals shape
func (mi *Minion) GeniShape() []pos.Rel {
	l := make([]pos.Rel, mi.GeniLen())
	for j := 0; j < mi.GeniLen(); j++ {
		l[j] = pos.Rel{Z: 0, W: -mi.Blen() - j - 1}
	}
	return l
}

// BodyShape : body shape
func (mi *Minion) BodyShape() []pos.Rel {
	l := make([]pos.Rel, mi.Xlen()*mi.Ylen())
	for i := 0; i < mi.Xlen(); i++ {
		for j := 0; j < mi.Ylen(); j++ {
			l[j*mi.Xlen()+i] = pos.Rel{Z: -mi.Alen() + i, W: -mi.Blen() + j}
		}
	}
	return l
}

//NosePoss : the relative positions of noses
func (mi *Minion) NosePoss() (pos.Rel, pos.Rel) {
	return pos.Rel{Z: -mi.Alen(), W: 0}, pos.Rel{Z: mi.Alen(), W: 0}
}

//JawPos : the relative position of jaws
func (mi *Minion) JawPos() (pos.Rel, pos.Rel) {
	return pos.Rel{Z: -mi.Alen() + mi.jawState.a, W: mi.Blen()}, pos.Rel{Z: mi.Alen() - mi.jawState.b, W: mi.Blen()}
}

//JawPoss : between-jaw relative positions
func (mi *Minion) JawPoss() []pos.Rel {
	l := make([]pos.Rel, mi.Xlen()-2-mi.jawState.a-mi.jawState.b)
	for i := 1; i <= mi.Xlen()-2-mi.jawState.a-mi.jawState.b; i++ {
		l[i-1] = pos.Rel{Z: -mi.Alen() + mi.jawState.a + i, W: mi.Blen()}
	}
	return l
}

//JawOpen : whether jaw is completely open
func (mi *Minion) JawOpen() bool {
	if mi.jawState.a == 0 && mi.jawState.b == 0 {
		return true
	}
	if mi.GetJawStateSum() != mi.Xlen()-2 {
		panic(errors.New("strange jawState"))
	}
	return false
}

//GetJawState : return (jawState.a,jawState.b)
func (mi *Minion) GetJawState() (int, int) {
	return mi.jawState.a, mi.jawState.b
}

//GetJawStateSum : return jawState.a+jawState.b
func (mi *Minion) GetJawStateSum() int {
	return mi.jawState.a + mi.jawState.b
}

//SetJawState : set jawState
func (mi *Minion) SetJawState(a, b int) {
	mi.jawState.a = a
	mi.jawState.b = b
}

//JawForce : force when shutting jaw
func (mi *Minion) JawForce() int {
	return util.LinVal(jawForce, mi.Mass())
}

//JawDamage : damage when beaten by this
func (mi *Minion) JawDamage() int {
	return util.LinVal(jawDamage, mi.Mass())
}
