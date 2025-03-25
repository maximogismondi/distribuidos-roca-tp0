COMMUNICATION_DELIMITER = "\n"


class ClientSocket:
    def __init__(self, socket):
        self._socket = socket
        self._overflown = ""

    def send_message(self, msg):
        msg += COMMUNICATION_DELIMITER

        msg_bytes = msg.encode("utf-8")
        bytes_sent = 0

        while bytes_sent < len(msg_bytes):
            bytes_sent += self._socket.send(msg_bytes[bytes_sent:])

            if bytes_sent == 0:
                raise ConnectionError("Socket connection broken")

    def receive_message(self):
        while COMMUNICATION_DELIMITER not in self._overflow:
            chunk = self._socket.recv(1024)
            if not chunk:
                # No more data, return whatever is left (could be empty)
                if self._overflow:
                    msg = self._overflow
                    self._overflow = ""
                    return msg
                else:
                    return ""
            self._overflow += chunk.decode("utf-8")

        # Extract the first full message and keep the rest in overflow
        message, self._overflow = self._overflow.split(COMMUNICATION_DELIMITER, 1)
        return message

    def address(self):
        return self._socket.getpeername()

    def close(self):
        self._socket.close()
