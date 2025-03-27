import socket

CHUNK_SIZE = 1024
SOCKET_TIMEOUT = 1.0

COMMUNICATION_DELIMITER = "\n"


class ClientSocket:

    _socket: socket.socket
    _listen_overflow: str
    address: tuple[str, int]

    def __init__(self, socket: socket.socket, address: tuple[str, int]) -> None:
        socket.settimeout(SOCKET_TIMEOUT)

        self._socket = socket
        self._listen_overflow = ""
        self.address = address

    def send_message(self, msg) -> None:
        msg += COMMUNICATION_DELIMITER

        msg_bytes: bytes = msg.encode("utf-8")
        bytes_sent: int = 0

        while bytes_sent < len(msg_bytes):
            bytes_sent += self._socket.send(msg_bytes[bytes_sent:])

            if bytes_sent == 0:
                raise ConnectionError("Socket connection broken")

    def receive_message(self) -> str:
        while COMMUNICATION_DELIMITER not in self._listen_overflow:
            chunk: bytes = self._socket.recv(CHUNK_SIZE)
            if not chunk:
                raise BrokenPipeError("Socket connection broken")

            self._listen_overflow += chunk.decode("utf-8")

        message, self._listen_overflow = self._listen_overflow.split(
            COMMUNICATION_DELIMITER, 1
        )

        return message

    def close(self):
        self._socket.close()
