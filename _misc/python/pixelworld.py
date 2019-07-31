import os
import random
from math import *
import torch
import torch.optim
gpu=False
if torch.cuda.is_available():
    cuda=torch.device("cuda")
    gpu=True
    print("gpu is available:",torch.cuda.get_device_name(0))
device=cuda if gpu else torch.device("cpu")
print("device: ",device)
import torch.nn as nn
import torch.distributions as d

def one(size,index):
    v=torch.zeros(size).to(device)
    v[index]+=1.0
    return v
def tensor(a):
    return torch.tensor(data=a,device=device)
def zero():
    return torch.tensor(data=0.0,device=device)


def betasample(a,b):
    return d.beta.Beta(a,b).sample()
def betalogpdf(a,b,x):
    return d.beta.Beta(a,b).log_prob(x)
def uniformsample():
    return d.uniform.Uniform(0,1).sample()
def poissonsample(a):
    return d.poisson.Poisson(a).sample()
def discretesample(l):
    if not type(l)==torch.Tensor:
        l=tensor(l)
    return d.categorical.Categorical(l).sample()

def beta0sample(mu,sigma):
    s = mu * (1 - mu) / sigma ** 2
    a = mu * (s - 1)
    b = (1 - mu) * (s - 1)
    if s>1:
        return betasample(a,b)
    else:
        return mu.detach()
def beta0logpdf(mu,sigma,x):
    s = mu * (1 - mu) / sigma ** 2
    a = mu * (s - 1)
    b = (1 - mu) * (s - 1)
    if s>1:
        return betalogpdf(a,b,x)
    else:
        return zero()
BETA_SD_CONST=0.25

#helpers---------------------------------------------------------------------
def plus(u,v):
    return (u[0]+v[0],u[1]+v[1])
def minus(u,v):
    return (u[0]-v[0],u[1]-v[1])
def next(u,direc):
    if direc=="n":
        return (u[0],u[1]-1)
    elif direc=="w":
        return (u[0]-1,u[1])
    elif direc=="s":
        return (u[0],u[1]+1)
    elif direc=="e":
        return (u[0]+1,u[1])
def previous(u,direc):
    if direc=="n":
        return (u[0],u[1]+1)
    elif direc=="w":
        return (u[0]+1,u[1])
    elif direc=="s":
        return (u[0],u[1]-1)
    elif direc=="e":
        return (u[0]-1,u[1])

DIRECTIONS=["n","w","s","e"]
def direc2num(s):
    if s=="n":
        return 0
    if s=="w":
        return 1
    if s=="s":
        return 2
    if s=="e":
        return 3
def num2direc(n):
    m=n%4
    if m==0:
        return "n"
    if m==1:
        return "w"
    if m==2:
        return "s"
    if m==3:
        return "e"
def rotate(v,n):
    m=n%4
    if m==0:
        return v
    return rotate((v[1],-v[0]),m-1)

def pos_plus_poslist(u,l):
    r=[]
    for v in l:
        r.append(plus(u,v))
    return r

#global constants------------------------------------------------------------------


WORLD_SIZE=(100,80)
CONVERSION_PERIOD=30
OUTER_WORLD_COLOR=(0.0,1.0,0.0,0.0) #FIRE HELL!!!
NUM_ODORS=30
DIM_ODORS=50
NUM_NOSES=2


NUM_CHILD=3
LIFE_UPPER_BOUND=1000
HP_MAX_CONST=100
HP_MIN_CONST=30
STAMINA_MAX_CONST=5.0
STAMINA_RECH_CONST=0.8
STAMINA_JAW_CONST=0.3
STAMINA_CONSUM_CONSTS=[0,1,1,2,2,0.6,0.6,1.0]

#MOVE_TYPES=["s","f","b","l","r","a","c","t"] #still, forward, backward, left, right, anti-clockwise, ,clocwise, turn180
JAWPOWER_PER_MASS=10.0
JAWDAMAGE_PER_MASS=1.0

DIM_STATE=300
VISION=(16,16)  #96,96
VISUAL_CLASSES=30
SMELL_CLASSES=20

REPLAY_MEMORY_CAPACITY=100

GUT_SIZE_CONST=0.6
GENITALS_LENGTH_CONST=0.2

EYE_COLOR=(0,1,1,0)
NOSE_COLOR=(0,1,0,0)
#other helpers----------------------------------------------

def mineral_color(a):
    plant,animal=a
    C=100
    return (tanh((animal+plant)/C),tanh((animal)/C),tanh((plant)/C),0.0)

def gut_size(size):
    return size[0]**GUT_SIZE_CONST,size[1]**GUT_SIZE_CONST

def genitals_length(size):
    return max(1,int(sqrt(size[0]*size[1])*GENITALS_LENGTH_CONST))

num_species=0
species=[]
def get_species_id():
    global num_species
    num_species+=1
    return num_species-1

#---------------------------------------------------------------

class MinionAppearance():
    def __init__(self,size,color1,color2):
        self.size=size
        self.color1=color1
        self.color2=color2
    def mixed_with(self,another):
        pass
class MinionProperties():
    def __init__(self,taste):
        self.taste=taste
    def mixed_with(self,another):
        pass
