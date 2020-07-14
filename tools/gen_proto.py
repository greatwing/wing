
import os,sys

PROTO_DIR = "../proto"

def Main():
    global PROTO_DIR
    if len(sys.argv) > 1:
        PROTO_DIR = sys.argv[1]
    # print(PROTO_DIR)

    dirs = os.listdir(PROTO_DIR)
    # print(dirs)

    thirdparty = os.path.join(PROTO_DIR, "thirdparty")

    for file in dirs:
        ext = os.path.splitext(file)[1]
        if ext  == ".proto":
            cmd = "protoc --gogofaster_out={} --proto_path={} --proto_path={} {}".format(PROTO_DIR, PROTO_DIR, thirdparty, file)
            print(cmd)
            os.system(cmd)

Main()