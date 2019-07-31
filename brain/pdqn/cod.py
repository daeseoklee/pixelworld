from typing import List
from typing import Dict
from typing import Tuple
import os
import struct
maind="C:/Users/daese/go/src/github.com/daeseoklee/pixelworld"
braind=os.path.join(maind,"brain")
tmpd=os.path.join(braind,"tmp")
file_to_read=os.path.join(tmpd,"dear_python.txt")
file_to_write=os.path.join(tmpd,"dear_go.txt")

class Behavior:
    def __init__(self,move_ty:int,move_dist:float,jawmove_ty:int,emission:int):
        self.move_ty=move_ty
        self.move_dist=move_dist
        self.jawmove_ty=jawmove_ty
        self.emission=emission
class Input:
    def __init__(self,vision:List[List[Tuple[float,float,float,float]]],smell:List[float],hp:float,sp:float,pregnant:int):
        self.vision=vision
        self.smell=smell
        self.hp=hp
        self.sp=sp
        self.pregnant=pregnant
class Info:
    def __init__(self,trait:int,xlen:int,ylen:int,id:int,reward:float,before:Input,action:Behavior,after:Input):
        self.trait=trait
        self.xlen=xlen
        self.ylen=ylen
        self.id=id
        self.reward=reward
        self.before=before
        self.action=action
        self.after=after
def read_int(r:bytes,i:int,n:int):
    s=0
    for j in range(n):
        s+=r[i+j]*256**(n-j-1)
    return s,i+n
def read_str(r:bytes,i:int,n:int):
    s=""
    for j in range(n):
        s+=chr(r[i+j])
    return s,i+n
def read_float(r:bytes,i:int):
    [x]=struct.unpack('>d',r[i:i+8])
    return float(x),i+8     
def decode_piece(r:bytes,i:int,n:int): #bytes*int*int->Info, i:read from,n:read length
    trait,i=read_int(r,i,2)
    id,i=read_int(r,i,4)
    xlen,i=read_int(r,i,1)
    ylen,i=read_int(r,i,1)
    reward,i=read_float(r,i)
    #before
    vision_before=[[(r[i+4*(j*3*ylen+k)],r[i+4*(j*3*ylen+k)+1],r[i+4*(j*3*ylen+k)+2],r[i+4*(j*3*ylen+k)+3]) for k in range(3*ylen)] for j in range(3*xlen) ]
    i+=36*xlen*ylen
    smell_before=[0.0 for j in range(100)]
    for j in range(100):
        a,i=read_float(r,i)
        smell_before[j]=a
    #print(r[i],r[i+1],r[i+2],r[i+3],r[i+4],r[i+5],r[i+6],r[i+7])
    #print(r[i:i+8])
    hp_before,i=read_float(r,i)
    #print(r[i],r[i+1],r[i+2],r[i+3],r[i+4],r[i+5],r[i+6],r[i+7])
    #print(r[i:i+8])
    sp_before,i=read_float(r,i)
    pregnant_before,i=read_int(r,i,1)
    #action
    move_ty,i=read_int(r,i,1)
    move_dist,i=read_float(r,i)
    #move_dist,i=read_int(r,i,2)
    jawmove_ty,i=read_int(r,i,1)
    emission,i=read_int(r,i,1)
    #after
    vision_after=[[(r[i+4*(j*3*ylen+k)],r[i+4*(j*3*ylen+k)+1],r[i+4*(j*3*ylen+k)+2],r[i+4*(j*3*ylen+k)+3]) for k in range(3*ylen)] for j in range(3*xlen) ]
    i+=36*xlen*ylen
    smell_after=[0.0 for j in range(100)]
    for j in range(100):
        a,i=read_float(r,i)
        smell_after[j]=a
    hp_after,i=read_float(r,i)
    sp_after,i=read_float(r,i)
    pregnant_after,i=read_int(r,i,1)
    return Info(trait,xlen,ylen,id,reward,Input(vision_before,smell_before,hp_before,sp_before,pregnant_before),
    Behavior(move_ty,move_dist,jawmove_ty,emission),Input(vision_after,smell_after,hp_after,sp_after,pregnant_after))


def read_token(r:bytes,i:int):
    if r[i]==1:
        return True, r[i+1]*256**2+r[i+2]*256+r[i+3],i+4
    elif r[i]==0:
        return False,0,i+1
    raise Exception("something is wrong")

def append_inf(r:bytes,i:int,n:int,l:List[Info]):
    l.append(decode_piece(r,i,n))
    return i+n