class MinionNN(nn.Module):
    def __init__(self):
        super(MinionNN,self).__init__()
        self.species_id=None
        #NN components
        self.visual_attention=nn.Sequential(
            nn.Linear(DIM_STATE,VISION[0]*VISION[1]),
            nn.Sigmoid()
        )
        self.visual_net=nn.Sequential(
            nn.Conv2d(4,10,5,stride=1,padding=2),
            nn.MaxPool2d(kernel_size=2,stride=2),
            nn.BatchNorm2d(10),
            nn.ReLU(),
            nn.Conv2d(10,20,5,stride=1,padding=2),
            nn.MaxPool2d(kernel_size=2,stride=2),
            nn.ReLU(),
            nn.BatchNorm2d(20),
            nn.Conv2d(20,40,5,stride=1,padding=2),
            nn.MaxPool2d(kernel_size=2,stride=2),
            nn.BatchNorm2d(40),
            nn.ReLU()
        )
        self.POST_V1=(VISION[0]*VISION[1]*40)//64 #at the end of convolutional layers
        self.visual_projection=nn.Sequential(
            nn.Linear(self.POST_V1,VISUAL_CLASSES)
        )#+normalization
        self.olfactory_attention=nn.Sequential(
            nn.Linear(DIM_STATE,NUM_NOSES*DIM_ODORS),
            nn.Sigmoid()
        )
        self.olfactory_net=nn.Sequential(
            nn.Linear(NUM_NOSES*DIM_ODORS,NUM_NOSES*DIM_ODORS),
            nn.ReLU(),
            nn.Linear(NUM_NOSES*DIM_ODORS,NUM_NOSES*DIM_ODORS),
            nn.ReLU(),
            nn.Linear(NUM_NOSES*DIM_ODORS,SMELL_CLASSES),
        ) #+normalization
        self.POST_SENSORY=VISUAL_CLASSES+SMELL_CLASSES+3+DIM_STATE
        self.state_update=nn.Sequential(
            nn.Linear(self.POST_SENSORY,DIM_STATE),
            nn.Softmax()
        )

        self.PREMOTOR=self.POST_SENSORY//2
        self.premotor=nn.Sequential(
            nn.Linear(self.POST_SENSORY,self.PREMOTOR),
            nn.ReLU()
        )
        MOVE_TYPES=8
        BITE_TYPES=3
        self.move_type=nn.Sequential(
            nn.Linear(self.PREMOTOR,MOVE_TYPES),
            nn.Softmax()
        )
        self.move_amount_mean=nn.Sequential(
            nn.Linear(self.PREMOTOR,1),
            nn.Sigmoid()
        )
        self.move_amount_sd=nn.Sequential(
            nn.Linear(self.PREMOTOR,1),
            nn.Sigmoid()
        )
        self.bite=nn.Sequential(
            nn.Linear(self.PREMOTOR,BITE_TYPES),
            nn.Softmax()
        )
        self.emission=nn.Sequential(
            nn.Linear(self.PREMOTOR,NUM_ODORS),
            nn.Softmax()
        )
    def forward(self,state,image,smell,mass,stamina,pregnancy,train=False,out=None): #if train, the return value is log probability of the out
        assert state.size()[0]==image.size()[0]==smell.size()[0]==mass.size()[0]==stamina.size()[0]==pregnancy.size()[0]
        N=state.size()[0]
        visual_input=image*self.visual_attention(state).repeat(1,4).reshape(N,4,VISION[0],VISION[1])
        smell_input=smell*self.olfactory_attention(state)
        post_visual=nn.functional.normalize(self.visual_projection(self.visual_net(visual_input).reshape(N,-1)))
        post_olfactory=nn.functional.normalize(self.olfactory_net(smell_input))
        post_sensory=torch.cat((post_visual,post_olfactory,mass,stamina,pregnancy,state),dim=1)
        state_probs=self.state_update(post_sensory)
        premotor=self.premotor(post_sensory)
        move_ty_probs=self.move_type(premotor)
        move_amount_mean=self.move_amount_mean(premotor).reshape(N,1)
        move_amount_sd=self.move_amount_sd(premotor).reshape(N,1)
        bite_probs=self.bite(premotor)
        emission_probs=self.emission(premotor)
        if train:
            state_sample,move_ty_sample,move_amount_sample,bite_sample,emission_sample=out
            return torch.log(state_probs.gather(1,state_sample.reshape(N,1)))+torch.log(move_ty_probs.gather(1,move_ty_sample.reshape(N,1)))\
                   +tensor(list(map(lambda p:beta0logpdf(p[0],p[1],p[2]),zip(list(move_amount_mean),list(move_amount_sd),list(move_amount_sample)))))\
                   +torch.log(bite_probs.gather(1,bite_sample.reshape(N,1)))+torch.log(emission_probs.gather(1,emission_sample.reshape(N,1)))
        else:
            state_sample=discretesample(state_probs)
            move_ty_sample=discretesample(move_ty_probs)
            move_amount_sample=tensor(list(map(lambda p:beta0sample(p[0],p[1]),zip(list(move_amount_mean),list(move_amount_sd)))))
            bite_sample=discretesample(bite_probs)
            emission_sample=discretesample(emission_probs)
            return (state_sample,move_ty_sample,move_amount_sample,bite_sample,emission_sample)


    def mixed_with(self,another):
        pass

