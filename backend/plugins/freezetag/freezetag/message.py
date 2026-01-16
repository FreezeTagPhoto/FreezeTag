
from .protocol import read_message, write_message, Message, MessageType

def log(msg: str):
    write_message(Message(MessageType.LOG, msg.encode("utf-8")))

def skip():
    write_message(Message(MessageType.PUT, {"action": "skip",}))

def add_tags(id: int, tags: list[str]):
    write_message(Message(MessageType.PUT, {"action": "tag", "id": id, "tags": tags}))