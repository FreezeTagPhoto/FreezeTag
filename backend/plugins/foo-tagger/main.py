import freezetag
from freezetag.hooks import process_func, TagAction
from freezetag.message import log

@process_func
def tag_image(img, id):
    log(f"tagging image ID {id} with 'foo'")
    return TagAction(id, ["foo"])

if __name__ == "__main__":
    freezetag.run()