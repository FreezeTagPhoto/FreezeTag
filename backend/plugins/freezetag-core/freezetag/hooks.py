from functools import wraps
from typing import Callable
from PIL import Image

from .message import Message, MessageType

_plugin_hooks: dict[string, Callable] = {}
_plugin_init = None
_plugin_teardown = None

class HookAction(Message):
    mtype = MessageType.PUT
    def __init__(self, info: dict[string, object], action="skip"):
        self.contents = info | {"action" : action}

class SkipAction(HookAction):
    def __init__(self):
        HookAction.__init__(self, {})

class TagAction(HookAction):
    def __init__(self, id: int, tags: list[str]):
        HookAction.__init__(self, {"tags": tags, "id": id}, "tag")

def init_func(func: Callable[[]]):
    global _plugin_init
    @wraps(func)
    def wrapper():
        return func()
    assert _plugin_init == None, "there can only be one init_func per plugin"
    _plugin_init = wrapper
    return wrapper

def process_func(func: Callable[[Image, int], HookAction]):
    global _plugin_hooks
    @wraps(func)
    def wrapper(img: Image, id: int) -> HookAction:
        return func(img, id)
    _plugin_hooks[func.__name__] = func
    return wrapper

def teardown_func(func: Callable[[]]):
    global _plugin_teardown
    @wraps(func)
    def wrapper():
        return func()
    assert _plugin_teardown == None, "there can only be one teardown_func per plugin"
    _plugin_teardown = wrapper
    return wrapper