class MinionTrait():
    def __init__(self,appear,props,brain):
        self.appear=appear
        self.props=props
        self.brain=brain
        self.species_id=None
    def register(self):
        self.species_id=get_species_id()
        self.brain.species_id=self.species_id
        global species
        species.append(self)

    def mixed_With(self,another):
        return MinionTrait(self.appear.mixed_with(another.appear),self.props.mixed_with(another.props),self.brain.mixed_with(another.brain))

class Obj():
    def __init__(self,mass,free_moving,nontype=True): #nontype means it does not belong to specific subtype(e.g. Minion), thus does not render
        self.nontype=nontype
        self.mass=mass
        self.free_moving=free_moving
        self.visible=False
        self.pos=None
        self.head=None #"w","e","n","s
        self.max_stamina=None
        self.stamina_recharge=None
        self.pregnant=None
        self.taste=None
    def get_max_mass(self):
        return -1
    def relpos2abspos(self,pos):
        return plus(self.pos,rotate(pos,direc2num(self.head)))
    def direc(self,movedirec):
        if movedirec==1: #forward
            return self.head
        elif movedirec==2: #backward
            return num2direc(direc2num(self.head)+2)
        elif movedirec==3: #left
            return num2direc(direc2num(self.head)+1)
        elif movedirec==4: #right
            return num2direc(direc2num(self.head)+3)
    def movedirec(self,direc):
        num=(direc2num(direc)-direc2num(self.head))%4
        if num==0:
            return 1
        elif num==1:
            return 3
        elif num==2:
            return 2
        elif num==3:
            return 4
    def rel_occupying_area(self):
        return []
    def occupying_area(self):
        l=[]
        for pos in self.rel_occupying_area():
            l.append(self.relpos2abspos(pos))
        return l
    def render(self):
        return None
class Minion(Obj):
    def __init__(self,trait,no_brain=False):
        super(Minion,self).__init__(mass=HP_MAX_CONST*trait.appear.size[0]*trait.appear.size[1],free_moving=True,nontype=False)
        self.head=random.choice(DIRECTIONS)

        self.no_brain=no_brain
        self.species_id=trait.species_id

        self.size=trait.appear.size
        self.a=(self.size[0]-1)//2
        self.b=(self.size[1]-1)//2
        self.area=self.size[0]*self.size[1]
        self.gut_size=gut_size(self.size)
        self.genitals_length=genitals_length(self.size)
        self.taste = trait.props.taste
        self.color1=trait.appear.color1
        self.color2=trait.appear.color2

        self.max_mass=HP_MAX_CONST*self.area
        self.min_mass=HP_MIN_CONST*self.area
        self.num_child=-1
        self.stamina=max(1,int(STAMINA_MAX_CONST*sqrt(self.area)))
        self.max_stamina=max(1,int(STAMINA_MAX_CONST*sqrt(self.area)))
        self.stamina_recharge=max(1,int(STAMINA_RECH_CONST*sqrt(self.area)))
        self.stamina_jaw=max(1,int(STAMINA_JAW_CONST*sqrt(self.area)))
        self.pregnant=False
        self.age=0
        self.jaw_state=(0,0) #0~self.size[0]-2

        self.brain=trait.brain

        self.state=random.choice(range(DIM_STATE))
        self.move_ty=random.choice(range(7))
        self.move_distance=1
        self.bite=0
        self.emission=random.choice(range(NUM_ODORS))

        self.replay_memory_capacity=REPLAY_MEMORY_CAPACITY
        self.replay_memory=[]
        self.lr=0.00001
    def get_max_age(self):
        return LIFE_UPPER_BOUND
    def policy_update(self,vision_input,smell_input,new_state,move_ty,move_amount,bite,new_emission): #all in tensor
        new=((one(DIM_STATE,self.state),vision_input,smell_input,tensor([self.mass/self.get_max_mass()]),tensor([self.stamina/self.max_stamina])
              ,tensor([1.0 if self.pregnant else 0.0]))
                                     ,(new_state,move_ty,move_amount,bite,new_emission))
        if len(self.replay_memory)<self.replay_memory_capacity:
            self.replay_memory.append(new)
        else:
            self.replay_memory.pop(0)
            self.replay_memory.append(new)
        self.state=int(new_state)
        self.move_ty=int(move_ty)
        self.move_distance=self.max_stamina*move_amount
        self.bite=int(bite)
        self.emission=new_emission

    def get_trait(self):
        return species[self.species_id]
    def rel_occupying_area(self):
        l=[]
        for i in [-self.a,self.a]:
            for j in range(-self.b,self.b+1):
                l.append((i,j))
        for i in range(-self.a+1,self.a):
            l.append((i,self.b))
            if i<=-self.a+self.jaw_state[0] or i>=self.a-self.jaw_state[1]:
                l.append((i,-self.b))
        for j in range(1,self.genitals_length+1):
            l.append((0,self.b+j))
        return l
    def get_num_child(self):
        self.num_child=int(poissonsample(NUM_CHILD))
        return self.num_child
    def get_max_mass(self):
        if self.pregnant:
            return self.max_mass*(1+self.num_child)
        return self.max_mass

    def locate(self,pos,direc):
        self.pos=pos
        self.head=direc
    def jaw_pos(self,i):
        return self.relpos2abspos((-self.a+i,-self.b))
    def noses_rel_pos(self):
        return [(-self.a,0),(self.a,0)]
    def noses_pos(self):
        return [self.relpos2abspos((-self.a,0)),self.relpos2abspos((self.a,0))]
    def eyes_rel_pos(self):
        return [(-self.a, -self.b),(self.a, -self.b)]
    def eyes_pos(self):
        return [self.relpos2abspos((-self.a,-self.b)),self.relpos2abspos((self.a,-self.b))]
    def genitals_rel_pos(self):
        l=[]
        for i in range(self.genitals_length):
            l.append((0,self.b+1+i))
        return l
    def genitals_pos(self):
        l=[]
        for i in range(self.genitals_length):
            l.append(self.relpos2abspos((0,self.b+1+i)))
        return l
    def occupying_rectangle(self):
        l=[]
        for i in range(-self.a,self.a+1):
            for j in range(-self.b,self.b+1):
                l.append(self.relpos2abspos((i,j)))
        return l

    def render(self):
        d=dict()
        for pos in self.rel_occupying_area():
            if pos in self.noses_rel_pos():
                d[self.relpos2abspos(pos)]=NOSE_COLOR
            elif pos in self.eyes_rel_pos():
                d[self.relpos2abspos(pos)]=EYE_COLOR
            elif pos in self.genitals_rel_pos():
                d[self.relpos2abspos(pos)]=self.color2
            else:
                d[self.relpos2abspos(pos)]=self.color1
        return d





