package minion

import "github.com/daeseoklee/pixelworld/util"

//Pregnant : pregnant?
func (mi *Minion) Pregnant() bool {
	return mi.pregnant
}

//SetPregnant : set pregnancy
func (mi *Minion) SetPregnant(b bool) {
	mi.pregnant = b
}

//SetChild : determine number of childs, given husband
func (mi *Minion) SetChild(husband *Minion) {
	mi.child = util.Poisson(int(numChild * mi.Compatibility(husband)))
}

//GetChild : get the child number in pregnancy
func (mi *Minion) GetChild() int {
	return mi.child
}

//ReduceChild : decrease # of child by 1
func (mi *Minion) ReduceChild() {
	mi.child--
}
