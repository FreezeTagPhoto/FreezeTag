import json, sys, os

from enum import Enum
from dataclasses import dataclass
from typing import Any

_stdin = sys.stdin
_stdout = sys.stdout

_null_fd = os.open(os.devnull, os.O_RDWR)

sys.stdin = os.fdopen(_null_fd, 'r')
sys.stdout = sys.stderr

class MessageType(Enum):
    GET = 0
    PUT = 1
    BIN = 2
    LOG = 3
    ERR = 4
    READY = 5
    SHUTDOWN = 6

@dataclass
class Message:
    mtype: MessageType
    contents: None | bytes | dict = None

def read_message() -> Message:
    msg_byte: bytes = _stdin.buffer.read(1)
    if _stdin.closed or len(msg_byte) == 0:
        raise EOFError("stdin closed before protocol exit")
    msg_type = MessageType(msg_byte[0])
    match msg_type:
        case MessageType.READY | MessageType.SHUTDOWN:
            return Message(msg_type)
        case MessageType.BIN | MessageType.LOG | MessageType.ERR:
            msg_size_buf: bytes = _stdin.buffer.read(8)
            msg_size: int = int.from_bytes(msg_size_buf, byteorder='little', signed=False)
            msg_contents: bytes = _stdin.buffer.read(msg_size)
            return Message(msg_type, msg_contents)
        case _:
            msg_size_buf: bytes = _stdin.buffer.read(8)
            msg_size: int = int.from_bytes(msg_size_buf, byteorder='little', signed=False)
            msg_bytes: bytes = _stdin.buffer.read(msg_size)
            msg_contents = json.loads(msg_bytes)
            return Message(msg_type, msg_contents)

def write_message(msg: Message):
    if _stdout.closed:
        raise EOFError("stdout closed before protocol exit")
    _stdout.buffer.write(bytes([msg.mtype.value]))
    match msg.mtype:
        case MessageType.READY | MessageType.SHUTDOWN:
            pass # these have nothing to do
        case MessageType.BIN | MessageType.LOG | MessageType.ERR:
            if isinstance(msg.contents, bytes):
                msg_size = len(msg.contents)
                _stdout.buffer.write(int.to_bytes(msg_size, 8, byteorder='little', signed=False))
                _stdout.buffer.write(msg.contents)
            else:
                raise ValueError("non-bytes BIN/LOG/ERR not allowed")
        case _:
            msg_contents = json.dumps(msg.contents).encode("utf-8")
            msg_size = len(msg_contents)
            _stdout.buffer.write(int.to_bytes(msg_size, 8, byteorder='little', signed=False))
            _stdout.buffer.write(msg_contents)
    _stdout.buffer.flush()
