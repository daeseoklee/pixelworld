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
epsilon=0.1


class Visual(nn.Module):
    def __init__(self,xlen,ylen):
        super(Visual,self).__init__()
        self.xlen=xlen
        self.ylen=ylen
        self.alen=(xlen-1)//2
        self.blen=(ylen-1)//2
        self.xvision=3*xlen
        self.yvision=3*ylen
        self.convnet=nn.Sequential(
            nn.Conv2d(in_channels=4,out_channels=8,kernel_size=6,stride=3),
            nn.ReLU6(),
            nn.Conv2d(in_channels=8,out_channels=16,kernel_size=2),
            nn.ReLU6()
        )
        self.final=nn.Sequential(
            nn.Linear(16*(2*self.alen-1)*(2*self.blen-1),64),
            nn.ReLU6()
        )
    def forward(self,visions):
        assert (3*self.xlen==len(visions[0]))&(3*self.ylen==len(visions[0][0]))
        l=[]
        for vision in visions:
            m=[[vision[i][j][0]/255 for j in range(3*self.ylen)] for i in range(3*self.xlen)]
            r=[[vision[i][j][1]/255 for j in range(3*self.ylen)] for i in range(3*self.xlen)]
            g=[[vision[i][j][2]/255 for j in range(3*self.ylen)] for i in range(3*self.xlen)]
            b=[[vision[i][j][3]/255 for j in range(3*self.ylen)] for i in range(3*self.xlen)]
            l.append([m,r,g,b])
        v=torch.tensor(l)
        return f.normalize(self.final(self.convnet(v).reshape(len(visions),-1)),dim=1)


class Olfactory(nn.Module):
    def __init__(self):
        super(Olfactory,self).__init__()
        self.net=nn.Sequential(
            nn.Linear(100,64),
            nn.ReLU6(),
            nn.Linear(64,64),
            nn.ReLU6(),
            nn.Linear(64,32),
            nn.ReLU6()
        )
    def forward(self,smells):
        v=torch.tensor(smells)
        return f.normalize(self.net(v),dim=1)

class Premotor(nn.Module):
    def __init__(self,visual,olfactory):
        super(Premotor,self).__init__()
        self.visual=visual #Visual()
        self.olfactory=olfactory #Olfactory()
        self.net=nn.Sequential(
            nn.Linear(99,256),
            nn.ReLU6()
        )
    def forward(self,inps:List[cod.Input]):
        inp:cod.Input
        visions=[]
        smells=[]
        hps=[]
        sps=[]
        pregs=[]
        for inp in inps:
            visions.append(inp.vision)
            smells.append(inp.smell)
            hps.append([inp.hp])
            sps.append([inp.sp])
            pregs.append([float(inp.pregnant)])
        v=self.visual(visions)
        o=self.olfactory(smells)
        h=torch.tensor(hps)
        s=torch.tensor(sps)
        p=torch.tensor(pregs)
        return self.net(torch.cat([v,o,h,s,p],-1))
        


class T_value_net(nn.Module): #translation, 0: forward, 1: backward, 2: left, 3: right
    def __init__(self,premotor):
        super(T_value_net,self).__init__()
        self.premotor=premotor #Premotor()
        self.net=nn.Sequential(
            nn.Linear(260,4),
            nn.ReLU()
        )
    def forward(self,inps:List[cod.Input],amounts):
        return self.net(torch.cat([self.premotor(inps),amounts],-1))
        

class T_argmax_net(nn.Module):
    def __init__(self,premotor):
        super(T_argmax_net,self).__init__()
        self.premotor=premotor
        self.net=nn.Sequential(
            nn.Linear(256,4),
            nn.ReLU()
        )
    def forward(self,inp:cod.Input):
        return self.net(self.premotor(inp))

class R_value_net(nn.Module): #rotation, 0:stop, 1:left 90 degree, 2: right 90 degree, 3: 180
    def __init__(self,premotor):
        super(R_value_net,self).__init__()
        self.premotor=premotor
        self.net=nn.Sequential(
            nn.Linear(256,4),
            nn.ReLU()
        )
    def forward(self,inps:List[cod.Input]):
        return self.net(self.premotor(inps))

class J_value_net(nn.Module): #jaw movement, 0:nothing, 1:once, 2:twice
    def __init__(self,premotor):
        super(J_value_net,self).__init__()
        self.premotor=premotor
        self.net=nn.Sequential(
            nn.Linear(256,3),
            nn.ReLU()
        )
    def forward(self,inps:List[cod.Input]):
        return self.net(self.premotor(inps))

