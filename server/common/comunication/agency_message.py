from enum import Enum
from typing import Tuple

from common.utils import Bet

HEARDER_SEPARATOR = ":"
BATCH_SEPARATOR = "*"
BET_SEPARATOR = "+"


class AgencyHeader(Enum):
    IDENTIFICATION = "agency"
    BET_BATCH = "bet_batch"
    FINISH_BETTING = "finish"
    REQUEST_RESULTS = "request_results"
    SHUTDOWN = "shutdown"


def decode_identification_message(msg: str) -> int:
    if msg == "":
        raise ValueError("No identification message received")

    header, payload = decode_message(msg)

    if header != AgencyHeader.IDENTIFICATION:
        raise ValueError("Invalid identification message")

    if not payload.isnumeric():
        raise ValueError("Invalid identification number")

    return int(payload)


def decode_message(msg: str) -> Tuple[AgencyHeader, str]:

    fields: Tuple[str, str] = msg.split(HEARDER_SEPARATOR, 1)

    if len(fields) != 2:
        raise ValueError("Invalid message format")

    header: str = fields[0]
    payload: str = fields[1]

    return AgencyHeader(header), payload


def decode_bet_batch(msg: str) -> list[Bet]:
    """
    Decode a message with multiple bets separated by DOCUMENT_SEPARATOR
    and return a list of Bet objects.
    """
    return [__decode_bet(bet) for bet in msg.split(BATCH_SEPARATOR)]


def __decode_bet(msg: str) -> Bet:
    """
    Decode a message with a single bet and return a Bet object.
    """

    fields: list[str] = msg.split(BET_SEPARATOR)

    if len(fields) != 6:
        raise ValueError("Invalid number of fields")

    if not fields[0].isnumeric():
        raise ValueError("Invalid agency")

    if not fields[3].isnumeric():
        raise ValueError("Invalid document")

    if not fields[5].isnumeric():
        raise ValueError("Invalid number")

    return Bet(*fields)
