
from .protocol import read_message, write_message, Message, MessageType

def log(msg: str):
    write_message(Message(MessageType.LOG, msg.encode("utf-8")))