class E_value_net(nn.Module): #emission, 30 different odors
    def __init__(self,premotor):
        super(E_value_net,self).__init__()
        self.premotor=premotor
        self.net=nn.Sequential(
            nn.Linear(256,30),
            nn.ReLU()
        )
    def forward(self,inps:List[cod.Input]):
        return self.net(self.premotor(inps))

class PDQN(nn.Module):
    def __init__(self,xlen:int,ylen:int):
        super(PDQN,self).__init__()
        self.decay_factor=0.95
        self.visual=Visual(xlen,ylen)
        self.olfactory=Olfactory()
        self.premotor=Premotor(self.visual,self.olfactory)
        self.t_argmax_net=T_argmax_net(self.premotor)
        self.t_value_net=T_value_net(self.premotor)
        self.r_value_net=R_value_net(self.premotor)
        self.j_value_net=J_value_net(self.premotor)
        self.e_value_net=E_value_net(self.premotor)
        self.theta_opt=torch.optim.SGD(params=self.t_argmax_net.parameters(),lr=0.1)
        self.omega_opt=torch.optim.SGD(params=list(self.t_value_net.parameters())+list(self.r_value_net.parameters()),lr=0.1)

    def theta_loss(self,infs): #for the argmax approximation to work properly #검토필요!!!!!!
        inps=[]
        for inf in infs:
            inps.append(inf.before)
        sum=torch.tensor(0.0)
        sum-=self.t_value_net(inps,self.t_argmax_net(inps)).sum()
        sum-=self.r_value_net(inps).sum()
        sum-=self.j_value_net(inps).sum()
        sum-=self.e_value_net(inp).sum()
        return sum
    def omega_loss(self,infs:list(cod.Info)): #for fitting Bellman's equation #검토필요!!!!!!
        
        rewards=[]
        afters=[]
        befores=[]
        jawmoves=[]
        emissions=[]
        for inf in infs:
            afters.append(inf.after)
            befores.sppend(inf.before)
            jawmoves.append(inf.action.jawmove_ty)
            emissions.append(inf.action.emission)
            rewards.append(reward(inf))
        tr_after=torch.cat([self.t_value_net(afters,self.t_argmax_net(afters)).detach(),self.r_value_net(afters).detach()],-1).detach()
        j_after=self.j_value_net(afters).detach()
        e_after=self.e_value_net(afters).detach()
        m=tr_after.max(1)[0]+j_after.max(1)[0]+e_after.max(1)[0]#max_{a} Q*(s',a)

        q_tr=torch.zeros(len(infs))        
        amounts=torch.zeros(len(infs),4)
        for i in range(len(infs)):
            j=infs[i].action.move_ty-1
            if j in range(4):
                amounts[i][j]+=infs[i].action.move_dist
        t=self.t_value_net(befores,amounts)
        r=self.r_value_net(befores)
        for i in range(len(infs)):
            j=infs[i].action.move_ty
            if j in [1,2,3,4]:
                q_tr[i]+=t[i][j-1]
            else:
                q_tr[i]+=t[i][max(0,j-4)]

        q_j=self.j_value_net(befores).gather(1,torch.tensor(jawmoves).view(-1,1)).view(-1)
        q_e=self.e_value_net(befores).gather(1,torch.tensor(emissions).view(-1,1)).view(-1)
        q=q_tr+q_j+q_e

        rew=torch.tensor(rewards)
        return ((q-rew-self.decay_factor*m)**2).sum()
    def learning_step(self,infs):
        self.train(True)
        self.theta_opt.zero_grad()
        self.theta_loss(infs).backward()
        self.theta_opt.step()
        self.omega_opt.zero_grad()
        self.omega_loss(infs).backward()
        self.omega_opt.step()
    def forward(self,inps:List[cod.Input],train): #검토필요!!!!!!
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
            argm_t=self.t_argmax_net(inp)
            m_t=self.t_value_net(inp,argm_t)[0]
            m_r=self.r_value_net(inp)[0]
            i=int(torch.argmax(m_t))
            j=int(torch.argmax(m_r))
            if m_t[i]>m_r[j]:
                move_ty=i+1
                move_dist=float(argm_t[0][i])
            else:
                if j==0:
                    move_ty=0
                else:
                    move_ty=j+4
                move_dist=0.0
            m_j=self.j_value_net(inp)[0]
            jawmove_ty=int(torch.argmax(m_j))
            m_e=self.e_value_net(inp)[0]
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