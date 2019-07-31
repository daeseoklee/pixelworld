from typing import List
import torch
import torch.nn as nn
import torch.nn.functional as f
from reward import reward
import cod
import os
import random
maind="C:/Users/daese/go/src/github.com/daeseoklee/pixelworld"
weightsd=os.path.join(maind,"brain/pdqn/weights")
epsilon=0.2


class Visual(nn.Module):
    def __init__(self,xlen,ylen):
        super(Visual,self).__init__()
        self.xlen=xlen
        self.ylen=ylen
        self.alen=(xlen-1)//2
        self.blen=(ylen-1)//2
        self.xvision=3*xlen
        self.yvision=3*ylen
        """
        self.convnet=nn.Sequential(
            nn.Conv2d(in_channels=4,out_channels=8,kernel_size=6,stride=3),
            nn.ReLU(),
            nn.Conv2d(in_channels=8,out_channels=16,kernel_size=2),
            nn.ReLU()
        )
        self.final=nn.Sequential(
            nn.Linear(16*(2*self.alen-1)*(2*self.blen-1),64),
            nn.ReLU()
        )
        """
        self.simple=nn.Sequential(
            nn.Linear(36*self.xlen*self.ylen,64),
            nn.ReLU()
        )
    def forward(self,vision):
        assert (3*self.xlen==len(vision))&(3*self.ylen==len(vision[0]))
        m=[[vision[i][j][0]/255 for j in range(3*self.ylen)] for i in range(3*self.xlen)]
        r=[[vision[i][j][1]/255 for j in range(3*self.ylen)] for i in range(3*self.xlen)]
        g=[[vision[i][j][2]/255 for j in range(3*self.ylen)] for i in range(3*self.xlen)]
        b=[[vision[i][j][3]/255 for j in range(3*self.ylen)] for i in range(3*self.xlen)]
        v=torch.tensor([[m,r,g,b]])
        #return f.normalize(self.final(self.convnet(v).reshape(1,-1)),dim=1)
        return f.normalize(self.simple(v.reshape(1,-1)))


class Olfactory(nn.Module):
    def __init__(self):
        super(Olfactory,self).__init__()
        """
        self.net=nn.Sequential(
            nn.Linear(100,64),
            nn.ReLU(),
            nn.Linear(64,64),
            nn.ReLU(),
            nn.Linear(64,32),
            nn.ReLU()
        )
        """
        self.net=nn.Sequential(
            nn.Linear(100,32),
            nn.ReLU()
        )
    def forward(self,smell):
        v=torch.tensor(smell).reshape(1,-1)
        return f.normalize(self.net(v),dim=1)

class Premotor(nn.Module):
    def __init__(self,visual,olfactory):
        super(Premotor,self).__init__()
        self.visual=visual #Visual()
        self.olfactory=olfactory #Olfactory()
        self.net=nn.Sequential(
            nn.Linear(99,256),
            nn.ReLU()
        )
    def forward(self,inp:cod.Input):
        v=self.visual(inp.vision)
        o=self.olfactory(inp.smell)
        h=torch.tensor([[inp.hp]])
        s=torch.tensor([[inp.sp]])
        p=torch.tensor([[float(inp.pregnant)]])
        return self.net(torch.cat([v,o,h,s,p],-1))
        


class M_value_net(nn.Module): #translation, 0: forward, 1: backward, 2: left, 3: right
    def __init__(self,premotor):
        super(M_value_net,self).__init__()
        self.premotor=premotor #Premotor()
        self.net=nn.Sequential(
            nn.Linear(264,8),
            nn.Tanh()
        )
    def forward(self,inp:cod.Input,amounts):
        return self.net(torch.cat([self.premotor(inp),amounts],-1))
        

class M_argmax_net(nn.Module):
    def __init__(self,premotor):
        super(M_argmax_net,self).__init__()
        self.premotor=premotor
        self.net=nn.Sequential(
            nn.Linear(256,8),
            nn.Sigmoid()
        )
    def forward(self,inp:cod.Input):
        return self.net(self.premotor(inp))


class J_value_net(nn.Module): #jaw movement, 0:nothing, 1:once, 2:twice
    def __init__(self,premotor):
        super(J_value_net,self).__init__()
        self.premotor=premotor
        self.net=nn.Sequential(
            nn.Linear(256,3),
            nn.Tanh()
        )
    def forward(self,inp:cod.Input):
        return self.net(self.premotor(inp))

class E_value_net(nn.Module): #emission, 30 different odors
    def __init__(self,premotor):
        super(E_value_net,self).__init__()
        self.premotor=premotor
        self.net=nn.Sequential(
            nn.Linear(256,30),
            nn.Tanh()
        )
    def forward(self,inp:cod.Input):
        return self.net(self.premotor(inp))

