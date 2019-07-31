from pixelworld import *
import pygame
import time
#SPECIES=[(31,31,0.8,7),(15,15,0.6,14),(5,5,0.3,28),(3,3,0.0,56)]
SPECIES=[(7,7,0.5,7),(5,5,0.3,10),(3,3,0.0,20)]
A=10
def main():
    nut_map = dict()
    mineral_map = dict()
    for k in range(WORLD_SIZE[0]):
        for l in range(WORLD_SIZE[0]):
            nut_map[(k, l)] = 0
            mineral_map[(k, l)] = (0, 0)
    world = World(nut_map, mineral_map)
    for i in range(len(SPECIES)):
        color1=(float(uniformsample()),float(uniformsample()),float(uniformsample()),float(uniformsample()))
        color2=(float(uniformsample()),float(uniformsample()),float(uniformsample()),float(uniformsample()))
        appearance=MinionAppearance(size=(SPECIES[i][0],SPECIES[i][1]),color1=color1,color2=color2)
        props=MinionProperties(taste=SPECIES[i][2])
        brain=MinionNN().to(device)
        for p in brain.parameters():
            p.register_hook(lambda grad:torch.clamp(grad,-tensor(1.0),tensor(1.0)))
        trait=MinionTrait(appear=appearance,props=props,brain=brain)
        trait.register()
        for j in range(SPECIES[i][3]):
            minion=Minion(trait=trait)
            world.locate_and_register(minion,rand=True)
    world.render()
    screen = pygame.display.set_mode((WORLD_SIZE[0]*A,WORLD_SIZE[1]*A))
    screen.fill((0, 0, 0))
    running = 1
    def to_color(p):
        return (int(255*p[1]),int(255*p[2]),int(255*p[3]))
    def draw():
        for i in range(WORLD_SIZE[0]):
            for j in range(WORLD_SIZE[1]):
                for k in range(A):
                    for l in range(A):
                        screen.set_at((A*i+k,A*j+l),to_color(world.snapshot[i][j]))
    while running:
        event = pygame.event.poll()
        if event.type == pygame.QUIT:
            running = 0
        elif event.type == pygame.MOUSEBUTTONDOWN:
            if event.button==1:
                print("computing...")
                a = time.time()
                world.one_step()
                """
                
                try:
                    world.one_step()
                except:
                    print("error")
                    running=1
                    while running:
                        event=pygame.event.poll()
                        if event.type==pygame.MOUSEBUTTONDOWN:
                            print((event.pos[0] // A, event.pos[1] // A))
                        draw()
                        pygame.display.flip()
                """

                b = time.time()
                print(b - a)
        draw()
        pygame.display.flip()


main()