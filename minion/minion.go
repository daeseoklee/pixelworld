package minion

import (
	"github.com/daeseoklee/pixelworld/colour"
	"github.com/daeseoklee/pixelworld/pos"
	"github.com/daeseoklee/pixelworld/util"
)

var (
	numOdor = 30
	dimOdor = 50

	numChild = 3.0
	life     = 200
	hpMax    = 100.0
	//BirthDamage : damage after each delievary
	BirthDamage = 10.0
	hpMin       = 30.0
	hpConsum    = 1.0
	spMax       = 5.0
	spRecharge  = 0.8
	spJaw       = 0.3
	spConsum    = []float64{0, 1, 1, 2, 2, 0.6, 0.6, 1}

	jawForce  = 10.0
	jawDamage = 1.0

	eyeColour  = colour.Colour{M: 0, R: 1, G: 1, B: 0}
	noseColour = colour.Colour{M: 0, R: 1, G: 0, B: 0}

	geniLen = 0.2
	gutSize = 0.6
)

//Appear :
type Appear struct {
	xlen       int
	ylen       int
	bodyColour colour.Colour
	geniColour colour.Colour
}

//ConstructAppear : Appear constructor
func ConstructAppear(xLen int, yLen int, bodyColour, geniColour colour.Colour) *Appear {
	return &Appear{xlen: xLen, ylen: yLen, bodyColour: bodyColour, geniColour: geniColour}
}

//Trait :
type Trait struct {
	appear *Appear
	taste  float64
}

//ConstructTrait : trait CONSTRUCTOR
func ConstructTrait(appear *Appear, taste float64) *Trait {
	return &Trait{appear: appear, taste: taste}
}

//JawState : minion jawstate
type JawState struct {
	a int
	b int
}

//Behavior : intended behavior
type Behavior struct {
	MoveTy    int
	MoveDist  int
	JawMoveTy int
	Emission  int
}

//Input : inputs
type Input struct {
	Vision   [][]colour.Colour
	Smell    []float64
	Hp       float64
	Sp       float64
	Pregnant float64
}

//Triplet : for Q-learning
type Triplet struct {
	Before   *Input
	Action   *Behavior
	After    *Input
	Pleasure float64
}

//--------------------------------------------------------------

//Minion :
type Minion struct {
	trait        *Trait
	id           int
	age          int
	hp           int
	sp           int
	life         int
	pregnant     bool
	behavior     *Behavior
	husbandTrait *Trait
	child        int
	visible      bool
	free         bool
	jawState     JawState
	Triple       *Triplet
}

//Construct : Minion constructor
func Construct(trait *Trait) *Minion {
	hpIni := util.LinVal(hpMax, trait.appear.xlen*trait.appear.ylen)
	spIni := util.SqrtVal(spMax, trait.appear.xlen*trait.appear.ylen)
	mi := Minion{trait: trait, id: -1, age: 0, life: -1, hp: hpIni, sp: spIni, husbandTrait: nil, pregnant: false, behavior: &Behavior{}, child: -1,
		visible: false, free: true, jawState: JawState{a: 0, b: 0},
		Triple: &Triplet{Before: &Input{Vision: make([][]colour.Colour, 3*trait.appear.xlen), Smell: make([]float64, 100), Hp: 0, Sp: 0, Pregnant: 0},
			Action:   &Behavior{MoveTy: 0, MoveDist: 0, JawMoveTy: 0, Emission: 0},
			After:    &Input{Vision: make([][]colour.Colour, 3*trait.appear.xlen), Smell: make([]float64, 100), Hp: 0, Sp: 0, Pregnant: 0},
			Pleasure: 0.0}}
	for i := 0; i < 3*trait.appear.ylen; i++ {
		mi.Triple.Before.Vision[i] = make([]colour.Colour, 3*trait.appear.ylen)
		mi.Triple.After.Vision[i] = make([]colour.Colour, 3*trait.appear.ylen)
	}
	mi.setLife()
	return &mi
}

