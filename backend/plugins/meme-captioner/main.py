import freezetag
from freezetag.hooks import single_image, form_data, SendFormAction, NoAction
from freezetag.message import log

@single_image
def make_caption_form(img, id):
    # TODO: This function should return a form through a like "MakeFormAction" or something
    return SendFormAction("<form>yo</form>")

@form_data
def process_caption_form(data):
    log("Running process caption form...")
    log(f"Expected to see name field as Dan, got {data["name"]}")
    return NoAction()

if __name__ == "__main__":
    freezetag.run()