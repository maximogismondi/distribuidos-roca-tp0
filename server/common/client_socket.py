COMMUNICATION_DELIMITER = "\n"


class ClientSocket:
    def __init__(self, socket):
        self._socket = socket

    def send_message(self, msg):
        msg += COMMUNICATION_DELIMITER

        msg_bytes = msg.encode("utf-8")
        bytes_sent = 0

        while bytes_sent < len(msg_bytes):
            bytes_sent += self._socket.send(msg_bytes[bytes_sent:])

            if bytes_sent == 0:
                raise ConnectionError("Socket connection broken")

    def receive_message(self):
        msg = b""
        while True:
            chunk = self._socket.recv(1024)
            if not chunk:
                break
            msg += chunk

            if chunk.endswith(COMMUNICATION_DELIMITER.encode("utf-8")):
                break

        return msg.decode("utf-8").rstrip(COMMUNICATION_DELIMITER)

    def address(self):
        return self._socket.getpeername()

    def close(self):
        self._socket.close()
