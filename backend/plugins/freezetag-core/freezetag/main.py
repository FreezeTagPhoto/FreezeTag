
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
                if msg.contents["action"] == "process-image":
                    hook = msg.contents["hook"]
                    id = msg.contents["id"]
                    msg = read_message()
                    if msg.mtype != MessageType.BIN:
                        write_message(Message(MessageType.ERR, b'no context provided with process-image'))
                        continue
                    img = Image.open(io.BytesIO(msg.contents))
                    hook_func = hooks._plugin_hooks.get(hook, None)
                    if hook_func is None:
                        write_message(Message(MessageType.ERR, f'process-image hook "{hook}" not found'.encode("utf-8")))
                        continue
                    action = hook_func(img, id)
                    write_message(action)
                else:
                    write_message(Message(MessageType.ERR, b'non-process-image requests not supported yet.'))
            case _:
                write_message(Message(MessageType.ERR, f'{msg.mtype.name} not supported yet'.encode("utf-8")))
    
    if hooks._plugin_teardown is not None:
        hooks._plugin_teardown()
    write_message(Message(MessageType.SHUTDOWN))
    exit(0)