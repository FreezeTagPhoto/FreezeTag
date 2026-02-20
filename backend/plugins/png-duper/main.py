import freezetag
from freezetag.hooks import single_image, AddImageAction

@single_image
def dupe_image(img, id):
    return AddImageAction("test", "png", img)

if __name__ == "__main__":
    freezetag.run()