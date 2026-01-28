import freezetag
from freezetag.hooks import init_func, Error
from freezetag.message import log

@init_func
def bad_init():
    log("this init hook does not work")
    return Error("told you")

if __name__ == "__main__":
    freezetag.run()