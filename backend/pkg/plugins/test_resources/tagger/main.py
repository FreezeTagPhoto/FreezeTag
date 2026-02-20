import freezetag
from freezetag.hooks import single_image, image_batch, AddTagsAction, MultipartAction, Error
from freezetag.message import get_metadata

@single_image
def tag_image(image, id):
    if id == 67:
        return Error("this number sucks")
    if id == 76:
        meta = get_metadata(id)
        return AddTagsAction(id, [meta["fileName"]])
    else:
        return AddTagsAction(id, ["foo", "bar"])

@image_batch
def tag_batch(ids):
    actions = []
    for id in ids:
        actions.append(AddTagsAction(id, ["foo"]))
    return MultipartAction(*actions)

if __name__ == "__main__":
    freezetag.run()