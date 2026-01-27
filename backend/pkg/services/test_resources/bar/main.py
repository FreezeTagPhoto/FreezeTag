import freezetag
from freezetag.hooks import process_func, init_func, TagAction
from freezetag.message import log

# this file also serves to show off the retention of context between calls in the same job

index = 0

@init_func
def init():
    global index
    index = 1

@process_func
def tag_image(img, id):
    global index
    log(f"tagging image ID {id}")
    action = TagAction(id, [str(index)])
    index = index + 1
    return action

if __name__ == "__main__":
    freezetag.run()