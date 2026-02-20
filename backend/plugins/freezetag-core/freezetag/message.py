
from PIL import Image
import datetime, io
from .protocol import read_message, write_message, Message, MessageType

def log(msg: str):
    write_message(Message(MessageType.LOG, msg.encode("utf-8")))

def get_metadata(id: int) -> dict[str, object]:
    write_message(Message(MessageType.GET, {"action": "metadata", "id": id}))
    response = read_message()
    if isinstance(response.contents, dict) and response.contents["action"] == "metadata":
        return response.contents["data"]
    return {}

def get_image(id: int) -> Image.Image:
    write_message(Message(MessageType.GET, {"action": "image", "id": id}))
    response = read_message()
    if isinstance(response.contents, dict) and response.contents["action"] == "image":
        response = read_message()
        img = Image.open(io.BytesIO(response.contents))
        return img
    return None

def get_image_tags(id: int) -> list[str]:
    write_message(Message(MessageType.GET, {"action": "tags", "id": id}))
    response = read_message()
    if isinstance(response.contents, dict) and response.contents["action"] == "tags":
        return response.contents["tags"]
    return []

def get_all_tags() -> dict[str, int]:
    write_message(Message(MessageType.GET, {"action": "all_tags"}))
    response = read_message()
    if isinstance(response.contents, dict) and response.contents["action"] == "all_tags":
        return response.contents["tags"]
    return {}

def search_images(
    make: str = "",
    model: str = "",
    makeLike: str = "",
    modelLike: str = "",
    near: (float, float, float) = None,
    tags: list[str] = [],
    tagsLike: list[str] = [],
    takenBefore: datetime.datetime = None,
    takenAfter: datetime.datetime = None, 
    uploadedBefore: datetime.datetime = None,
    uploadedAfter: datetime.datetime = None
    ) -> list[int]:
    tbs = ""
    if takenBefore is not None:
        tbs = str(int(takenBefore.timestamp()))
    tas = ""
    if takenAfter is not None:
        tas = str(int(takenAfter.timestamp()))
    ubs = ""
    if uploadedBefore is not None:
        ubs = str(int(uploadedBefore.timestamp()))
    uas = ""
    if uploadedAfter is not None:
        uas = str(int(uploadedAfter.timestamp()))
    nears = ""
    if near is not None:
        nears = f"{near[0]},{near[1]},{near[2]}"
    write_message(Message(MessageType.GET, {"action": "search", "query": {
        "make": make,
        "model": model,
        "makeLike": makeLike,
        "modelLike": modelLike,
        "takenBefore": tbs,
        "takenAfter": tas,
        "uploadedBefore": ubs,
        "uploadedAfter": uas,
        "near": nears,
        "tags": tags,
        "tagsLike": tagsLike
    }}))
    response = read_message()
    if isinstance(response.contents, dict) and response.contents["action"] == "search":
        return response.contents["results"]
    return []

def query_tags(
    make: str = "",
    model: str = "",
    makeLike: str = "",
    modelLike: str = "",
    near: (float, float, float) = None,
    tags: list[str] = [],
    tagsLike: list[str] = [],
    takenBefore: datetime.datetime = None,
    takenAfter: datetime.datetime = None, 
    uploadedBefore: datetime.datetime = None,
    uploadedAfter: datetime.datetime = None
    ) -> dict[str, int]:
    tbs = ""
    if takenBefore is not None:
        tbs = str(int(takenBefore.timestamp()))
    tas = ""
    if takenAfter is not None:
        tas = str(int(takenAfter.timestamp()))
    ubs = ""
    if uploadedBefore is not None:
        ubs = str(int(uploadedBefore.timestamp()))
    uas = ""
    if uploadedAfter is not None:
        uas = str(int(uploadedAfter.timestamp()))
    nears = ""
    if near is not None:
        nears = f"{near[0]},{near[1]},{near[2]}"
    write_message(Message(MessageType.GET, {"action": "tags_query", "query": {
        "make": make,
        "model": model,
        "makeLike": makeLike,
        "modelLike": modelLike,
        "takenBefore": tbs,
        "takenAfter": tas,
        "uploadedBefore": ubs,
        "uploadedAfter": uas,
        "near": nears,
        "tags": tags,
        "tagsLike": tagsLike
    }}))
    response = read_message()
    if isinstance(response.contents, dict) and response.contents["action"] == "tags_query":
        return response.contents["tags"]
    return {}