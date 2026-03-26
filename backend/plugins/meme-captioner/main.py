import freezetag
from freezetag.hooks import single_image, AddImageAction

@single_image
def make_caption_form(img, id):
    # TODO: This function should return a form through a like "MakeFormAction" or something
    return None

@form_data
def process_caption_form(data):
    # TODO: This function can actually choose to modify images and such based on the form data
    # TODO: The form data should probably just be akin to a dictionary?
    return None

if __name__ == "__main__":
    freezetag.run()