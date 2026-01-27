import freezetag
from freezetag.hooks import process_func, TagAction

@process_func
def process(img, id):
    return TagAction(id, ["foo", "bar"])

if __name__ == "__main__":
    freezetag.run()