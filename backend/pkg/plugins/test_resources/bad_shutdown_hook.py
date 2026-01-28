import freezetag
from freezetag.hooks import teardown_func, Error
from freezetag.message import log

@teardown_func
def bad_teardown():
    log("this teardown hook does not work")
    return Error("told you")

if __name__ == "__main__":
    freezetag.run()