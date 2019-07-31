from socket import *
import cod
import argparse
import pdqn
parser=argparse.ArgumentParser()
parser.add_argument('port',type=int)
args=parser.parse_args()

def to_int(bs):
    sum=0
    for i in range(len(bs)):
        sum+=bs[-1-i]*256**i
    return sum
def read(cl):
    x=cl.recv(4)
    return x


def main():
    server = socket()
    server.settimeout(5)
    server.bind(("127.0.0.1", args.port))
    server.listen(1)
    client, ip = server.accept()
    while True:
        recv=read(client)
        if recv==b'sent':
            title,kingdomname,mapname,epoch,weightfrom,saveat,dolearn,moment,save,train,infs=cod.decode_states()
            if moment==0:
                m=-1
                for inf in infs:
                    if inf.trait+1>m:
                        m=inf.trait+1
            else:
                m=nets.m
            collection=[[] for i in range(m)]
            for inf in infs:
                collection[inf.trait].append(inf)
            if moment==0:
                sizes=[None for i in range(m)]
                for inf in infs:
                    if sizes[inf.trait]==None:
                        sizes[inf.trait]=(inf.xlen,inf.ylen)
                net_list=[(pdqn.PDQN(sizes[i][0],sizes[i][1]) if sizes[i]!=None else None) for i in range(m)]
                nets=pdqn.Nets(m,net_list,title,kingdomname,mapname,epoch,weightfrom,saveat,dolearn)
                if weightfrom>0:
                    nets.load()
            actions=nets.compute(collection,train)
            cod.write_actions(actions)
            if save:
                nets.save()
            client.send(b'sent')
        else:
            raise Exception("received: "+str(recv))
main()