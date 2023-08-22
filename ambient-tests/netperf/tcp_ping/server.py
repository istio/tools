import socket

sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM) 
sock.bind(("", 9999))
sock.listen()
while True:
    con, addr = sock.accept()
    buf = con.recv(1000)
    if buf:
        con.send(b"REPLY\n")
    else:
        con.close()
