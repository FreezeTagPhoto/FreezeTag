import freezetag
from freezetag.hooks import single_image, form_data, SendFormAction, AddImageAction
from freezetag.message import get_image

from PIL import ImageDraw, ImageFont


@single_image
def make_caption_form(_img, id):
    return SendFormAction(
        f"""<form><input type="hidden" value="{id}" name="id"></input><label for="top_text">Top Text:</label><input type="text" name="top_text" id="top_text" required></input><label for="bottom_text">Bottom Text:</label><input type="text" name="bottom_text" id="bottom_text" required></input></form>"""
    )


@form_data
def process_caption_form(data):
    id = data["id"]
    top = data["top_text"]
    bottom = data["bottom_text"]

    font_size = 64

    image = get_image(int(id))
    draw = ImageDraw.Draw(image)
    font = ImageFont.truetype("impact.ttf", font_size)

    draw.text((0, 0), top, (255, 255, 255), font=font)

    width, height = image.size
    draw.text((0, height - font_size), bottom, (255, 255, 255), font=font)

    return AddImageAction(f"{id}_meme", "png", image)


if __name__ == "__main__":
    freezetag.run()
