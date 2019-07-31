from multiprocessing import Pool
import time

class C():
    def __init__(self):
        self.num=0

    def increase(self,added):
        self.num += added
        print(added)
    def do(self,num,num_workers):
        a=time.time()
        if __name__=="__main__":
            with Pool(num_workers) as p:
                p.map(self.increase,list(range(num)))
        b=time.time()
        return b-a
a=C()
times=[]
times.append(a.do(1000000,1))
times.append(a.do(1000000,5))
times.append(a.do(1000000,10))
print(times)