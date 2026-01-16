import freezetag
from freezetag.message import add_tags

@freezetag.hooks.process_func
def process(img, id):
    add_tags(id, ["foo", "bar"])

if __name__ == "__main__":
    freezetag.run()