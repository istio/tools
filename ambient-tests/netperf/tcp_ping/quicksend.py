import socket
import sys

HOST = sys.argv[1]
PORT = int(sys.argv[2])

def ping():
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.connect((HOST, PORT))
    sock.sendall(b"Hello, world")
    buf = sock.recv(1024)
    print(buf)
    sock.close()

ping()
ping()
ping()
ping()
ping()
ping()