class PDQN(nn.Module):
    def __init__(self,xlen:int,ylen:int):
        super(PDQN,self).__init__()
        self.decay_factor=0.7
        self.visual=Visual(xlen,ylen)
        self.olfactory=Olfactory()
        self.premotor=Premotor(self.visual,self.olfactory)
        self.m_argmax_net=M_argmax_net(self.premotor)
        self.m_value_net=M_value_net(self.premotor)
        self.j_value_net=J_value_net(self.premotor)
        self.e_value_net=E_value_net(self.premotor)
        self.theta_opt=torch.optim.Adadelta(params=self.m_argmax_net.parameters(),lr=1.0,rho=0.9)
        #self.theta_opt=torch.optim.SGD(params=self.m_argmax_net.parameters(),lr=1)
        self.omega_opt=torch.optim.Adadelta(params=list(self.m_value_net.parameters())+list(self.j_value_net.parameters())+list(self.e_value_net.parameters()),lr=1.0,rho=0.9)
        #self.omega_opt=torch.optim.SGD(params=list(self.m_value_net.parameters())+list(self.j_value_net.parameters())+list(self.e_value_net.parameters()),lr=1)
        nn.utils.clip_grad_norm(self.parameters(),max_norm=5.0)

    def theta_loss(self,inf): #for the argmax approximation to work properly #검토필요!!!!!!
        inp=inf.before
        sum=torch.tensor(0.0)
        sum-=self.m_value_net(inp,self.m_argmax_net(inp)).sum()
        sum-=self.j_value_net(inp).sum()
        sum-=self.e_value_net(inp).sum()
        return sum
    def omega_loss(self,inf:cod.Info): #for fitting Bellman's equation #검토필요!!!!!!

        m_m=self.m_value_net(inf.after,self.m_argmax_net(inf.after)).detach().max()
        m_j=self.j_value_net(inf.after).detach().max()
        m_e=self.e_value_net(inf.after).detach().max()
        m=m_m+m_j+m_e

        q_m=self.m_value_net(inf.before,self.m_argmax_net(inf.before))[0][inf.action.move_ty]
        q_j=self.j_value_net(inf.before)[0][inf.action.jawmove_ty]
        q_e=self.e_value_net(inf.before)[0][inf.action.emission]
        q=q_m+q_j+q_e
        return (q-inf.reward-self.decay_factor*m)**2
    def learning_step(self,inf):
        self.train(True)
        self.theta_opt.zero_grad()
        self.theta_loss(inf).backward()
        self.theta_opt.step()
        self.omega_opt.zero_grad()
        self.omega_loss(inf).backward()
        self.omega_opt.step()
    def forward(self,inp:cod.Input,train): #검토필요!!!!!!
        self.train(False)
        if train and random.uniform(0,1)<epsilon:
            move_ty=random.randint(0,7)
            if move_ty in [1,2,3,4]:
                move_dist=random.uniform(0,1)
            else:
                move_dist=0.0
            jawmove_ty=random.randint(0,2)
            emission=random.randint(0,29)
        else:
            argm_m=self.m_argmax_net(inp)
            m_m=self.m_value_net(inp,argm_m)[0]
            m_j=self.j_value_net(inp)[0]
            m_e=self.e_value_net(inp)[0]
            move_ty=int(torch.argmax(m_m))
            move_dist=float(argm_m[0][move_ty])
            jawmove_ty=int(torch.argmax(m_j))
            emission=int(torch.argmax(m_e))
        return cod.Behavior(move_ty,move_dist,jawmove_ty,emission)

class Nets():
    def __init__(self,m:int,nets:List[PDQN],title:str,kingdomname:str,mapname:str,epoch:int,weightfrom:int,saveat:int,dolearn:List[bool]):
        self.m=m
        self.nets=nets
        self.title=title      
        self.kingdomname=kingdomname
        self.mapname=mapname
        self.epoch=epoch
        self.weightfrom=weightfrom
        self.saveat=saveat
        self.dolearn=dolearn
    def save(self):
        global weightsd
        for i in range(self.m):
            if self.nets[i]!=None and self.dolearn[i]:
                torch.save(self.nets[i].theta_opt.state_dict(),os.path.join(weightsd,self.title,self.kingdomname,self.mapname,str(i)+"_theta_"+str(self.saveat)+".weight"))
                torch.save(self.nets[i].omega_opt.state_dict(),os.path.join(weightsd,self.title,self.kingdomname,self.mapname,str(i)+"_omega_"+str(self.saveat)+".weight"))
    def load(self):
        global weightsd
        for i in range(self.m):
            if self.nets[i]!=None:
                self.nets[i].theta_opt.load_state_dict(torch.load(os.path.join(weightsd,self.title,self.kingdomname,self.mapname,str(i)+"_theta_"+str(self.weightfrom)+".weight")))
                self.nets[i].omega_opt.load_state_dict(torch.load(os.path.join(weightsd,self.title,self.kingdomname,self.mapname,str(i)+"_omega_"+str(self.weightfrom)+".weight")))
    def compute(self,collection,train):
        inf:cod.Info
        #learning
        if train:
            for i in range(self.m):
                if self.nets[i]!=None and self.dolearn[i] and len(collection[i])>0:
                    infs=collection[i]
                    random.shuffle(infs)
                    m=len(infs)
                    m=min(10,m)
                    for inf in infs[:m]:
                        self.nets[i].learning_step(inf)
        #determining actions
        actions=dict()
        for i in range(self.m):
            if self.nets[i]!=None:
                for inf in collection[i]:
                    actions[inf.id]=(self.nets[i](inf.after,train),inf.xlen,inf.ylen)
        return actions
    
`q  `