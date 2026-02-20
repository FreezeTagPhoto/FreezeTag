import freezetag
from freezetag.hooks import single_image, image_batch, AddTagsAction, MultipartAction
from freezetag.message import log

@single_image
def tag_image(img, id):
    log(f"tagging image ID {id} on function 1")
    action = AddTagsAction(id, ["bar"])
    return action

@image_batch
def tag_all(ids):
    global index_B
    log(f"tagging image IDs {ids} on function 2")
    actions = []
    for id in ids:
        actions.append(AddTagsAction(id, ["foo"]))
    return MultipartAction(*actions)

if __name__ == "__main__":
    freezetag.run()