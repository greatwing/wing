
import os,sys

def Main():
    os.system("go build -ldflags \"-w -s\" -o ../bin/client.exe ../client")
    os.system("go build -ldflags \"-w -s\" -o ../bin/gateway.exe ../server/gateway")
    os.system("go build -ldflags \"-w -s\" -o ../bin/game.exe ../server/game")
    os.system("go build -ldflags \"-w -s\" -o ../bin/login.exe ../server/login")

Main()