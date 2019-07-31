package world

import (
	"github.com/daeseoklee/pixelworld/colour"
	"github.com/daeseoklee/pixelworld/minion"
	"github.com/daeseoklee/pixelworld/pos"
)

var (

	//Period : conversion period
	Period = 15

	//OuterColour : color of the external world(fire hell)
	OuterColour = colour.Colour{M: 0, R: 1, G: 0, B: 0}

	//LocateTry : number of iteration tried to locate an object
	LocateTry = 50

	//BirthForce : powerful forward movment at births
	BirthForce = 20.0

	//C : constant related to mineral colour
	C = 100.0

	//Traits : available traits
	Traits []*minion.Trait
)

//Mineral : (plant,animal)
type Mineral struct{ p, a int }

//---------------------------------------------------------

//World : the world
type World struct {
	Xlen        int
	Ylen        int
	Moment      int
	Outer       *outer
	Objects     map[Obj]bool
	Minions     map[*minion.Minion]bool
	Traits      []*minion.Trait
	OccupiedBy  map[pos.Abs]Obj
	Loc         map[Obj]pos.Abs
	Head        map[Obj]pos.Direc
	Min         map[pos.Abs]*Mineral
	Nut         map[pos.Abs]int
	SnapShot    [][]colour.Colour
	ID          int
	Title       string
	Kingdomname string
	Mapname     string
	Epoch       int
	Weightfrom  int
	Saveat      int
	Dolearn     []int
	Save        bool
	Train       bool
}