def decode_whole(r:bytes): #bytes->Info[]
    print("python:total length:",len(r))
    infs=[]
    i=0
    #--------------------------
    p,n,i=read_token(r,i)
    title,i=read_str(r,i,n)
    #-------------------------
    p,n,i=read_token(r,i)
    kingdomname,i=read_str(r,i,n)
    #-------------------------
    p,n,i=read_token(r,i)
    mapname,i=read_str(r,i,n)
    #--------------------------
    p,n,i=read_token(r,i)
    assert n==4
    epoch,i=read_int(r,i,n) #epoch>=1 
    #-------------------------
    p,n,i=read_token(r,i)
    assert n==4
    weightfrom,i=read_int(r,i,n) #weightfrom==0:renew
    #-------------------------
    p,n,i=read_token(r,i)
    assert n==4
    saveat,i=read_int(r,i,n)
    #-------------------------
    p,n,i=read_token(r,i)
    assert n<20
    dolearn=[False for j in range(n)]
    for j in range(n):
        x,i=read_int(r,i,1)
        if x==1:
            dolearn[j]=True
        elif x==0:
            dolearn[j]=False
        else:
            raise Exception("wrong encoding")
    #-------------------------
    p,n,i=read_token(r,i)
    assert n==5
    moment,i=read_int(r,i,n)
    #-------------------------
    p,n,i=read_token(r,i)
    assert n==1
    save,i=read_int(r,i,n)
    #-------------------------
    p,n,i=read_token(r,i)
    assert n==1
    train,i=read_int(r,i,n)
    #-------------------------
    while i<len(r):
        p,n,i=read_token(r,i)
        if p:
            i=append_inf(r,i,n,infs)
        else:
            break
    return title,kingdomname,mapname,epoch,weightfrom,saveat,dolearn,moment,save,train,infs

def decode_states():
    f=open(file_to_read,mode='rb')
    r=f.read()
    return decode_whole(r)

def write_int(l:List[int],i:int,d:int,n:int):
    assert n<256**d
    q=n
    for j in range(d):
        q,r=q//256,q%256
        l[i+d-j-1]=r
    return i+d

def write_float(l:List[int],i:int,a:float):
    b=struct.pack(">d",a)
    for j in range(8):
        l[i+j]=b[j]
    return i+8

def write_byte(l:List[int],i:int,b:int):
    l[i]=b
    return i+1

def join(segs:List[List[int]]):
    l:List[int]
    n=1
    for seg in segs:
        n+=4+len(seg)
    l=[0 for i in range(n)] 
    i=0
    for seg in segs:
        l[i]=1
        i+=1
        i=write_int(l,i,3,len(seg))
        for j in range(len(seg)):
            i=write_byte(l,i,seg[j])
    assert i==n-1
    l[i]=0
    return l 
    
            
def bytes_from_atom(id:int,action:Behavior,xlen:int,ylen:int):
    l=[0 for i in range(4+1+1+1+8+1+1)]
    i=0
    i=write_int(l,i,4,id)
    i=write_byte(l,i,xlen)
    i=write_byte(l,i,ylen)
    i=write_byte(l,i,action.move_ty)
    i=write_float(l,i,action.move_dist)
    i=write_byte(l,i,action.jawmove_ty)
    i=write_byte(l,i,action.emission)
    return l

def bytes_from_actions(actions:Dict[int,Tuple[Behavior,int,int]]):
    action:Behavior
    return bytes(join([bytes_from_atom(id,actions[id][0],actions[id][1],actions[id][2]) for id in actions.keys()]))

def write_actions(actions:Dict[int,Tuple[Behavior,int,int]]):
    f=open(file_to_write,mode="wb")
    f.write(bytes_from_actions(actions)) 
    
    

def main():
    l=decode_states()
    print(len(l))

    for i in range(50):
        print("---------------------------------------")
        inf=l[i]
        print("xlen:",inf.xlen)
        print("ylen:",inf.ylen)
        print("trait:",inf.trait)
        print("id:",inf.id)
        print("pleasure:",inf.pleasure)
        print("hp_before",inf.before.hp)
        print("sp_before",inf.before.sp)
        print("hp_after:",inf.after.hp)
        print("sp_After:",inf.after.sp)
        print("mofety:",inf.action.move_ty)
        print("moveDist:",inf.action.move_dist)
        print("emission:",inf.action.emission)

    f.close()

#main()