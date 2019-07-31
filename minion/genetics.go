package minion

import "errors"

//GetTrait : return trait
func (mi *Minion) GetTrait() *Trait {
	return mi.trait
}

//SetHusbandTrait : set husbandTrait
func (mi *Minion) SetHusbandTrait(husband *Minion) {
	mi.husbandTrait = husband.trait
}

//Compatibility : trait compatibility
func (mi *Minion) Compatibility(other *Minion) float64 {
	if mi.trait == other.trait {
		return 1
	}
	return 0
}

//Marriageable : can marry?
func (mi *Minion) Marriageable(other *Minion) bool {
	return mi.trait == other.trait
}

//ChildTrait : determine new child's trait based on its own and husbandTrait
func (mi *Minion) ChildTrait() *Trait {
	if mi.trait != mi.husbandTrait {
		panic(errors.New("have married with other species"))
	}
	return mi.trait
}
