import freezetag
from freezetag.hooks import single_image, AddTagsAction

import numpy as np

@single_image
def tag_image(image, id):
    tags = []
    for i in np.arange(2, 7):
        tags.append(str(i))
    return AddTagsAction(id, tags)

if __name__ == "__main__":
    freezetag.run()