import argparse
parser=argparse.ArgumentParser()
parser.add_argument('port' ,type=int)
parser.add_argument('second',type=str)
args=parser.parse_args()
print(args.port,args.second)