class Socket:
    def __init__(self, sock):
        self._sock = sock

    def sendall(self, data):
        self._sock.sendall(data)

    def recvall(self, length):
        data = b''
        while len(data) < length:
            packet = self._sock.recv(length - len(data))
            if not packet:
                break
            data += packet
            
        return data

    def close(self):
        self._sock.close()