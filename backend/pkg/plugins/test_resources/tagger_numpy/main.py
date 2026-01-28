import freezetag
from freezetag.hooks import process_func, TagAction

import numpy as np

@process_func
def tag_image(image, id):
    tags = []
    for i in np.arange(2, 7):
        tags.append(str(i))
    return TagAction(id, tags)

if __name__ == "__main__":
    freezetag.run()