import freezetag
from freezetag.message import skip, log

@freezetag.hooks.process_func
def process(img, id):
    width, height = img.size
    log(f"image is {width}x{height}")
    skip()

if __name__ == "__main__":
    freezetag.run()