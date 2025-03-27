import socket
import logging
import threading

from common.utils import store_bets, load_bets, has_won

from common.comunication.server_socket import ServerSocket
from common.comunication.client_socket import ClientSocket

from common.comunication.server_message import (
    ServerHeader,
    encode_message,
    encode_winners_message,
)
from common.comunication.agency_message import (
    BATCH_SEPARATOR,
    AgencyHeader,
    decode_identification_message,
    decode_message,
)


class Server:
    _server_socket: ServerSocket
    _running: bool
    _first_accept_try: bool
    _number_agencies: int
    _connected_clients: set[tuple[threading.Thread, ClientSocket]]
    _agencies_ready: set[int]
    _winners_by_agency: dict[int, list[str]]
    _lock: threading.Lock

    def __init__(self, port, listen_backlog, number_agencies):
        # Initialize server socket
        self._server_socket = ServerSocket(("", port), listen_backlog)
        self._running = True
        self._first_accept_try = True
        self._number_agencies = number_agencies
        self._connected_clients = set()
        self._agencies_ready = set()
        self._winners_by_agency = {}
        self._lock = threading.Lock()

    def run(self):
        """
        Server that accept a new connections and establishes a
        communication with an agency. After agency with communucation
        finishes, servers starts to accept new connections again.
        """

        while self._running:
            try:
                if self._first_accept_try:
                    logging.info("action: accept_connections | result: in_progress")
                    self._first_accept_try = False

                client_socket: ClientSocket = self._server_socket.accept()

                t = threading.Thread(target=self.__handle_client, args=(client_socket,))
                t.start()

                self._connected_clients.add((t, client_socket))
                self._first_accept_try = True

                self.__close_finished_connections()
            except socket.timeout:
                continue
            except OSError as e:
                if self._running:
                    logging.error(
                        f"action: accept_connections | result: fail | error: {e}"
                    )
                break

        self.__cleanup()

    def __handle_client(self, client_socket: ClientSocket) -> None:
        """
        Handle the communication with an agency

        This function is called in a new thread to handle the communication
        with an agency. The agency is identified by the agency_socket
        """
        try:
            msg: str = self.__wait_for_message(client_socket)
            agency_id = decode_identification_message(msg)

            while self._running:
                msg: str = self.__wait_for_message(client_socket)
                header, payload = decode_message(msg)

                if header == AgencyHeader.BET_BATCH:
                    self.__handle_bet_batch(client_socket, payload)
                elif header == AgencyHeader.FINISH_BETTING:
                    self.__handle_finish_betting(agency_id)
                elif header == AgencyHeader.REQUEST_RESULTS:
                    self.__handle_request_result(client_socket, agency_id)
                elif header == AgencyHeader.SHUTDOWN:
                    return

        except ValueError as _:
            logging.error(
                "action: receive_message | result: fail | error: invalid message format"
            )
        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            client_socket.close()

    def __handle_bet_batch(self, client_socket: ClientSocket, payload: str) -> None:
        try:
            bets = self.__batch_to_bets(payload)

            if len(bets) == 0:
                client_socket.send_message(encode_message(ServerHeader.FAILURE))
                return

            with self._lock:
                store_bets(bets)

            logging.info(
                f"action: apuesta_recibida | result: success | cantidad: {len(bets)}"
            )

            client_socket.send_message(encode_message(ServerHeader.SUCCESS))
        except ValueError as _:
            n_bets: int = len(payload.split(BATCH_SEPARATOR))
            logging.error(
                f"action: apuesta_recibida | result: fail | cantidad: {n_bets}"
            )

            client_socket.send_message(encode_message(ServerHeader.FAILURE))

    def __handle_finish_betting(self, agency_id: int) -> None:
        with self._lock:
            self._agencies_ready.add(agency_id)

    def __handle_request_result(
        self, client_socket: ClientSocket, agency_id: int
    ) -> None:
        with self._lock:
            if agency_id not in self._agencies_ready:
                client_socket.send_message(encode_message(ServerHeader.FAILURE))
                return

            if len(self._agencies_ready) < self._number_agencies:
                client_socket.send_message(encode_message(ServerHeader.NOT_READY))
                return

            if not self._winners_by_agency:
                self.__draw_winners()

            self.__socket.send_message(
                encode_winners_message(self._winners_by_agency[agency_id])
            )

    def __wait_for_message(self, client_socket: ClientSocket) -> str:
        while self._running:
            try:
                return client_socket.receive_message()
            except socket.timeout:
                continue

    def __draw_winners(self):
        """
        Draw the winning number for the lottery. Then store the winning bets for each agency
        """

        self._winners_by_agency = {agency: [] for agency in self._agencies_ready}

        for bet in load_bets():
            if has_won(bet):
                self._winners_by_agency[bet.agency].append(bet.document)

        logging.info("action: sorteo | result: success")

    def __close_finished_connections(self) -> None:
        connected_clients: set[tuple[threading.Thread, ClientSocket]] = set()

        for t, s in self._connected_clients:
            if t.is_alive():
                connected_clients.add((t, s))
            else:
                s.close()
                t.join()

        self._connected_clients = connected_clients

    def __cleanup(self) -> None:
        try:
            for t, s in self._connected_clients:
                s.close()
                t.join()

            self._server_socket.close()
        except Exception as e:
            logging.error(f"action: exit | result: fail | error: {e}")

    def stop(self, _signum, _frame) -> None:
        self._running = False
        self._server_socket.close()
