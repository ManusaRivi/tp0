import socket

class Socket:
    def __init__(self, sock: socket.socket):
        self.sock = sock

    def send(self, data):
        sz = self.sock.send(data)

        return 0 if sz <= 0 else sz
    
    def send_all(self, data: bytes):
        sent = 0
        while sent < len(data):
            sz = self.sock.send(data[sent:])
            if sz == 0:
                raise ConnectionError('Connection closed while sending data')
            sent += sz

    def recv_all(self, size: int) -> bytes:
        data = bytearray()
        while len(data) < size:
            chunk = self.sock.recv(size - len(data))
            if not chunk:
                raise ConnectionError(
                    f'Connection closed while reading: expected {size} bytes, got {len(data)}'
                )
            data.extend(chunk)
        return bytes(data)

    def recv(self, buffer_size=1024):
        return self.sock.recv(buffer_size)

    def close(self):
        self.sock.close()