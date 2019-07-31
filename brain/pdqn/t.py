import pdqn
net=pdqn.PDQN(3,3)
print(len(list(net.visual.parameters())))
print(len(list(net.olfactory.parameters())))
print(len(net.omega_opt.param_groups[0]['params']))