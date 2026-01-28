import freezetag
from freezetag.hooks import process_func, SkipAction
from freezetag.message import log

@process_func
def process(img, id):
    width, height = img.size
    log(f"image is {width}x{height}")
    return SkipAction()

if __name__ == "__main__":
    freezetag.run()