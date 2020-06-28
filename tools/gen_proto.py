
import os,sys

PROTO_DIR = "../proto"

def Main():
    global PROTO_DIR
    if len(sys.argv) > 1:
        PROTO_DIR = sys.argv[1]
    # print(PROTO_DIR)

    for(root,_,files) in os.walk(PROTO_DIR):
        for filename in files:
            fileext = os.path.splitext(filename)[1]
            if fileext == ".proto":
                cmd = "protoc --gogofaster_out={} --proto_path={} {}".format(root, root, filename)
                print(cmd)
                os.system(cmd)

Main()