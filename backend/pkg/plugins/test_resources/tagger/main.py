import freezetag
from freezetag.hooks import process_func, TagAction, Error
from freezetag.message import get_metadata

@process_func
def tag_image(image, id):
    if id == 67:
        return Error("this number sucks")
    if id == 76:
        meta = get_metadata(id)
        return TagAction(id, [meta["fileName"]])
    else:
        return TagAction(id, ["foo", "bar"])

if __name__ == "__main__":
    freezetag.run()