package world

import (
	"github.com/daeseoklee/pixelworld/minion"
	"github.com/daeseoklee/pixelworld/pos"
	"github.com/daeseoklee/pixelworld/util"
)

//smell,emission

var (
	//NumPhero : number of pheromons
	NumPhero = 30

	//DimPhero : dimension of pheromon vector
	DimPhero = 50
)

var pheros = make([]Phero, NumPhero, NumPhero)

//Phero : pheromon struct
type Phero struct {
	index int
	vec   []float64
}

func newPhero(ind int) Phero {
	v := make([]float64, DimPhero)
	for i := 0; i < DimPhero; i++ {
		v[i] = util.RandNorm()
	}
	return Phero{index: ind, vec: v}
}

//ChoosePheros : randomly determine list of available pheromones
func ChoosePheros() {
	for i := 0; i < NumPhero; i++ {
		pheros[i] = newPhero(i)
	}
}

//SetPheros : ~
func SetPheros(ph [][]float64) {
	for i := 0; i < NumPhero; i++ {
		pheros[i] = Phero{index: i, vec: ph[i]}
	}
}

//GetVec : get pheromon vector from index
func GetVec(ind int) []float64 {
	return pheros[ind].vec
}

//dist : minimum distance between a pos.Abs and a Minion
func (w *World) dist(loc pos.Abs, mi *minion.Minion) float64 {
	var xMin, xMax, yMin, yMax, xDist, yDist int
	d := w.Head[mi].D
	if (d == "n") || (d == "s") {
		xMin, xMax = w.Loc[mi].X-mi.Alen(), w.Loc[mi].X+mi.Alen()
		yMin, yMax = w.Loc[mi].Y-mi.Blen(), w.Loc[mi].Y+mi.Blen()
	} else {
		xMin, xMax = w.Loc[mi].X-mi.Blen(), w.Loc[mi].X+mi.Blen()
		yMin, yMax = w.Loc[mi].Y-mi.Alen(), w.Loc[mi].Y+mi.Alen()
	}
	if (xMin <= loc.X) && (loc.X <= xMax) {
		xDist = 0
	} else {
		xDist = util.Min(util.Abs(xMin-loc.X), util.Abs(loc.X-xMax))
	}
	if (yMin <= loc.Y) && (loc.Y <= yMax) {
		yDist = 0
	} else {
		yDist = util.Min(util.Abs(yMin-loc.Y), util.Abs(loc.Y-yMax))
	}
	if xDist+yDist == 0 { //when loc is within the body of mi
		return 0.5
	}
	return float64(xDist + yDist)
}

//WriteSmell : write smell input in the iven slice
func (w *World) WriteSmell(mi *minion.Minion, here []float64) {
	a, b := mi.NosePoss()
	leftNose, rightNose := w.ToAbs(mi, a), w.ToAbs(mi, b)
	var dl, dr float64
	var v []float64
	for other := range w.Minions {
		if other != mi {
			dl, dr = w.dist(leftNose, other), w.dist(rightNose, other)
			v = GetVec(other.GetBehavior().Emission)
			for i := 0; i < DimPhero; i++ {
				here[i] += v[i] / dl
				here[DimPhero+i] += v[i] / dr
			}
		}
	}
}