class World():
    def __init__(self,nut_map,mineral_map):   #nut_map : dict that represents nutrition distribution
        self.size=WORLD_SIZE
        self.moment=1
        self.nut_map=nut_map
        self.mineral_map=mineral_map
        self.occupy_map=dict() #value : None or Obj
        self.border=Obj(mass=0,free_moving=False)
        for i in range(self.size[0]):
            for j in range(self.size[1]):
                self.occupy_map[(i,j)]=None
        for i in range(-100,0):
            for j in range(-100,self.size[1]+100):
                self.occupy_map[(i,j)]=self.border
        for i in range(self.size[0],self.size[0]+100):
            for j in range(-100,self.size[1]+100):
                self.occupy_map[(i,j)]=self.border
        for i in range(-100,self.size[0]+100):
            for j in range(-100,0):
                self.occupy_map[(i,j)]=self.border
        for i in range(-100,self.size[0]+100):
            for j in range(self.size[1],self.size[1]+100):
                self.occupy_map[(i,j)]=self.border
        self.odors = []
        for i in range(NUM_ODORS):
            self.odors.append(torch.randn(DIM_ODORS).to(device))
        self.minions=[]
        self.objects=[self.border]  #objects except minions
        self.species_ids=[]
        self.species_id2entities=dict()
        self.species_id2opt=dict()
        self.snapshot=None
        self.render()
    def in_range(self,pos):
        return (0<=pos[0]<self.size[0]) and (0<=pos[1]<self.size[1])
    def vision_input(self,mi):
        raw_data=[[None for j in range(3*mi.size[1])] for i in range(3*mi.size[0])]
        for i in range(-3*mi.a-1,3*mi.a+2):
            for j in range(-5*mi.b-2,mi.b+1):
                pos=mi.relpos2abspos((i,j))
                raw_data[i+3*mi.a+1][j+5*mi.b+2]=list(self.snapshot[pos[0]][pos[1]]) if self.in_range(pos) else list(OUTER_WORLD_COLOR)
        data=[[raw_data[int(i*3*mi.size[0]/VISION[0])][int(i*3*mi.size[1]/VISION[1])] for j in range(VISION[1])] for i in range(VISION[0])]
        return tensor(data).permute(2,0,1)
    def smell_input(self,mi:Minion):
        def min_taxi_dist(another:Minion,pos):
            if another.head=="n" or "s":
                x_min,x_max=another.pos[0]-another.a,another.pos[0]+another.a
                y_min,y_max=another.pos[1]-another.b,another.pos[1]+another.b
            elif another.head=="w" or "e":
                x_min,x_max=another.pos[0]-another.b,another.pos[0]+another.b
                y_min,y_max=another.pos[1]-another.a,another.pos[1]+another.a
            return (0 if x_min<=pos[0]<=x_max else min(abs(x_min-pos[0]),abs(x_max-pos[0])))+ \
                   (0 if y_min<=pos[1]<=y_max else min(abs(y_min-pos[1]),abs(y_max-pos[1])))
        odor_strength_left=[0]*NUM_ODORS
        odor_strength_right=[0]*NUM_ODORS
        for another in self.minions:
            if another!=mi:
                odor_strength_left[another.emission]+=1/(1+min_taxi_dist(another,mi.noses_pos()[0]))
                odor_strength_right[another.emission]+=1/(1+min_taxi_dist(another,mi.noses_pos()[1]))
        smell=torch.zeros(2*DIM_ODORS).to(device)
        for i in range(NUM_ODORS):
            smell+=torch.cat([odor_strength_left[i]*self.odors[i],odor_strength_right[i]*self.odors[i]])
        return smell*sqrt(2*DIM_ODORS)/torch.norm(smell)

    def schedule_update(self):
        #compute collectively
        for id in self.species_ids:
            st=[]
            im=[]
            sm=[]
            ma=[]
            sp=[]
            pr=[]
            mis=self.species_id2entities[id]
            for mi in mis:
                st.append(one(DIM_STATE,mi.state))
                im.append(self.vision_input(mi))
                sm.append(self.smell_input(mi))
                ma.append(tensor([mi.mass/mi.get_max_mass()]))
                sp.append(tensor([mi.stamina/mi.max_stamina]))
                pr.append(tensor([1.0 if mi.pregnant else 0.0]))

            st=torch.stack(st)
            im=torch.stack(im)
            sm=torch.stack(sm)
            ma=torch.stack(ma)
            sp=torch.stack(sp)
            pr=torch.stack(pr)
            net=species[id].brain
            new_states,move_tys,move_amounts,bites,new_emissions=\
                net.forward(state=st,image=im,smell=sm,mass=ma,stamina=sp,pregnancy=pr,train=False)
            for i,mi in enumerate(mis):
                mi.policy_update(vision_input=im[i],smell_input=sm[i],new_state=new_states[i]
                                 ,move_ty=move_tys[i],move_amount=move_amounts[i],bite=bites[i],new_emission=new_emissions[i])

    def groupft(self,id):
        mis=self.species_id2entities[id]
        ft=0
        for mi in mis:
            ft+=mi.mass/(mi.area*HP_MAX_CONST)
        return ft


    def learn_from_groupft(self):
        for id in self.species_ids:
            opt=self.species_id2opt[id]
            mis=self.species_id2entities[id]
            net=mis[0].brain
            state=[]
            image=[]
            smell=[]
            mass=[]
            stamina=[]
            pregnancy=[]

            state_sample=[]
            move_ty_sample=[]
            move_amount_sample=[]
            bite_sample=[]
            emission_sample=[]

            for mi in mis:
                ((a, b, c, d, e, f), (g, h, i, j, k))=random.choice(mi.replay_memory)
                state.append(a)
                image.append(b)
                smell.append(c)
                mass.append(d)
                stamina.append(e)
                pregnancy.append(f)

                state_sample.append(g)
                move_ty_sample.append(h)
                move_amount_sample.append(i)
                bite_sample.append(j)
                emission_sample.append(k)
            state=torch.stack(state)
            image=torch.stack(image)
            smell=torch.stack(smell)
            mass=torch.stack(mass)
            stamina=torch.stack(stamina)
            pregnancy=torch.stack(pregnancy)

            state_sample=torch.stack(state_sample)
            move_ty_sample=torch.stack(move_ty_sample)
            move_amount_sample=torch.stack(move_amount_sample)
            bite_sample=torch.stack(bite_sample)
            emission_sample=torch.stack(emission_sample)
            log_prob=net.forward(state,image,smell,mass,stamina,pregnancy,train=True,
                                     out=(state_sample,move_ty_sample,move_amount_sample,bite_sample,emission_sample)).sum()
            loss=self.groupft(id)*log_prob
            opt.zero_grad()
            loss.backward()
            opt.step()

    def learn(self):
        self.learn_from_groupft()

    def register_minion(self,mi:Minion):
        assert (not mi.visible)
        assert self.available(mi.occupying_area())
        #find order
        i=0
        while i<len(self.minions) and self.minions[i].mass>=mi.mass:
            i+=1
        self.minions.insert(i,mi)
        #make visible
        self.mk_visible(mi)
        #other class variables
        if mi.species_id not in self.species_ids:
            self.species_ids.append(mi.species_id)
            self.species_id2entities[mi.species_id]=[mi]
            self.species_id2opt[mi.species_id]=torch.optim.SGD(params=mi.brain.parameters(),lr=mi.lr)
        else:
            self.species_id2entities[mi.species_id].append(mi)

    def find_available(self,relposlist,rand=True,pos=None,head=None):
        if not rand:
            return pos
        x_list=list(map(lambda p:p[0],relposlist))
        y_list=list(map(lambda p:p[1],relposlist))
        x_min,x_max=min(x_list),max(x_list)
        y_min,y_max=min(y_list),max(x_list)
        if random:
            for i in range(1000):
                pos=(random.choice(range(max(0,-x_min),self.size[0]-max(0,x_max))),random.choice(range(max(0,-y_min),self.size[1]-max(0,y_max))))
                if self.available(pos_plus_poslist(pos,relposlist)):
                    return pos
            return None

    def locate_and_register(self,mi:Minion,rand=True,pos=None,head=None):
        if not rand:
            mi.head=head
        pos=self.find_available(list(map(lambda v:rotate(v,direc2num(mi.head)),mi.rel_occupying_area())),rand,pos,head)
        if pos==None:
            return False
        mi.pos=pos
        self.register_minion(mi)
        return True

    def available(self,poslist):
        can=True
        for pos in poslist:
            if self.occupy_map[pos]!=None:
                can=False
                break
        return can

    def mk_invisible(self,obj:Obj):
        obj.visible=False
        for pos in obj.occupying_area():
            self.occupy_map[pos]=None
    def mk_visible(self,obj:Obj):
        assert (obj.visible==False) and self.available(obj.occupying_area())
        obj.visible=True
        for pos in obj.occupying_area():
            self.occupy_map[pos]=obj
    def attack(self,mi:Minion,amount):
        if amount>=mi.mass:
            self.kill(mi)
            return
        #mineral distribution
        mineral_per_cell=amount//mi.area
        remain=amount%mi.area
        l=mi.occupying_rectangle()
        for i in range(len(l)):
            if i<remain:
                self.mineral_map[l[i]]=(self.mineral_map[l[i]][0],self.mineral_map[l[i]][1]+mineral_per_cell+1)
            else:
                self.mineral_map[l[i]]=(self.mineral_map[l[i]][0],self.mineral_map[l[i]][1]+mineral_per_cell)
        #damage
        mi.mass-=amount
        #death
        if mi.mass<mi.min_mass:
            self.kill(mi)
    def kill(self,mi:Minion):
        if not mi in self.minions:
            return
        #mineral distribution
        mineral_per_cell=mi.mass//mi.area
        remain=mi.mass%mi.area
        l=mi.occupying_rectangle()
        for i in range(len(l)):
            if i<remain:
                self.mineral_map[l[i]]=(self.mineral_map[l[i]][0],self.mineral_map[l[i]][1]+mineral_per_cell+1)
            else:
                self.mineral_map[l[i]]=(self.mineral_map[l[i]][0],self.mineral_map[l[i]][1]+mineral_per_cell)
        #occupy_map adjustment
        if mi.visible: #temporary invisibles can be killed
            for pos in mi.occupying_area():
                self.occupy_map[pos]=None
        #removing from the world
        self.minions.remove(mi)
        if len(self.species_id2entities[mi.species_id])==1:
            self.species_id2entities.__delitem__(mi.species_id)
            self.species_id2opt.__delitem__(mi.species_id)
            self.species_ids.remove(mi.species_id)
        else:
            self.species_id2entities[mi.species_id].remove(mi)
    def increase_ages(self):
        for mi in self.minions:
            mi.age+=1
            if mi.age==mi.get_max_age():
                self.kill(mi)
    def mineral_conversion(self):
        for i in range(self.size[0]):
            for j in range(self.size[1]):
                self.mineral_map[(i,j)]=(self.mineral_map[(i,j)][0]+self.nut_map[(i,j)],self.mineral_map[(i,j)][1])
                self.nut_map[(i,j)]=0
    def excrete(self,mi,amount=None): #synonom of damaging
        if not mi in self.minions:
            return
        if amount==None:
            amount=mi.area
        if amount>mi.mass:
            amount=mi.mass
        #doesn't excrete in pregnancy
        if mi.pregnant:
            return
        #nutrition distribution
        l=mi.genitals_pos()
        nut_per_cell=amount//len(l)
        remain=amount%len(l)
        for i in range(len(l)):
            if i<remain:
                self.nut_map[l[i]]+=nut_per_cell+1
            else:
                self.nut_map[l[i]]+=nut_per_cell
        #mass management
        mi.mass-=amount
        #death-by-excretion
        if mi.mass<mi.min_mass:
            self.kill(mi)
    def digest(self,mi:Obj): #each digestion is done within self.move
        if (not type(mi)==Minion) or (not mi in self.minions):
            return
        #nutrition distribution
        c,d=mi.gut_size
        a,b=int((c-1)/2),int((d-1)/2)
        total_plant_intake,total_animal_intake=0,0
        full=False
        for i in range(-a-1,a+2):
            for j in range(-b-1,b+2):
                plant,animal=self.mineral_map[mi.relpos2abspos((i,j))]
                if plant+animal==0:
                    continue
                if i in [-a-1,a+2]:
                    eat=(c-1-2*a)/2
                elif j in [-b-1,b+2]:
                    eat=(d-1-2*b)/2
                else:
                    eat=1.0
                plant_intake=int(eat*plant)
                animal_intake=int(eat*animal)
                if mi.get_max_mass()-mi.mass<=plant_intake+animal_intake:
                    plant_intake=int((mi.get_max_mass()-mi.mass)*plant/(plant+animal))
                    animal_intake=int((mi.get_max_mass()-mi.mass)*animal/(plant+animal))
                    full=True
                mi.mass+=plant_intake+animal_intake
                p=self.mineral_map[mi.relpos2abspos((i, j))]
                self.mineral_map[mi.relpos2abspos((i, j))]=(p[0]-plant_intake,p[1]-animal_intake)
                total_plant_intake+=plant_intake
                total_animal_intake+=animal_intake
                if full:
                    break
        #childbirth management
        if mi.pregnant and full:
            self.childbirth(mi)
        #excretion
        self.excrete(mi,amount=int(mi.taste*total_plant_intake+(1-mi.taste)*total_animal_intake))
        #order update
        i=self.minions.index(mi)
        j=i
        while j>0:
            if self.minions[j-1].mass<mi.mass:
                j-=1
            else:
                break
        if j<i:
            self.minions.remove(mi)
            self.minions.insert(j,mi)



    def bundle(self,mi,direc,exclude=[]):
        adjacents=[]
        for pos in mi.occupying_area():
            nextpos=next(pos,direc)
            obj=self.occupy_map[nextpos]
            if (not nextpos in mi.occupying_area()) and (obj!=None) and (not obj in adjacents) and (not obj in exclude):
                adjacents.append(obj)
        l=[]
        crash=False
        for obj in adjacents:
            if not obj.free_moving:
                crash=True
            else:
                l.append(self.bundle(obj,direc,exclude=exclude+[mi]))
        for p in l:
            if p[1]:
                crash=True
                break
        r = [mi]
        mass=mi.mass
        if crash:
            for p in l:
                if p[1]:
                    for obj in p[0]:
                        if not obj in r:
                            r.append(obj)
                            mass+=obj.mass
        else:
            for p in l:
                for obj in p[0]:
                    if not obj in r:
                        r.append(obj)
                        mass+=obj.mass
        return r,crash,mass
    def pushed(self,obj:Obj,direc:str,force=-1,damage=-1)->str:
        return self.move(obj,obj.movedirec(direc),1,voluntary=False,force=force)
    def auto_move(self,mi):
        if not mi in self.minions:
            return
        return self.move(mi,mi.move_ty,mi.move_distance,voluntary=True,force=mi.mass,damage=mi.mass)
    def move(self,mi:Obj,move_ty,distance,voluntary=True,force=-1,damage=-1):
        #MOVE_TYPES=["s","f","b","l","r","a","c","t"] #digest=>after every unit movement, stamina recharge=>after whole movement
        if force==-1:
            force=mi.mass
        if damage==-1:
            damage=mi.mass
        if move_ty==0:
            self.digest(mi)
            if voluntary:
                mi.stamina=min(mi.max_stamina,mi.stamina+mi.stamina_recharge)
            return ""
        if move_ty in [1,2,3,4]:
            if distance==0:
                if voluntary:
                    mi.stamina = min(mi.max_stamina, mi.stamina + mi.stamina_recharge)
                #print("success")
                return "success"
            if voluntary and mi.stamina<int(STAMINA_CONSUM_CONSTS[move_ty]):
                mi.stamina = min(mi.max_stamina, mi.stamina + mi.stamina_recharge)
                #print("lack stamina")
                return "lack stamina"
            direc=mi.direc(move_ty)
            bund,crash,mass=self.bundle(mi,direc)
            if crash:
                for victim in bund:
                    if (type(victim)==Minion) and ((victim!=mi) or (not voluntary)):
                        self.attack(victim,amount=damage)
                if voluntary:
                    mi.stamina = min(mi.max_stamina, mi.stamina + mi.stamina_recharge)
                #print("crash")
                return "crash"
            if voluntary and mass-mi.mass>force:
                mi.stamina = min(mi.max_stamina, mi.stamina + mi.stamina_recharge)
                #print("heavy")
                return "heavy"
            if (not voluntary) and mass>force:
                #print("mass",mass)
                #print("force",force)
                return "heavy"
            #position change
            to_be_empty=[]
            for obj in bund:
                for pos in obj.occupying_area():
                    if not self.occupy_map[previous(pos,direc)] in bund:
                        to_be_empty.append(pos)
            for pos in to_be_empty:
                self.occupy_map[pos]=None
            for obj in bund:
                for pos in obj.occupying_area():
                    self.occupy_map[next(pos,direc)]=obj
                obj.pos=next(obj.pos,direc)
                if type(obj)==Minion:
                    self.digest(obj)
            #stamina
            if voluntary:
                mi.stamina-=int(STAMINA_CONSUM_CONSTS[move_ty])
            return self.move(mi,move_ty,distance-1)
        #turn
        if move_ty in [5,6,7]:
            if move_ty==5:
                angle=1
            elif move_ty==6:
                angle=3
            else:
                angle=2
            consum=max(1,int(STAMINA_CONSUM_CONSTS[move_ty]*sqrt(mi.area)))
            if voluntary and mi.stamina<consum:
                self.digest(mi)
                mi.stamina = min(mi.max_stamina, mi.stamina + mi.stamina_recharge)
                return ""
            #already occupied
            for pos in mi.rel_occupying_area():
                if not self.occupy_map[mi.relpos2abspos(rotate(pos,angle))] in [None,mi]:
                    self.digest(mi)
                    if voluntary:
                        mi.stamina = min(mi.max_stamina, mi.stamina + mi.stamina_recharge)
                    return ""
            #turning
            to_be_empty=[]
            for pos in mi.rel_occupying_area():
                if self.occupy_map[mi.relpos2abspos(rotate(pos,-angle))]!=mi:
                    to_be_empty.append(mi.relpos2abspos(pos))
            for abspos in to_be_empty:
                self.occupy_map[abspos] = None
            for pos in mi.rel_occupying_area():
                self.occupy_map[mi.relpos2abspos(rotate(pos, angle))]=mi
            mi.head=num2direc(direc2num(mi.head)+angle)
            if voluntary:
                mi.stamina-=consum
                mi.stamina = min(mi.max_stamina, mi.stamina + mi.stamina_recharge)
            self.digest(mi)
    def doable(self,female,male):
        return type(female)==Minion and type(male)==Minion and (female.species_id==male.species_id)
    def intercourse(self,female,male):
        if female.pregnant:
            return
        #what happens?
        female.get_num_child()
        female.pregnant=True

    def childbirth(self,mi:Minion):
        assert mi.pregnant
        print("birth!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
        self.mk_invisible(mi)
        num_success=0
        for i in range(mi.num_child):
            child=Minion(trait=mi.get_trait())
            child.pos=mi.pos
            if not self.available(child.occupying_area()):
                if self.available(mi.occupying_area()):
                    self.mk_visible(mi)
                else:
                    self.kill(mi)
                return
            self.register_minion(child)
            mi.mass-=child.mass
            mi.num_child-=1
            num_success+=1
            movedirec=child.movedirec(random.choice(DIRECTIONS))
            self.move(mi=child,move_ty=movedirec,distance=2*child.size[1-(movedirec-1)//2],voluntary=False,force=1000*child.mass)
        if self.available(mi.occupying_area()):
            self.mk_visible(mi)
            mi.pregnant=False
        else:
            self.kill(mi)
    def auto_jaw_movement(self,mi):
        return self.jaw_movement(mi,jaw_move_ty=mi.bite)
    def jaw_movement(self,mi:Minion,jaw_move_ty)->bool: #0:no change, #1:1 change, #2:2 changes in row
        if jaw_move_ty==0:
            return True
        elif jaw_move_ty==2: #if one move is successful, finish the rest. Otherwise, stop.
            if not self.jaw_movement(mi,1):
                return False
            return self.jaw_movement(mi,1)
        elif jaw_move_ty==1 and mi.jaw_state!=(0,0):
            if int(STAMINA_JAW_CONST * sqrt(mi.area)) > mi.stamina:
                return False
            mi.stamina -= int(STAMINA_JAW_CONST * sqrt(mi.area))
            for i in range(-mi.a+1,mi.a):
                self.occupy_map[mi.relpos2abspos((i,-mi.b))]=None
            mi.jaw_state=(0,0)
            return True
        elif jaw_move_ty==1 and mi.jaw_state==(0,0):
            if int(STAMINA_JAW_CONST * sqrt(mi.area)) > mi.stamina:
                return False
            mi.stamina -= int(STAMINA_JAW_CONST * sqrt(mi.area))
            #intercourse mount
            sexy=False
            sexy_pos=None
            detect=None
            for i in range(-mi.a+1,mi.a):
                pos=mi.relpos2abspos((i,-mi.b))
                if self.occupy_map[pos]!=None:
                    if detect==None:
                        detect=self.occupy_map[pos]
                        sexy=True
                        sexy_pos=pos
                    else:
                        sexy=False
            if sexy:
                if self.doable(mi,detect) and (sexy_pos in detect.genitals_pos()):
                    self.intercourse(detect,mi)
                    return False
            #bite
            mi.free_moving=False
            turnback=False
            blocked=[False,False]
            while mi.jaw_state[0]+mi.jaw_state[1]<mi.size[0]-2:
                if blocked[0] and (not blocked[1]):
                    left=False
                elif blocked[1] and (not blocked[0]):
                    left=True
                elif blocked[0] and blocked[1]:
                    turnback=True
                    break
                else:
                    left=random.choice([True,False])
                new_pos=mi.jaw_pos(mi.jaw_state[0]+1) if left else mi.jaw_pos(mi.size[0]-mi.jaw_state[1]-2)
                obj=self.occupy_map[new_pos]
                if obj == None:
                    mi.jaw_state=(mi.jaw_state[0]+1,mi.jaw_state[1]) if left else (mi.jaw_state[0],mi.jaw_state[1]+1)
                    self.occupy_map[new_pos] = mi
                elif not obj.free_moving:
                    blocked[0 if left else 1]=True
                    continue
                else:
                    outcome = self.pushed(obj=obj, direc=num2direc(direc2num(mi.head) +(-1 if left else 1)),
                                          force=int(mi.mass * JAWPOWER_PER_MASS),damage=int(mi.mass * JAWDAMAGE_PER_MASS))
                    if outcome=="heavy":
                        blocked[0 if left else 1] = True
                        continue
                    elif outcome == "crash":
                        turnback = True
                        break
                    else:
                        mi.jaw_state = (mi.jaw_state[0] + 1, mi.jaw_state[1]) if left else (mi.jaw_state[0], mi.jaw_state[1] + 1)
                        self.occupy_map[new_pos] = mi


            if turnback:
                for i in range(1,mi.jaw_state[0]+1):
                    self.occupy_map[mi.jaw_pos(i)]=None
                for i in range(1,mi.jaw_state[1]+1):
                    self.occupy_map[mi.jaw_pos(mi.size[0]-1-i)]=None
                mi.jaw_state=(0,0)
                mi.free_moving=True
                return False
            mi.free_moving=True
            return True


    def one_step(self):
        self.moment+=1
        if self.moment%CONVERSION_PERIOD==0:
            self.mineral_conversion()
        self.schedule_update()
        for mi in self.minions:
            self.auto_jaw_movement(mi)
        for mi in self.minions:
            self.auto_move(mi)  #auto_move(digest(birth,excrete))
        for mi in self.minions:
            self.excrete(mi)
        self.increase_ages()
        self.learn_from_groupft()
        self.render()

    def render(self):
        l=[[mineral_color(self.mineral_map[(i,j)]) for j in range(self.size[1])] for i in range(self.size[0])]
        d=dict()
        for mi in self.minions:
            d[mi]=mi.render()
        for obj in self.objects:
            d[obj]=obj.render()
        for i in range(self.size[0]):
            for j in range(self.size[1]):
                if self.occupy_map[(i,j)]!=None:
                    try:
                        l[i][j]=d[self.occupy_map[(i,j)]][(i,j)]
                    except:
                        print("pos:",(i,j))
                        mi=self.occupy_map[(i,j)]
                        print("mi:",mi)
                        print("jaw",mi.jaw_state)
                        #print("the other jaw",self.minions[1-self.minions.index(mi)].jaw_state)
                        #raise Exception()

        self.snapshot=l


