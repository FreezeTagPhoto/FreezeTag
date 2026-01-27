
from .protocol import read_message, write_message, Message, MessageType

def log(msg: str):
    write_message(Message(MessageType.LOG, msg.encode("utf-8")))

def get_metadata(id: int) -> dict[str, object]:
    write_message(Message(MessageType.GET, {"action": "metadata", "id": id}))
    response = read_message()
    if isinstance(response.contents, dict) and response.contents["action"] == "metadata":
        return response.contents["data"]
    return {}