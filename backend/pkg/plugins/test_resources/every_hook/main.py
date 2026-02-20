import freezetag
from freezetag.hooks import *
from freezetag.message import *

@single_image
def every(image, id) -> HookAction:
    match id:
        case 1:
            return AddTagsAction(1, ["foo"])
        case 2:
            return RemoveTagsAction(1, ["foo"])
        case 3:
            return DeleteTagsAction(["foo"])
        case 4:
            return DeleteImageAction(1)
        case 5:
            return AddImageAction("foo", "webp", Image.open("gopher.webp"))
        case 6:
            res = get_image(1)
            if res is None:
                return Error("couldn't get image")
            return NoAction()
        case 7:
            res = get_image_tags(1)
            if res is None:
                return Error("couldn't get image tags")
            return NoAction()
        case 8:
            res = get_all_tags()
            if res is None:
                return Error("couldn't get all tags")
            return NoAction()
        case 9:
            res = search_images(near=(6.9, 42.0, 6.7), tags=["foo", "bar"])
            if res is None:
                return Error("couldn't search")
            return NoAction()
        case 10:
            res = query_tags(tagsLike=["foo"])
            if res is None:
                return Error("couldn't query tags")
            return NoAction()
    return NoAction()

if __name__ == "__main__":
    freezetag.run()