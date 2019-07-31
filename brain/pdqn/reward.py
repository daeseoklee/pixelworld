import cod

def reward(inf:cod.Info):
    return inf.pleasure+200*(inf.after.hp-inf.before.hp)