import socket
import logging
import sys

from common.utils import Bet, store_bets, from_string
from common.client_socket import ClientSocket

BATCH_SEPARATOR = "*"
FINISH_MSG = "finish"


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(("", port))
        self._server_socket.listen(listen_backlog)
        self._server_socket.settimeout(1.0)
        self._running = True
        self._first_accept_try = True

    def run(self):
        """
        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again.
        """

        while self._running:
            try:
                client_sock = self.__accept_new_connection()

                if self._running and client_sock:
                    self.__handle_client_connection(client_sock)
                elif client_sock:
                    client_sock.close()
                    break
            except socket.timeout:
                continue
            except OSError as e:
                if self._running:
                    logging.error(
                        f"action: accept_connections | result: fail | error: {e}"
                    )
                break

        self.__cleanup()

    def __batch_to_bets(self, batch):
        betsStr = batch.split(BATCH_SEPARATOR)
        bets = [None] * len(betsStr)

        for i, betStr in enumerate(betsStr):
            try:
                bets[i] = from_string(betStr)
            except ValueError as _e:
                logging.error(
                    f"action: apuesta_almacenada | result: fail | cantidad: {len(bets)}"
                )
                return []
            except Exception as _e:
                logging.error(
                    f"action: apuesta_almacenada | result: fail | cantidad: {len(bets)}"
                )
                return []

        return bets

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """

        # Add this flag to avoid logging every accept try
        self._first_accept_try = True

        try:
            while True:
                msg = client_sock.receive_message()

                if msg == FINISH_MSG:
                    client_sock.send_message("success")
                    return

                bets = self.__batch_to_bets(msg)

                if len(bets) == 0:
                    client_sock.send_message("failure")
                    return

                store_bets(bets)

                logging.info(
                    f"action: apuesta_almacenada | result: success | cantidad: {len(bets)}"
                )

                client_sock.send_message("success")
        except ConnectionError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Add this flag to avoid logging every accept try
        if self._first_accept_try:
            logging.info("action: accept_connections | result: in_progress")
            self._first_accept_try = False

        try:
            c, addr = self._server_socket.accept()
            logging.info(
                f"action: accept_connections | result: success | ip: {addr[0]}"
            )
            return ClientSocket(c)
        except socket.timeout:
            return None

    def stop(self, _signum, _frame):
        """
        Stop the server

        Change the running flag to False so the main loop can exit
        """
        self._running = False

    def __cleanup(self):
        """
        Cleanup server resources

        Close server socket
        """
        try:
            self._server_socket.close()
            logging.info("action: exit | result: success")
        except Exception as e:
            logging.error(f"action: exit | result: fail | error: {e}")
        sys.exit(0)
