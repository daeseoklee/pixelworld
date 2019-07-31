package world

import (
	"github.com/daeseoklee/pixelworld/colour"
	"github.com/daeseoklee/pixelworld/pos"
)

//Obj : objects in the world
type Obj interface {
	Free() bool
	Visible() bool
	SetVisible(bool)
	Shape() []pos.Rel
	Mass() int
	Render(pos.Rel) colour.Colour
}

//outer boundary----------------------------------------------------------------------------------------

//Outer : the outer boundary of the world. Regarded as an Obj for convenience
type outer struct {
	Xlen int
	Ylen int
}

//Free : outer boundary is regarded not free-moving
func (outer outer) Free() bool {
	return false
}

//Visible : false in order to make World.SetVisible work
func (outer outer) Visible() bool {
	return false
}

//SetVisible : formality
func (outer outer) SetVisible(b bool) {
	return
}

//Shape : outer boundary shape(loc:(0,0),head:"n")
func (outer outer) Shape() []pos.Rel {
	l := make([]pos.Rel, 2*(outer.Xlen+outer.Ylen))
	for i := 0; i < outer.Xlen; i++ {
		l[i] = pos.Rel{Z: i, W: -1}
		l[outer.Xlen+i] = pos.Rel{Z: i, W: outer.Ylen}
	}
	for j := 0; j < outer.Ylen; j++ {
		l[2*outer.Xlen+j] = pos.Rel{Z: -1, W: j}
		l[2*outer.Xlen+outer.Ylen+j] = pos.Rel{Z: outer.Xlen, W: j}
	}
	return l
}

//Render : formality
func (outer outer) Render(p pos.Rel) colour.Colour {
	return colour.Colour{}
}

//Mass : formality
func (outer outer) Mass() int {
	return -1
}

//------------------------------------------------------------------------------

type wall struct {
	shape   []pos.Rel
	render  map[pos.Rel]colour.Colour
	visible bool
}

func (wall wall) Free() bool {
	return false
}

func (wall wall) Visible() bool {
	return wall.visible
}

func (wall wall) SetVisible(b bool) {
	wall.visible = b
}

func (wall wall) Shape() []pos.Rel {
	return wall.shape
}

func (wall wall) Render(rp pos.Rel) colour.Colour {
	return wall.render[rp]
}

func (wall wall) Mass() int {
	return -1
}
