from functools import wraps
from typing import Callable, Any
from PIL import Image
import io, base64

from .message import Message, MessageType

_plugin_hooks: dict[str, Callable] = {}
_plugin_init = None
_plugin_teardown = None

class HookAction(Message):
    mtype = MessageType.PUT
    def __init__(self, info: dict[str, object], action="none"):
        self.contents = info | {"action" : action}

class Error(Message):
    mtype = MessageType.ERR
    def __init__(self, msg: str):
        self.contents = msg.encode("utf-8")

class NoAction(HookAction):
    def __init__(self):
        HookAction.__init__(self, {})

class AddTagsAction(HookAction):
    def __init__(self, id: int, tags: list[str]):
        HookAction.__init__(self, {"tags": tags, "id": id}, "add_tags")

class RemoveTagsAction(HookAction):
    def __init__(self, id: int, tags: list[str]):
        HookAction.__init__(self, {"tags": tags, "id": id}, "remove_tags")

class DeleteTagsAction(HookAction):
    def __init__(self, tags: list[str]):
        HookAction.__init__(self, {"tags": tags}, "delete_tags")

class DeleteImageAction(HookAction):
    def __init__(self, id: int):
        HookAction.__init__(self, {"id": id}, "delete_image")

class AddImageAction(HookAction):
    def __init__(self, name: str, format: str, image: Image.Image):
        byte_arr = io.BytesIO()
        image.save(byte_arr, format=format)
        image_bytes = byte_arr.getvalue()
        data = base64.b64encode(image_bytes)
        HookAction.__init__(self, {"name": name + "." + format, "data": data.decode("utf-8")}, "add_image")

class SendFormAction(HookAction):
    def __init__(self, form: str):
        HookAction.__init__(self, {"form": form}, "send_form")
        
class MultipartAction(HookAction):
    def __init__(self, *hooks: HookAction):
        actions = []
        for action in hooks:
            actions.append(action.contents)
        HookAction.__init__(self, {"parts": actions}, "multipart")


def init_func(func: Callable[[], None | Message]):
    global _plugin_init
    @wraps(func)
    def wrapper():
        return func()
    assert _plugin_init == None, "there can only be one init_func per plugin"
    _plugin_init = wrapper
    return wrapper

def single_image(func: Callable[[Image, int], HookAction]):
    global _plugin_hooks
    @wraps(func)
    def wrapper(img: Image, id: int) -> HookAction:
        return func(img, id)
    _plugin_hooks[func.__name__] = func
    return wrapper

def image_batch(func: Callable[[list[int]], HookAction]):
    global _plugin_hooks
    @wraps(func)
    def wrapper(ids: list[int]) -> HookAction:
        return func(ids)
    _plugin_hooks[func.__name__] = func
    return wrapper

_form_data_exists = False
def form_data(func: Callable[[dict[str, Any]], HookAction]):
    global _plugin_hooks, _form_data_exists
    if _form_data_exists:
        raise Exception("Only support one form data signature hook per plugin (you can filter the forms yourself :))")
    @wraps(func)
    def wrapper(form: dict[str, Any]) -> HookAction:
        return func(form)
    _plugin_hooks[func.__name__] = func
    _form_data_exists = True
    return wrapper

def teardown_func(func: Callable[[], None | Message]):
    global _plugin_teardown
    @wraps(func)
    def wrapper():
        return func()
    assert _plugin_teardown == None, "there can only be one teardown_func per plugin"
    _plugin_teardown = wrapper
    return wrapper
