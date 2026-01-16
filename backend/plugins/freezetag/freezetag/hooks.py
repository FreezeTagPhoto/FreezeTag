from functools import wraps
from typing import Callable
from PIL import Image

_plugin_init = None
_plugin_process = None
_plugin_teardown = None

def init_func(func: Callable[[]]):
    global _plugin_init
    @wraps(func)
    def wrapper():
        return func()
    assert _plugin_init == None, "there can only be one init_func per plugin"
    _plugin_init = wrapper
    return wrapper

def process_func(func: Callable[[Image, int]]):
    global _plugin_process
    @wraps(func)
    def wrapper(img: Image, id: int):
        return func(img, id)
    assert _plugin_process == None, "there can only be one process_func per plugin"
    _plugin_process = wrapper
    return wrapper

def teardown_func(func: Callable[[]]):
    global _plugin_teardown
    @wraps(func)
    def wrapper():
        return func()
    assert _plugin_teardown == None, "there can only be one teardown_func per plugin"
    _plugin_teardown = wrapper
    return wrapper
