package world

func reward(inf info) float64 {
	return inf.pleasure + 200*(inf.after.Hp-inf.before.Hp)
}
