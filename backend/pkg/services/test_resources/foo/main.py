import freezetag
from freezetag.hooks import process_func, init_func, TagAction
from freezetag.message import log

# this file shows off the way functions are called in the same context
# and also sequentially (although in no guaranteed order)

index_A = 0
index_B = 3

@init_func
def init():
    global index_A
    index_A = 1

@process_func
def tag_image(img, id):
    global index_A
    log(f"tagging image ID {id} on function 1")
    action = TagAction(id, [str(index_A)])
    index_A = index_A + 1
    return action

@process_func
def tag_image_2(img, id):
    global index_B
    log(f"tagging image ID {id} on function 2")
    action = TagAction(id, [str(index_B)])
    index_B = index_B + 3
    return action

if __name__ == "__main__":
    freezetag.run()