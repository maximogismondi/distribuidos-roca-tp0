from enum import Enum

HEARDER_SEPARATOR = ":"
DOCUMENT_SEPARATOR = ","


class ServerHeader(Enum):
    WINNERS = "winners"
    NOT_READY = "not_ready"
    SUCCESS = "success"
    FAILURE = "failure"


def encode_winners_message(winners: list[str]) -> str:
    return encode_message(ServerHeader.WINNERS, DOCUMENT_SEPARATOR.join(winners))


def encode_message(header: ServerHeader, payload: str = "") -> str:
    return f"{header.value}{HEARDER_SEPARATOR}{payload}"