//Minion As Obj interface + related ----------------------------------------

//Free : Minion is a free-moving object except temporarily in jaw movement
func (mi *Minion) Free() bool {
	return mi.free
}

//SetFree : set Minion.free
func (mi *Minion) SetFree(b bool) {
	mi.free = b
}

//Visible : Minion is visible when Minion.visible==true
func (mi *Minion) Visible() bool {
	return mi.visible
}

//SetVisible : set Minion.visible
func (mi *Minion) SetVisible(b bool) {
	mi.visible = b
}

//Shape : Minion's shape is as in the doc
func (mi *Minion) Shape() []pos.Rel {
	l := make([]pos.Rel, mi.Xlen()+2*mi.Ylen()-2+mi.GeniLen()+mi.jawState.a+mi.jawState.b)
	for i := -mi.Alen(); i <= mi.Alen(); i++ {
		l[mi.Alen()+i] = pos.Rel{Z: i, W: -mi.Blen()}
	}
	for j := -mi.Blen() + 1; j <= mi.Blen(); j++ {
		l[mi.Xlen()+mi.Blen()-1+j] = pos.Rel{Z: -mi.Alen(), W: j}
		l[mi.Xlen()+mi.Ylen()-2+mi.Blen()+j] = pos.Rel{Z: mi.Alen(), W: j}
	}
	for j := 0; j < mi.GeniLen(); j++ {
		l[mi.Xlen()+2*mi.Ylen()-2+j] = pos.Rel{Z: 0, W: -mi.Blen() - j - 1}
	}
	for i := 1; i <= mi.jawState.a; i++ {
		l[mi.Xlen()+2*mi.Ylen()-2+mi.GeniLen()+i-1] = pos.Rel{Z: -mi.Alen() + i, W: mi.Blen()}
	}
	for i := 1; i <= mi.jawState.b; i++ {
		l[mi.Xlen()+2*mi.Ylen()-2+mi.GeniLen()+mi.jawState.a+i-1] = pos.Rel{Z: mi.Alen() - i, W: mi.Blen()}
	}
	return l
}

//Mass : Minion's mass is equal to it's hp
func (mi *Minion) Mass() int {
	return mi.hp
}

//Render : Minion's color is as in the doc
func (mi *Minion) Render(p pos.Rel) colour.Colour {
	switch {
	case p == pos.Rel{Z: -mi.Alen(), W: mi.Blen()} || p == pos.Rel{Z: mi.Alen(), W: mi.Blen()}:
		return eyeColour
	case p == pos.Rel{Z: -mi.Alen(), W: 0} || p == pos.Rel{Z: mi.Alen(), W: 0}:
		return noseColour
	case p.Member(mi.GeniShape()):
		return mi.GeniColour()
	case p.Member(mi.Shape()):
		return mi.BodyColour()
	default:
		return colour.Colour{}
	}
}

//-------------------------------------------------

//GetID : return Minion.id
func (mi *Minion) GetID() int {
	return mi.id
}

//SetID : set Minion.id
func (mi *Minion) SetID(n int) {
	mi.id = n
}

//GetTaste : return taste
func (mi *Minion) GetTaste() float64 {
	return mi.trait.taste
}

//GetBehavior : return behavior
func (mi *Minion) GetBehavior() *Behavior {
	return mi.behavior
}

//SetBehavior : set behavior
func (mi *Minion) SetBehavior(moveTy, moveDist, jawMoveTy, emission int) {
	mi.behavior.MoveTy = moveTy
	mi.behavior.MoveDist = moveDist
	mi.behavior.JawMoveTy = jawMoveTy
	mi.behavior.Emission = emission
	mi.Triple.Action.MoveTy = moveTy
	mi.Triple.Action.MoveDist = moveDist
	mi.Triple.Action.JawMoveTy = jawMoveTy
	mi.Triple.Action.Emission = emission
}
