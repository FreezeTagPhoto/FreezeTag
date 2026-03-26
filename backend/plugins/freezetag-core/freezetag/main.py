
from .protocol import read_message, write_message, MessageType, Message
from .hooks import NoAction
from .message import log
from . import hooks

from PIL import Image
import io

def run():
    msg = read_message()
    assert msg.mtype == MessageType.READY
    if hooks._plugin_init is not None:
        msg = hooks._plugin_init()
        if msg is not None:
            write_message(msg)
            exit(0)
        else:
            write_message(Message(MessageType.READY))
    else:
        write_message(Message(MessageType.READY))
    
    while True:
        msg = read_message()
        if msg.mtype == MessageType.SHUTDOWN:
            break
        match msg.mtype:
            case MessageType.GET:
                match msg.contents["action"]:
                    case "single_image":
                        hook = msg.contents["hook"]
                        id = msg.contents["id"]
                        msg = read_message()
                        if msg.mtype != MessageType.BIN:
                            write_message(Message(MessageType.ERR, b'no context provided with process-image'))
                            continue
                        img = Image.open(io.BytesIO(msg.contents))
                        hook_func = hooks._plugin_hooks.get(hook, None)
                        if hook_func is None:
                            write_message(Message(MessageType.ERR, f'single_image hook "{hook}" not found'.encode("utf-8")))
                            continue
                        try:
                            action = hook_func(img, id)
                        except Exception as error:
                            write_message(Message(MessageType.ERR, f'exception during hook: {error}'))
                            continue
                        if action is not None:
                            write_message(action)
                        else:
                            write_message(NoAction())
                    case "image_batch":
                        hook = msg.contents["hook"]
                        ids = msg.contents["ids"]
                        hook_func = hooks._plugin_hooks.get(hook, None)
                        if hook_func is None:
                            write_message(Message(MessageType.ERR, f'image_batch hook "{hook}" not found'))
                            continue
                        action = hooks.NoAction()
                        try:
                            action = hook_func(ids)
                        except Exception as error:
                            write_message(Message(MessageType.ERR, f'exception during hook: {error}'))
                            continue
                        if action is not None:
                            write_message(action)
                        else:
                            write_message(NoAction())
                    case "form_data":
                        # TODO: this is for taking in form data from the server
                        write_message(Message(MessageType.ERR, f'unimplemented form_data hook'))
                        continue

                        # postlude
                        try:
                            action = hook_func(ids)
                        except Exception as error:
                            write_message(Message(MessageType.ERR, f'exception during hook: {error}'))
                            continue
                        if action is not None:
                            write_message(action)
                        else:
                            write_message(NoAction())
                    case _:
                        write_message(Message(MessageType.ERR, f'unsupported hook signature'))
            case _:
                write_message(Message(MessageType.ERR, f'{msg.mtype.name} not supported.'.encode("utf-8")))
    
    if hooks._plugin_teardown is not None:
        msg = hooks._plugin_teardown()
        if msg is not None:
            write_message(msg)
            exit(0)
        else:
            write_message(Message(MessageType.SHUTDOWN))
    else:
        write_message(Message(MessageType.SHUTDOWN))
    exit(0)