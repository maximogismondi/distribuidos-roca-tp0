import socket
import logging
import threading

from common.utils import store_bets, from_string, load_bets, has_won
from common.agency_socket import AgencySocket

BATCH_SEPARATOR = "*"
DOCUMENT_SEPARATOR = ","

SUCCESS_MSG = "success"
FAILURE_MSG = "failure"

FINISH_MSG = "finish"
REQUEST_RESULT_MESSAGE = "request"
WINNERS_MESSAGE = "winners"
NOT_READY_MESSAGE = "not_ready"


class Server:
    def __init__(self, port, listen_backlog, number_agencies):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(("", port))
        self._server_socket.listen(listen_backlog)
        self._server_socket.settimeout(1.0)
        self._running = True
        self._first_accept_try = True
        self._number_agencies = number_agencies
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
                agency_socket = self.__accept_new_connection()

                if self._running and agency_socket:
                    t = threading.Thread(
                        target=self.__handle_agency, args=(agency_socket,)
                    )
                    t.start()

            except socket.timeout:
                continue
            except OSError as e:
                if self._running:
                    logging.error(
                        f"action: accept_connections | result: fail | error: {e}"
                    )
                break

    def __draw_winners(self):
        """
        Draw the winning number for the lottery. Then store the winning bets for each agency
        """

        self._winners_by_agency = {agency: [] for agency in self._agencies_ready}

        for bet in load_bets():
            if has_won(bet):
                self._winners_by_agency[bet.agency].append(bet.document)

        logging.info("action: sorteo | result: success")

    def __send_winners_results(self, agency_socket):
        """
        Send winners to the agencies
        """

        resultsStr = DOCUMENT_SEPARATOR.join(
            [WINNERS_MESSAGE] + self._winners_by_agency[agency_socket.agency_id()]
        )
        agency_socket.send_message(resultsStr)

    def __batch_to_bets(self, batch):
        betsStr = batch.split(BATCH_SEPARATOR)
        bets = [None] * len(betsStr)

        for i, betStr in enumerate(betsStr):
            try:
                bets[i] = from_string(betStr)
            except ValueError as _e:
                logging.error(
                    f"action: apuesta_recibida | result: fail | cantidad: {len(bets)}"
                )
                return []
            except Exception as _e:
                logging.error(
                    f"action: apuesta_recibida | result: fail | cantidad: {len(bets)}"
                )
                return []

        return bets

    def __handle_ready_agency(self, agency_socket):
        """
        Handle the communication with an agency that has finished sending bets
        """
        msg = agency_socket.receive_message()

        if msg != REQUEST_RESULT_MESSAGE:
            agency_socket.send_message(FAILURE_MSG)
            return

        if len(self._agencies_ready) == self._number_agencies:
            if not self._winners_by_agency:
                self.__draw_winners()

            self.__send_winners_results(agency_socket)
        else:
            agency_socket.send_message(NOT_READY_MESSAGE)

    def __handle_new_agency(self, agency_socket):
        """
        Read message from a specific agency and store the bets

        If a problem arises in the communication with the agency, the
        agency socket will also be closed
        """

        # Add this flag to avoid logging every accept try
        self._first_accept_try = True

        try:
            while True:
                msg = agency_socket.receive_message()

                if msg == FINISH_MSG:
                    self._agencies_ready.add(agency_socket.agency_id())
                    self.__handle_ready_agency(agency_socket)
                    return

                bets = self.__batch_to_bets(msg)

                if len(bets) == 0:
                    agency_socket.send_message(FAILURE_MSG)
                    return

                with self._lock:
                    store_bets(bets)

                logging.info(
                    f"action: apuesta_recibida | result: success | cantidad: {len(bets)}"
                )

                agency_socket.send_message(SUCCESS_MSG)
        except ConnectionError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")

    def __handle_agency(self, agency_socket):
        """
        Handle the communication with an agency

        This function is called in a new thread to handle the communication
        with an agency. The agency is identified by the agency_socket
        """

        try:
            if agency_socket.agency_id() in self._agencies_ready:
                self.__handle_ready_agency(agency_socket)
            else:
                self.__handle_new_agency(agency_socket)
        finally:
            if agency_socket:
                agency_socket.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to an agency is made.
        Then connection created is printed and returned
        """

        # Add this flag to avoid logging every accept try
        if self._first_accept_try:
            logging.info("action: accept_connections | result: in_progress")
            self._first_accept_try = False

        try:
            c, addr = self._server_socket.accept()
            agency_socket = AgencySocket.new_socket(c)

            logging.info(
                f"action: accept_connections | result: success | ip: {addr[0]}"
            )

            if not agency_socket:
                return None

            return agency_socket
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
        except Exception as e:
            logging.error(f"action: exit | result: fail | error: {e}")
