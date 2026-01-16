
from .protocol import read_message, write_message, MessageType, Message
from . import hooks

from PIL import Image
import io

def run():
    msg = read_message()
    assert msg.mtype == MessageType.READY
    if hooks._plugin_init is not None:
        hooks._plugin_init()
    write_message(Message(MessageType.READY))
    while True:
        msg = read_message()
        if msg.mtype == MessageType.SHUTDOWN:
            break
        match msg.mtype:
            case MessageType.GET:
                if msg.contents["action"] == "process":
                    id = msg.contents["id"]
                    msg = read_message()
                    if msg.mtype != MessageType.BIN:
                        write_message(Message(MessageType.ERR, b'non-bin image processes not supported yet'))
                        continue
                    img = Image.open(io.BytesIO(msg.contents))
                    if hooks._plugin_process is not None:
                        hooks._plugin_process(img, id)
                else:
                    write_message(Message(MessageType.ERR, b'non-process requests not supported yet.'))
            case _:
                write_message(Message(MessageType.ERR, f'{msg.mtype.name} not supported yet'.encode("utf-8")))
    
    if hooks._plugin_teardown is not None:
        hooks._plugin_teardown()
    write_message(Message(MessageType.SHUTDOWN))
    exit(0)