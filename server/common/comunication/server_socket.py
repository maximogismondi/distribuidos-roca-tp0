import logging
import socket
from typing import Tuple

from common.comunication.client_socket import ClientSocket

BLOCKING_TIMEOUT = 1.0


class ServerSocket:
    _socket: socket.socket
    address: Tuple[str, int]

    def __init__(self, address: Tuple[str, int], listen_backlog: int) -> None:
        self._socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._socket.bind(address)
        self._socket.listen(listen_backlog)
        self._socket.settimeout(BLOCKING_TIMEOUT)
        self.address = address

    def accept(self) -> ClientSocket:
        c, addr = self._socket.accept()

        logging.info(f"action: accept_connections | result: success | ip: {addr[0]}")

        return ClientSocket(c, addr)

    def close(self) -> None:
        self.socket.close()
